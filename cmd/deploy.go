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

	"github.com/joomcode/errorx"
	"github.com/pingcap-incubator/tiup/pkg/repository"
	"github.com/pingcap-incubator/tiup/pkg/set"
	tiuputils "github.com/pingcap-incubator/tiup/pkg/utils"
	"github.com/pingcap/errors"
	"github.com/spf13/cobra"

	"github.com/pingcap-incubator/tiops/pkg/cliutil"
	"github.com/pingcap-incubator/tiops/pkg/errutil"
	"github.com/pingcap-incubator/tiops/pkg/executor"
	"github.com/pingcap-incubator/tiops/pkg/logger"
	"github.com/pingcap-incubator/tiops/pkg/meta"
	"github.com/pingcap-incubator/tiops/pkg/task"
	"github.com/pingcap-incubator/tiops/pkg/utils"
)

var (
	errNS            = errorx.NewNamespace("cmd.deploy")
	errNameDuplicate = errNS.NewType("name_dup", errutil.ErrTraitPreCheck)
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
			shouldContinue, err := cliutil.CheckCommandArgsAndMayPrintHelp(cmd, args, 3)
			if err != nil {
				return err
			}
			if !shouldContinue {
				return nil
			}

			if opt.usePasswd {
				// FIXME: We should prompt for password when necessary automatically.
				opt.password = utils.GetPasswd("Password:")
			}

			if len(opt.keyFile) == 0 && !opt.usePasswd {
				// FIXME: We should lookup identity key automatically.
				return executor.ErrSSHRequireCredential.
					New("Identity file and password is unspecified").
					WithProperty(cliutil.SuggestionFromTemplate(`
You should specify either SSH identity file or password.

To SSH connect using identity file:
  {{ColorCommand}}{{OsArgs}} --identity_file <file>{{ColorReset}}

To SSH connect using password:
  {{ColorCommand}}{{OsArgs}} --password{{ColorReset}}

`, nil))
			}

			logger.EnableAuditLog()
			return deploy(args[0], args[1], args[2], opt)
		},
	}

	cmd.Flags().StringVar(&opt.user, "user", "root", "Specify the system user name")
	cmd.Flags().BoolVar(&opt.usePasswd, "password", false, "Specify the password of system user")
	cmd.Flags().StringVar(&opt.keyFile, "identity_file", "", "Specify the path of the SSH identity file")
	// FIXME: We should prompt for passphrase automatically
	cmd.Flags().StringVar(&opt.passphrase, "passphrase", "", "Specify the passphrase of the SSH identity file")

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
		// FIXME: When change to use args, the suggestion text need to be updated.
		return errNameDuplicate.
			New("Cluster name '%s' is duplicated", clusterName).
			WithProperty(cliutil.SuggestionFromFormat("Please specify another cluster name"))
	}

	var topo meta.TopologySpecification
	if err := utils.ParseTopologyYaml(topoFile, &topo); err != nil {
		return err
	}
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
		version := getComponentVersion(inst.ComponentName(), version)
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
			CopyComponent(inst.ComponentName(), version, inst.GetHost(), deployDir).
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
