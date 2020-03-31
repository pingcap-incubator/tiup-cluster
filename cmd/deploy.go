// Copyright 2020 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pingcap-incubator/tiops/pkg/meta"
	"github.com/pingcap-incubator/tiops/pkg/task"
	"github.com/pingcap-incubator/tiops/pkg/utils"
	"github.com/pingcap-incubator/tiup/pkg/repository"
	"github.com/pingcap-incubator/tiup/pkg/set"
	tiuputils "github.com/pingcap-incubator/tiup/pkg/utils"
	"github.com/pingcap/errors"
	"github.com/spf13/cobra"
)

type componentInfo struct {
	component string
	version   repository.Version
}

type deployOptions struct {
	user       string // username to login to the SSH server
	usePasswd  bool   // use password for authentication
	password   string // password of the user
	keyFile    string // path to the private key file
	passphrase string // passphrase of the private key file
}

func newDeploy() *cobra.Command {
	opt := deployOptions{}
	cmd := &cobra.Command{
		Use:          "deploy <cluster-name> <version> <topology.yaml>",
		Short:        "Deploy a cluster for production",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 3 {
				return cmd.Help()
			}
			if opt.usePasswd {
				opt.password = utils.GetPasswd("Password:")
			}
			if len(opt.keyFile) == 0 && len(opt.password) == 0 {
				return errPasswordKeyAtLeastOne
			}

			auditConfig.enable = true
			return deploy(args[0], args[1], args[2], opt)
		},
	}

	cmd.Flags().StringVar(&opt.user, "user", "root", "Specify the system user name")
	cmd.Flags().BoolVar(&opt.usePasswd, "password", false, "Specify the password of system user")
	cmd.Flags().StringVar(&opt.keyFile, "key", "", "Specify the key path of system user")
	cmd.Flags().StringVar(&opt.passphrase, "passphrase", "", "Specify the passphrase of the key")

	return cmd
}

// getComponentVersion maps the TiDB version to the third components binding version
func getComponentVersion(comp, version string) repository.Version {
	switch comp {
	case meta.ComponentPrometheus:
		return "v2.16.0"
	case meta.ComponentGrafana:
		return "v6.1.6"
	case meta.ComponentAlertManager:
		return "v0.20.0"
	case meta.ComponentBlackboxExporter:
		return "v0.16.0"
	case meta.ComponentNodeExporter:
		return "v0.18.1"
	case meta.ComponentPushwaygate:
		return "v1.2.0"
	default:
		return repository.Version(version)
	}
}

