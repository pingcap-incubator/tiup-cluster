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

package ansible

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/creasty/defaults"
	"github.com/pingcap-incubator/tiops/pkg/meta"
	"github.com/relex/aini"
)

var (
	ansibleInventoryFile = "inventory.ini"
)

// ImportAnsible imports a TiDB cluster deployed by TiDB-Ansible
func ImportAnsible(dir string) error {
	inventoryFile, err := os.Open(filepath.Join(dir, ansibleInventoryFile))
	if err != nil {
		return err
	}
	defer inventoryFile.Close()

	inventory, err := aini.Parse(inventoryFile)
	if err != nil {
		return err
	}

	clsName, clsMeta, err := parseInventory(inventory)
	if err != nil {
		return err
	}

	// TODO: add output of imported cluster name and version
	// TODO: check cluster name with other clusters managed by us for conflicts
	// TODO: prompt user for a chance to set a new cluster name

	clsMeta, err = parseGroupVars(clsName, clsMeta)

	// TODO: get values from templates of roles to overwrite defaults
	defaults.Set(clsMeta)
	return meta.SaveClusterMeta(clsName, clsMeta)
}

// parseInventory builds a basic ClusterMeta from the main Ansible inventory
func parseInventory(inv *aini.InventoryData) (string, *meta.ClusterMeta, error) {
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
	clsMeta := &meta.ClusterMeta{
		Topology: topo,
	}
	clsName := ""

	// get global vars
	if grp, ok := inv.Groups["all"]; ok && len(grp.Hosts) > 0 {
		for _, host := range grp.Hosts {
			if host.Vars["process_supervision"] != "systemd" {
				return "", nil, errors.New("only support cluster deployed with systemd")
			}
			clsMeta.User = host.Vars["ansible_user"]
			clsMeta.Version = host.Vars["tidb_version"]
			clsName = host.Vars["cluster_name"]

			// only read the first host, all global vars should be the same
			break
		}
	} else {
		return "", nil, errors.New("no available host in the inventory file")
	}

	// set hosts
	// tidb_servers
	if grp, ok := inv.Groups["tidb_servers"]; ok && len(grp.Hosts) > 0 {
		for _, srv := range grp.Hosts {
			topo.TiDBServers = append(topo.TiDBServers, meta.TiDBSpec{
				Host:      srv.Name,
				SSHPort:   srv.Port,
				DeployDir: srv.Vars["deploy_dir"],
			})
		}
	}

	// tikv_servers
	if grp, ok := inv.Groups["tikv_servers"]; ok && len(grp.Hosts) > 0 {
		for _, srv := range grp.Hosts {
			topo.TiKVServers = append(topo.TiKVServers, meta.TiKVSpec{
				Host:      srv.Name,
				SSHPort:   srv.Port,
				DeployDir: srv.Vars["deploy_dir"],
			})
		}
	}

	// pd_servers
	if grp, ok := inv.Groups["pd_servers"]; ok && len(grp.Hosts) > 0 {
		for _, srv := range grp.Hosts {
			topo.PDServers = append(topo.PDServers, meta.PDSpec{
				Host:      srv.Name,
				SSHPort:   srv.Port,
				DeployDir: srv.Vars["deploy_dir"],
			})
			fmt.Printf("%s\n", srv.Vars)
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
				Host:      srv.Name,
				SSHPort:   srv.Port,
				DeployDir: srv.Vars["deploy_dir"],
			})
		}
	}

	// monitored_servers
	// ^- ignore, we use auto generated full list

	// alertmanager_servers
	if grp, ok := inv.Groups["alertmanager_servers"]; ok && len(grp.Hosts) > 0 {
		for _, srv := range grp.Hosts {
			topo.Alertmanager = append(topo.Alertmanager, meta.AlertManagerSpec{
				Host:      srv.Name,
				SSHPort:   srv.Port,
				DeployDir: srv.Vars["deploy_dir"],
			})
		}
	}

	// kafka_exporter_servers

	// pump_servers
	if grp, ok := inv.Groups["pump_servers"]; ok && len(grp.Hosts) > 0 {
		for _, srv := range grp.Hosts {
			topo.PumpServers = append(topo.PumpServers, meta.PumpSpec{
				Host:      srv.Name,
				SSHPort:   srv.Port,
				DeployDir: srv.Vars["deploy_dir"],
			})
		}
	}

	// drainer_servers
	if grp, ok := inv.Groups["drainer_servers"]; ok && len(grp.Hosts) > 0 {
		for _, srv := range grp.Hosts {
			topo.Drainers = append(topo.Drainers, meta.DrainerSpec{
				Host:      srv.Name,
				SSHPort:   srv.Port,
				DeployDir: srv.Vars["deploy_dir"],
			})
		}
	}

	return clsName, clsMeta, nil
}

// parseGroupVars sets values in the group_vars/ configs
func parseGroupVars(clsName string, clsMeta *meta.ClusterMeta) (*meta.ClusterMeta, error) {
	return clsMeta, nil
}
