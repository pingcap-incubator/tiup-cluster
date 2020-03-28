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
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/creasty/defaults"
	"github.com/pingcap-incubator/tiops/pkg/meta"
	"github.com/relex/aini"
	"gopkg.in/yaml.v2"
)

var (
	ansibleInventoryFile  = "inventory.ini"
	groupVarsGlobal       = "group_vars/all.yml"
	groupVarsTiDB         = "group_vars/tidb_servers.yml"
	groupVarsTiKV         = "group_vars/tikv_servers.yml"
	groupVarsPD           = "group_vars/pd_servers.yml"
	groupVarsPump         = "group_vars/pump_servers.yml"
	groupVarsDrainer      = "group_vars/drainer_servers.yml"
	groupVarsAlertManager = "group_vars/alertmanager_servers.yml"
	groupVarsGrafana      = "group_vars/grafana_servers.yml"
	groupVarsMonitorAgent = "group_vars/monitored_servers.yml"
	groupVarsPrometheus   = "group_vars/monitoring_servers.yml"
	//groupVarsLightning    = "group_vars/lightning_server.yml"
	//groupVarsImporter     = "group_vars/importer_server.yml"
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

	clsName, clsMeta, err := parseInventory(dir, inventory)
	if err != nil {
		return err
	}

	// TODO: add output of imported cluster name and version
	// TODO: check cluster name with other clusters managed by us for conflicts
	// TODO: prompt user for a chance to set a new cluster name

	// TODO: get values from templates of roles to overwrite defaults
	defaults.Set(clsMeta)
	return meta.SaveClusterMeta(clsName, clsMeta)
}

