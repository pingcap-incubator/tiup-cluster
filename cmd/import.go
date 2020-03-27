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
	"os"
	"path/filepath"

	"github.com/pingcap-incubator/tiops/pkg/meta"
	"github.com/pingcap-incubator/tiops/pkg/task"
	"github.com/pingcap/errors"
	"github.com/relex/aini"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func newImportCmd() *cobra.Command {
	var (
		ansibleDir string
	)

	cmd := &cobra.Command{
		Use:    "import [OPTIONS]",
		Short:  "Import a TiDB cluster from tidb-ansible",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return importAnsible(ansibleDir)
		},
	}

	cmd.Flags().StringVarP(&ansibleDir, "dir", "d", "", "The path to TiDB-Ansible's directory")

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

var (
	ansibleInventoryFile = "inventory.ini"
)

func importAnsible(dir string) error {
	inventoryFile, err := os.Open(filepath.Join(dir, ansibleInventoryFile))
	if err != nil {
		return err
	}
	defer inventoryFile.Close()

	inventory, err := aini.Parse(inventoryFile)
	if err != nil {
		return err
	}

	topo, err := parseInventory(inventory)
	if err != nil {
		return err
	}
	fmt.Printf("%v\n", topo)

	return nil
}

func parseInventory(inv *aini.InventoryData) (*meta.TopologySpecification, error) {
	topo := &meta.TopologySpecification{
		TiDBServers:  make([]meta.TiDBSpec, 0),
		TiKVServers:  make([]meta.TiKVSpec, 0),
		PDServers:    make([]meta.PDSpec, 0),
		PumpServers:  make([]meta.PumpSpec, 0),
		Drainers:     make([]meta.DrainerSpec, 0),
		Monitors:     make([]meta.PrometheusSpec, 0),
		Grafana:      make([]meta.GrafanaSpec, 0),
		Alertmanager: make([]meta.AlertManagerSpec, 0),
	}

	// set hosts
	// tidb_servers
	if grp, ok := inv.Groups["tidb_servers"]; ok && len(grp.Hosts) > 0 {
		for _, srv := range grp.Hosts {
			topo.TiDBServers = append(topo.TiDBServers, meta.TiDBSpec{
				Host:    srv.Name,
				SSHPort: srv.Port,
			})
		}
	}

	// tikv_servers
	if grp, ok := inv.Groups["tikv_servers"]; ok && len(grp.Hosts) > 0 {
		for _, srv := range grp.Hosts {
			topo.TiKVServers = append(topo.TiKVServers, meta.TiKVSpec{
				Host:    srv.Name,
				SSHPort: srv.Port,
			})
		}
	}

	// pd_servers
	if grp, ok := inv.Groups["pd_servers"]; ok && len(grp.Hosts) > 0 {
		for _, srv := range grp.Hosts {
			topo.PDServers = append(topo.PDServers, meta.PDSpec{
				Host:    srv.Name,
				SSHPort: srv.Port,
			})
		}
	}

	// spark_master
	// spark_slaves
	// lightning_server
	// importer_server

	// monitoring_servers
	if grp, ok := inv.Groups["monitoring_servers"]; ok && len(grp.Hosts) > 0 {
		for _, srv := range grp.Hosts {
			topo.Monitors = append(topo.Monitors, meta.PrometheusSpec{
				Host:    srv.Name,
				SSHPort: srv.Port,
			})
		}
	}

	// monitored_servers

	// alertmanager_servers
	if grp, ok := inv.Groups["alertmanager_servers"]; ok && len(grp.Hosts) > 0 {
		for _, srv := range grp.Hosts {
			topo.Alertmanager = append(topo.Alertmanager, meta.AlertManagerSpec{
				Host:    srv.Name,
				SSHPort: srv.Port,
			})
		}
	}

	// kafka_exporter_servers

	// pump_servers
	if grp, ok := inv.Groups["pump_servers"]; ok && len(grp.Hosts) > 0 {
		for _, srv := range grp.Hosts {
			topo.PumpServers = append(topo.PumpServers, meta.PumpSpec{
				Host:    srv.Name,
				SSHPort: srv.Port,
			})
		}
	}

	// drainer_servers
	if grp, ok := inv.Groups["drainer_servers"]; ok && len(grp.Hosts) > 0 {
		for _, srv := range grp.Hosts {
			topo.Drainers = append(topo.Drainers, meta.DrainerSpec{
				Host:    srv.Name,
				SSHPort: srv.Port,
			})
		}
	}

	return topo, nil
}