func deploy(clusterName, version, topoFile string, opt deployOptions) error {
	if tiuputils.IsExist(meta.ClusterPath(clusterName, meta.MetaFileName)) {
		return errors.Errorf("cluster name '%s' exists, please choose another cluster name", clusterName)
	}

	var topo meta.TopologySpecification
	if err := utils.ParseYaml(topoFile, &topo); err != nil {
		return err
	}
	topo.ClusterName = clusterName
	if err := os.MkdirAll(meta.ClusterPath(clusterName), 0755); err != nil {
		return err
	}

	var (
		envInitTasks      []task.Task // tasks which are used to initialize environment
		downloadCompTasks []task.Task // tasks which are used to download components
		deployCompTasks   []task.Task // tasks which are used to copy components to remote host
	)

	// Initialize environment
	uniqueHosts := set.NewStringSet()
	topo.IterInstance(func(inst meta.Instance) {
		if !uniqueHosts.Exist(inst.GetHost()) {
			uniqueHosts.Insert(inst.GetHost())
			t := task.NewBuilder().
				RootSSH(inst.GetHost(), inst.GetSSHPort(), opt.user, opt.password, opt.keyFile, opt.passphrase).
				EnvInit(inst.GetHost(), topo.GlobalOptions.User).
				UserSSH(inst.GetHost(), topo.GlobalOptions.User).
				Build()
			envInitTasks = append(envInitTasks, t)
		}
	})

	// Download missing component
	downloadCompTasks = buildDownloadCompTasks(version, &topo)

	// Deploy components to remote
	topo.IterInstance(func(inst meta.Instance) {
		//version := getComponentVersion(inst.ComponentName(), version)
		deployDir := inst.DeployDir()
		if !strings.HasPrefix(deployDir, "/") {
			deployDir = filepath.Join("/home/", topo.GlobalOptions.User, deployDir)
		}
		// data dir would be empty for components which don't need it
		dataDir := inst.DataDir()
		if dataDir != "" && !strings.HasPrefix(dataDir, "/") {
			dataDir = filepath.Join("/home/", topo.GlobalOptions.User, dataDir)
		}
		// log dir will always be with values, but might not used by the component
		logDir := inst.LogDir()
		if !strings.HasPrefix(logDir, "/") {
			logDir = filepath.Join("/home/", topo.GlobalOptions.User, logDir)
		}
		// Deploy component
		t := task.NewBuilder().
			Mkdir(inst.GetHost(),
				filepath.Join(deployDir, "bin"),
				filepath.Join(deployDir, "conf"),
				filepath.Join(deployDir, "scripts"),
				dataDir,
				logDir).
			CopyComponent(inst.ComponentName(), repository.Version(version), inst.GetHost(), deployDir).
			InitConfig(
				clusterName,
				inst,
				topo.GlobalOptions.User,
				meta.DirPaths{
					Deploy: deployDir,
					Data:   dataDir,
					Log:    logDir,
				},
			).
			Build()
		deployCompTasks = append(deployCompTasks, t)
	})

	// Deploy monitor relevant components to remote
	dlTasks, dpTasks := buildMonitoredDeployTask(clusterName, uniqueHosts, topo.GlobalOptions, topo.MonitoredOptions, version)
	downloadCompTasks = append(downloadCompTasks, dlTasks...)
	deployCompTasks = append(deployCompTasks, dpTasks...)

	t := task.NewBuilder().
		SSHKeyGen(meta.ClusterPath(clusterName, "ssh", "id_rsa")).
		Parallel(downloadCompTasks...).
		Parallel(envInitTasks...).
		Parallel(deployCompTasks...).
		Build()

	if err := t.Execute(task.NewContext()); err != nil {
		return errors.Trace(err)
	}

	return meta.SaveClusterMeta(clusterName, &meta.ClusterMeta{
		User:     topo.GlobalOptions.User,
		Version:  version,
		Topology: &topo,
	})
}

func buildDownloadCompTasks(version string, topo *meta.Specification) []task.Task {
	var tasks []task.Task
	topo.IterComponent(func(comp meta.Component) {
		if len(comp.Instances()) < 1 {
			return
		}
		version := getComponentVersion(comp.Name(), version)
		t := task.NewBuilder().Download(comp.Name(), version).Build()
		tasks = append(tasks, t)
	})
	return tasks
}

func buildMonitoredDeployTask(
	clusterName string,
	uniqueHosts set.StringSet,
	globalOptions meta.GlobalOptions,
	monitoredOptions meta.MonitoredOptions,
	version string) (downloadCompTasks, deployCompTasks []task.Task) {
	for _, comp := range []string{meta.ComponentNodeExporter, meta.ComponentBlackboxExporter} {
		version := getComponentVersion(comp, version)
		t := task.NewBuilder().
			Download(comp, version).
			Build()
		downloadCompTasks = append(downloadCompTasks, t)

		for host := range uniqueHosts {
			deployDir := monitoredOptions.DeployDir
			if !strings.HasPrefix(deployDir, "/") {
				deployDir = filepath.Join("/home/", globalOptions.User, deployDir)
			}
			// data dir would be empty for components which don't need it
			dataDir := monitoredOptions.DataDir
			if dataDir != "" && !strings.HasPrefix(dataDir, "/") {
				dataDir = filepath.Join("/home/", globalOptions.User, dataDir)
			}
			logDir := monitoredOptions.LogDir
			if !strings.HasPrefix(logDir, "/") {
				logDir = filepath.Join("/home/", globalOptions.User, logDir)
			}

			// Deploy component
			t := task.NewBuilder().
				UserSSH(host, globalOptions.User).
				Mkdir(host,
					filepath.Join(deployDir, "bin"),
					filepath.Join(deployDir, "conf"),
					filepath.Join(deployDir, "scripts"),
					dataDir,
					logDir).
				CopyComponent(comp, version, host, deployDir).
				MonitoredConfig(
					clusterName,
					comp,
					host,
					monitoredOptions,
					globalOptions.User,
					meta.DirPaths{
						Deploy: deployDir,
						Data:   dataDir,
						Log:    logDir,
					},
				).Build()
			deployCompTasks = append(deployCompTasks, t)
		}
	}
	return
}
