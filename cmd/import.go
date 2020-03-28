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
	"io/ioutil"

	"github.com/pingcap-incubator/tiops/pkg/meta"
	"github.com/pingcap-incubator/tiops/pkg/task"
	"github.com/pingcap/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func newImportCmd() *cobra.Command {
	var (
		ansible string
	)

	cmd := &cobra.Command{
		Use:    "import",
		Short:  "Import a TiDB cluster from tidb-ansible",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.Flags().StringVarP(&ansible, "ansible-path", "A", "", "the path for tidb-ansible")
	return cmd
}

// copy config file from cluster which deployed through tidb-ansible
func importConfig(name, topoFile string) error {
	var topo meta.TopologySpecification
	yamlFile, err := ioutil.ReadFile(topoFile)
	if err != nil {
		return errors.Trace(err)
	}
	if err = yaml.Unmarshal(yamlFile, &topo); err != nil {
		return errors.Trace(err)
	}
	// there may be already cluster dir, skip create
	//if err := os.MkdirAll(meta.ClusterPath(name), 0755); err != nil {
	//	return err
	//}
	//if err := ioutil.WriteFile(meta.ClusterPath(name, "topology.yaml"), yamlFile, 0664); err != nil {
	//	return err
	//}
	copyFileTasks := task.NewBuilder()
	for _, comp := range topo.ComponentsByStartOrder() {
		for idx, inst := range comp.Instances() {
			switch inst.ComponentName() {
			case "pd", "tikv", "pump", "tidb":
				if idx != 0 {
					break
				}
				copyFileTasks.
					UserSSH(inst.GetHost(), topo.GlobalOptions.User).
					CopyFile(inst.DeployDir()+"/conf/"+inst.ComponentName(),
						inst.GetHost(),
						meta.ClusterPath(name, "config", inst.ComponentName()+".toml"))
			case "dariner":
				copyFileTasks.
					UserSSH(inst.GetHost(), topo.GlobalOptions.User).
					CopyFile(inst.DeployDir()+"/conf/"+inst.ComponentName(),
						inst.GetHost(),
						meta.ClusterPath(name,
							"config",
							fmt.Sprintf("%s_%s_%s.toml", inst.ComponentName(), inst.GetHost(), inst.GetPort())))
			default:
				break
			}
		}
	}
	if err := copyFileTasks.Build().Execute(task.NewContext()); err != nil {
		return errors.Trace(err)
	} else {
		return nil
	}
}
