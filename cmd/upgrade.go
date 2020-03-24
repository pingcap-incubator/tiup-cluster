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
	"fmt"
	"github.com/pingcap-incubator/tiops/pkg/meta"
	"github.com/pingcap-incubator/tiops/pkg/task"
	"github.com/pingcap-incubator/tiup/pkg/repository"
	"github.com/pingcap/errors"
	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
)

type upgradeOptions struct {
	version string
	force   bool
	node    string
	role    string
}

func newUpgradeCmd() *cobra.Command {
	opt := upgradeOptions{}
	cmd := &cobra.Command{
		Use:   "upgrade <cluster-name>",
		Short: "Upgrade a specified TiDB cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return cmd.Help()
			}
			metaInfo, err := meta.ClusterMetadata(args[0])
			if err != nil {
				return err
			}
			curVersion := metaInfo.Version
			newVersion := opt.version
			if err := versionCompare(curVersion, newVersion); err != nil {
				return err
			}
			return upgrade(args[0], opt) // TODO
		},
	}

	cmd.Flags().StringVar(&opt.version, "version", "", "version of cluster")
	cmd.Flags().BoolVar(&opt.force, "force", false, "upgrade without transfer leader, fast but affects stability (default false)")
	cmd.Flags().StringVar(&opt.node, "node-id", "", "node id")
	cmd.Flags().StringVar(&opt.role, "role", "", "role name")
	return cmd
}

func versionCompare(curVersion, newVersion string) error {

	switch semver.Compare(curVersion, newVersion) {
	case -1:
		return nil
	case 1:
		if newVersion == "nightly" {
			return nil
		} else {
			return errors.New(fmt.Sprintf("unsupport upgrade from %s to %s", curVersion, newVersion))
		}
	default:
		return errors.New("unkown error")
	}
}

// TODO
func upgrade(name string, opt upgradeOptions) error {
	topo, err := meta.ClusterTopology(name)
	if err != nil {
		return err
	}

	type componentInfo struct {
		component string
		version   repository.Version
	}

	var (
		uniqueComps = map[componentInfo]struct{}{}
		//upgradeTasks []task.Task
	)
	upgradeTasks := task.NewBuilder()

	for _, comp := range topo.ComponentsByStartOrder() {
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
				upgradeTasks.Download(inst.ComponentName(), version)
			}

			deployDir := inst.DeployDir()

			switch inst.ComponentName() {
			case meta.ComponentPD:
				return nil // TODO: 1. check pd cluster status; 2. split pd node to follower and leader; 3. upgrade follower node; 4. transfer leader first and upgrade leader node
			case meta.ComponentTiKV:
				return nil // TODO: 1. add evict scheduler in pd to evict all region leader until leader count is 0; 2. upgrade tikv; 3. remove evict scheduler from pd
			case meta.ComponentPump:
				return nil // TODO(need to discuss, maybe step is): 1. stop pump; 2. upgrade; 3. start
			default:
				upgradeTasks.ClusterOperate(topo, "stop", inst.ComponentName(), inst.GetUuid()).
					CopyComponent(inst.ComponentName(), version, inst.GetHost(), deployDir).
					CopyConfig(name, topo, inst.ComponentName(), inst.GetHost(), inst.GetPort(), deployDir).
					ClusterOperate(topo, "start", inst.ComponentName(), inst.GetUuid())
				//return nil // TODO: 1. stop; 2. upgrade; 3. start

			}
		}
	}
	return upgradeTasks.Build().Execute(task.NewContext())
}
