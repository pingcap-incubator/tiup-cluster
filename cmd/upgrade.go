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
	"path/filepath"
	"strings"

	"github.com/pingcap-incubator/tiops/pkg/meta"
	operator "github.com/pingcap-incubator/tiops/pkg/operation"
	"github.com/pingcap-incubator/tiops/pkg/task"
	"github.com/pingcap/errors"
	"github.com/spf13/cobra"
)

type upgradeOptions struct {
	cluster string
	version string
	options operator.Options
}

func newUpgradeCmd() *cobra.Command {
	opt := upgradeOptions{}
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade a TiDB cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return upgrade(opt)
		},
	}

	cmd.Flags().StringVarP(&opt.cluster, "cluster", "c", "", "Specify the cluster name")
	cmd.Flags().StringVarP(&opt.version, "target-version", "t", "", "Specify the target version")
	cmd.Flags().BoolVar(&opt.options.Force, "force", false, "Force upgrade won't transfer leader")

	_ = cmd.MarkFlagRequired("cluster")
	_ = cmd.MarkFlagRequired("target-version")

	return cmd
}

func upgrade(opt upgradeOptions) error {
	metadata, err := meta.ClusterMetadata(opt.cluster)
	if err != nil {
		return err
	}

	var (
		downloadCompTasks []task.Task // tasks which are used to download components
		copyCompTasks     []task.Task // tasks which are used to copy components to remote host

		uniqueComps = map[componentInfo]struct{}{}
	)

	for _, comp := range metadata.Topology.ComponentsByStartOrder() {
		for _, inst := range comp.Instances() {
			version := getComponentVersion(inst.ComponentName(), opt.version)
			if version == "" {
				return errors.Errorf("unsupported component: %v", inst.ComponentName())
			}
			compInfo := componentInfo{
				component: inst.ComponentName(),
				version:   version,
			}

			// Download component from repository
			if _, found := uniqueComps[compInfo]; !found {
				uniqueComps[compInfo] = struct{}{}
				t := task.NewBuilder().
					Download(inst.ComponentName(), version).
					Build()
				downloadCompTasks = append(downloadCompTasks, t)
			}

			deployDir := inst.DeployDir()
			if !strings.HasPrefix(deployDir, "/") {
				deployDir = filepath.Join("/home/"+metadata.User+"/deploy", deployDir)
			}
			// Deploy component
			t := task.NewBuilder().
				BackupComponent(inst.ComponentName(), metadata.Version, inst.GetHost(), deployDir).
				CopyComponent(inst.ComponentName(), version, inst.GetHost(), deployDir).
				Build()
			copyCompTasks = append(copyCompTasks, t)
		}
	}

	t := task.NewBuilder().
		SSHKeySet(
			meta.ClusterPath(opt.cluster, "ssh", "id_rsa"),
			meta.ClusterPath(opt.cluster, "ssh", "id_rsa.pub")).
		ClusterSSH(metadata.Topology, metadata.User).
		Parallel(downloadCompTasks...).
		Parallel(copyCompTasks...).
		ClusterOperate(metadata.Topology, operator.UpgradeOperation, opt.options).
		Build()

	return t.Execute(task.NewContext())
}