// parseInventory builds a basic ClusterMeta from the main Ansible inventory
func parseInventory(dir string, inv *aini.InventoryData) (string, *meta.ClusterMeta, error) {
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
		grpVars, err := readGroupVars(dir, groupVarsTiDB)
		if err != nil {
			return "", nil, err
		}
		for _, srv := range grp.Hosts {
			tmpIns := meta.TiDBSpec{
				Host:       srv.Name,
				SSHPort:    srv.Port,
				IsImported: true,
			}

			if _port, ok := grpVars["tidb_port"]; ok {
				tmpIns.Port, _ = strconv.Atoi(_port)
			}
			if _status_port, ok := grpVars["tidb_status_port"]; ok {
				tmpIns.StatusPort, _ = strconv.Atoi(_status_port)
			}

			ins, err := parseDirs(tmpIns)
			if err != nil {
				return "", nil, err
			}

			topo.TiDBServers = append(topo.TiDBServers, ins.(meta.TiDBSpec))
		}
	}

	// tikv_servers
	if grp, ok := inv.Groups["tikv_servers"]; ok && len(grp.Hosts) > 0 {
		grpVars, err := readGroupVars(dir, groupVarsTiKV)
		if err != nil {
			return "", nil, err
		}
		for _, srv := range grp.Hosts {
			tmpIns := meta.TiKVSpec{
				Host:       srv.Name,
				SSHPort:    srv.Port,
				IsImported: true,
			}

			if _port, ok := grpVars["tikv_port"]; ok {
				tmpIns.Port, _ = strconv.Atoi(_port)
			}
			if _status_port, ok := grpVars["tikv_status_port"]; ok {
				tmpIns.StatusPort, _ = strconv.Atoi(_status_port)
			}

			ins, err := parseDirs(tmpIns)
			if err != nil {
				return "", nil, err
			}

			topo.TiKVServers = append(topo.TiKVServers, ins.(meta.TiKVSpec))
		}
	}

	// pd_servers
	if grp, ok := inv.Groups["pd_servers"]; ok && len(grp.Hosts) > 0 {
		grpVars, err := readGroupVars(dir, groupVarsPD)
		if err != nil {
			return "", nil, err
		}
		for _, srv := range grp.Hosts {
			tmpIns := meta.PDSpec{
				Host:       srv.Name,
				SSHPort:    srv.Port,
				IsImported: true,
			}

			if _port, ok := grpVars["pd_client_port"]; ok {
				tmpIns.ClientPort, _ = strconv.Atoi(_port)
			}
			if _status_port, ok := grpVars["pd_peer_port"]; ok {
				tmpIns.PeerPort, _ = strconv.Atoi(_status_port)
			}

			ins, err := parseDirs(tmpIns)
			if err != nil {
				return "", nil, err
			}

			topo.PDServers = append(topo.PDServers, ins.(meta.PDSpec))
		}
	}

	// spark_master
	// spark_slaves
	// lightning_server
	// importer_server

	// monitoring_servers
	if grp, ok := inv.Groups["monitoring_servers"]; ok && len(grp.Hosts) > 0 {
		grpVars, err := readGroupVars(dir, groupVarsPrometheus)
		if err != nil {
			return "", nil, err
		}
		for _, srv := range grp.Hosts {
			tmpIns := meta.PrometheusSpec{
				Host:       srv.Name,
				SSHPort:    srv.Port,
				IsImported: true,
			}

			if _port, ok := grpVars["prometheus_port"]; ok {
				tmpIns.Port, _ = strconv.Atoi(_port)
			}
			// pushgateway no longer needed, just ignore
			// NOTE: storage retention is not used at present, only for record
			if _retention, ok := grpVars["prometheus_storage_retention"]; ok {
				tmpIns.Retention = _retention
			}

			ins, err := parseDirs(tmpIns)
			if err != nil {
				return "", nil, err
			}

			topo.Monitors = append(topo.Monitors, ins.(meta.PrometheusSpec))
		}
	}

	// monitored_servers
	// ^- ignore, we use auto generated full list

	// alertmanager_servers
	if grp, ok := inv.Groups["alertmanager_servers"]; ok && len(grp.Hosts) > 0 {
		grpVars, err := readGroupVars(dir, groupVarsAlertManager)
		if err != nil {
			return "", nil, err
		}
		for _, srv := range grp.Hosts {
			tmpIns := meta.AlertManagerSpec{
				Host:       srv.Name,
				SSHPort:    srv.Port,
				IsImported: true,
			}

			if _port, ok := grpVars["alertmanager_port"]; ok {
				tmpIns.WebPort, _ = strconv.Atoi(_port)
			}
			if _cluster_port, ok := grpVars["alertmanager_cluster_port"]; ok {
				tmpIns.ClusterPort, _ = strconv.Atoi(_cluster_port)
			}

			ins, err := parseDirs(tmpIns)
			if err != nil {
				return "", nil, err
			}

			topo.Alertmanager = append(topo.Alertmanager, ins.(meta.AlertManagerSpec))
		}
	}

	// kafka_exporter_servers

	// pump_servers
	if grp, ok := inv.Groups["pump_servers"]; ok && len(grp.Hosts) > 0 {
		/*
			grpVars, err := readGroupVars(dir, groupVarsPump)
			if err != nil {
				return "", nil, err
			}
		*/
		for _, srv := range grp.Hosts {
			tmpIns := meta.PumpSpec{
				Host:       srv.Name,
				SSHPort:    srv.Port,
				IsImported: true,
			}

			// nothing in pump_servers.yml

			ins, err := parseDirs(tmpIns)
			if err != nil {
				return "", nil, err
			}

			topo.PumpServers = append(topo.PumpServers, ins.(meta.PumpSpec))
		}
	}

	// drainer_servers
	if grp, ok := inv.Groups["drainer_servers"]; ok && len(grp.Hosts) > 0 {
		/*
			grpVars, err := readGroupVars(dir, groupVarsDrainer)
			if err != nil {
				return "", nil, err
			}
		*/
		for _, srv := range grp.Hosts {
			tmpIns := meta.DrainerSpec{
				Host:       srv.Name,
				SSHPort:    srv.Port,
				IsImported: true,
			}

			// nothing in drainer_servers.yml

			ins, err := parseDirs(tmpIns)
			if err != nil {
				return "", nil, err
			}

			topo.Drainers = append(topo.Drainers, ins.(meta.DrainerSpec))
		}
	}

	return clsName, clsMeta, nil
}

// parseDirs sets values of directories of component
func parseDirs(ins meta.InstanceSpec) (meta.InstanceSpec, error) {
	switch ins.Role() {
	case meta.RoleTiDB:
	case meta.RoleTiKV:
	case meta.RolePD:
	case meta.RolePump:
	case meta.RoleDrainer:
	case meta.RoleMonitor:
	case meta.RoleGrafana:
	}
	return ins, nil
}

// readGroupVars sets values from configs in group_vars/ dir
func readGroupVars(dir, filename string) (map[string]string, error) {
	result := make(map[string]string)

	fileData, err := ioutil.ReadFile(filepath.Join(dir, filename))
	if err != nil {
		return nil, err
	}

	if err = yaml.Unmarshal(fileData, &result); err != nil {
		return nil, err
	}
	return result, nil
}
