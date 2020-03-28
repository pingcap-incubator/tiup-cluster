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
	"path/filepath"
	"strconv"

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

// parseInventory builds a basic ClusterMeta from the main Ansible inventory
func parseInventory(dir string, inv *aini.InventoryData) (string, *meta.ClusterMeta, error) {
	topo := &meta.TopologySpecification{
		GlobalOptions:    meta.GlobalOptions{},
		MonitoredOptions: meta.MonitoredOptions{},
		TiDBServers:      make([]meta.TiDBSpec, 0),
		TiKVServers:      make([]meta.TiKVSpec, 0),
		PDServers:        make([]meta.PDSpec, 0),
		PumpServers:      make([]meta.PumpSpec, 0),
		Drainers:         make([]meta.DrainerSpec, 0),
		Monitors:         make([]meta.PrometheusSpec, 0),
		Grafana:          make([]meta.GrafanaSpec, 0),
		Alertmanager:     make([]meta.AlertManagerSpec, 0),
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

	// set global vars in group_vars/all.yml
	grpVarsAll, err := readGroupVars(dir, groupVarsGlobal)
	if err != nil {
		return "", nil, err
	}
	if port, ok := grpVarsAll["blackbox_exporter_port"]; ok {
		clsMeta.Topology.MonitoredOptions.BlackboxExporterPort, _ = strconv.Atoi(port)
	}
	if port, ok := grpVarsAll["node_exporter_port"]; ok {
		clsMeta.Topology.MonitoredOptions.NodeExporterPort, _ = strconv.Atoi(port)
	}

	// set hosts
	// tidb_servers
	if grp, ok := inv.Groups["tidb_servers"]; ok && len(grp.Hosts) > 0 {
		grpVars, err := readGroupVars(dir, groupVarsTiDB)
		if err != nil {
			return "", nil, err
		}
		for _, srv := range grp.Hosts {
			host := srv.Vars["ansible_host"]
			if host == "" {
				host = srv.Name
			}
			tmpIns := meta.TiDBSpec{
				Host:       host,
				SSHPort:    srv.Port,
				IsImported: true,
			}

			if port, ok := grpVars["tidb_port"]; ok {
				tmpIns.Port, _ = strconv.Atoi(port)
			}
			if statusPort, ok := grpVars["tidb_status_port"]; ok {
				tmpIns.StatusPort, _ = strconv.Atoi(statusPort)
			}

			// apply values from the host
			if port, ok := srv.Vars["tidb_port"]; ok {
				tmpIns.Port, _ = strconv.Atoi(port)
			}
			if statusPort, ok := srv.Vars["tidb_status_port"]; ok {
				tmpIns.StatusPort, _ = strconv.Atoi(statusPort)
			}
			if logDir, ok := srv.Vars["tidb_log_dir"]; ok {
				tmpIns.LogDir = logDir
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
			host := srv.Vars["ansible_host"]
			if host == "" {
				host = srv.Name
			}
			tmpIns := meta.TiKVSpec{
				Host:       host,
				SSHPort:    srv.Port,
				IsImported: true,
			}

			if port, ok := grpVars["tikv_port"]; ok {
				tmpIns.Port, _ = strconv.Atoi(port)
			}
			if statusPort, ok := grpVars["tikv_status_port"]; ok {
				tmpIns.StatusPort, _ = strconv.Atoi(statusPort)
			}

			// apply values from the host
			if port, ok := srv.Vars["tikv_port"]; ok {
				tmpIns.Port, _ = strconv.Atoi(port)
			}
			if statusPort, ok := srv.Vars["tikv_status_port"]; ok {
				tmpIns.StatusPort, _ = strconv.Atoi(statusPort)
			}
			if dataDir, ok := srv.Vars["tikv_data_dir"]; ok {
				tmpIns.DataDir = dataDir
			}
			if logDir, ok := srv.Vars["tikv_log_dir"]; ok {
				tmpIns.LogDir = logDir
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
			host := srv.Vars["ansible_host"]
			if host == "" {
				host = srv.Name
			}
			tmpIns := meta.PDSpec{
				Host:       host,
				SSHPort:    srv.Port,
				IsImported: true,
			}

			if clientPort, ok := grpVars["pd_client_port"]; ok {
				tmpIns.ClientPort, _ = strconv.Atoi(clientPort)
			}
			if peerPort, ok := grpVars["pd_peer_port"]; ok {
				tmpIns.PeerPort, _ = strconv.Atoi(peerPort)
			}

			// apply values from the host
			if clientPort, ok := srv.Vars["pd_client_port"]; ok {
				tmpIns.ClientPort, _ = strconv.Atoi(clientPort)
			}
			if peerPort, ok := srv.Vars["pd_peer_port"]; ok {
				tmpIns.PeerPort, _ = strconv.Atoi(peerPort)
			}
			if dataDir, ok := srv.Vars["pd_data_dir"]; ok {
				tmpIns.DataDir = dataDir
			}
			if logDir, ok := srv.Vars["pd_log_dir"]; ok {
				tmpIns.LogDir = logDir
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
			host := srv.Vars["ansible_host"]
			if host == "" {
				host = srv.Name
			}
			tmpIns := meta.PrometheusSpec{
				Host:       host,
				SSHPort:    srv.Port,
				IsImported: true,
			}

			if port, ok := grpVars["prometheus_port"]; ok {
				tmpIns.Port, _ = strconv.Atoi(port)
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
			host := srv.Vars["ansible_host"]
			if host == "" {
				host = srv.Name
			}
			tmpIns := meta.AlertManagerSpec{
				Host:       host,
				SSHPort:    srv.Port,
				IsImported: true,
			}

			if port, ok := grpVars["alertmanager_port"]; ok {
				tmpIns.WebPort, _ = strconv.Atoi(port)
			}
			if clusterPort, ok := grpVars["alertmanager_cluster_port"]; ok {
				tmpIns.ClusterPort, _ = strconv.Atoi(clusterPort)
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
			host := srv.Vars["ansible_host"]
			if host == "" {
				host = srv.Name
			}
			tmpIns := meta.PumpSpec{
				Host:       host,
				SSHPort:    srv.Port,
				IsImported: true,
			}

			// nothing in pump_servers.yml
			if port, ok := grpVarsAll["pump_port"]; ok {
				tmpIns.Port, _ = strconv.Atoi(port)
			}
			// apply values from the host
			if port, ok := srv.Vars["pump_port"]; ok {
				tmpIns.Port, _ = strconv.Atoi(port)
			}
			if dataDir, ok := srv.Vars["pump_data_dir"]; ok {
				tmpIns.DataDir = dataDir
			}
			if logDir, ok := srv.Vars["pump_log_dir"]; ok {
				tmpIns.LogDir = logDir
			}

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
			host := srv.Vars["ansible_host"]
			if host == "" {
				host = srv.Name
			}
			tmpIns := meta.DrainerSpec{
				Host:       host,
				SSHPort:    srv.Port,
				IsImported: true,
			}

			// nothing in drainer_servers.yml
			if port, ok := grpVarsAll["drainer_port"]; ok {
				tmpIns.Port, _ = strconv.Atoi(port)
			}

			ins, err := parseDirs(tmpIns)
			if err != nil {
				return "", nil, err
			}

			topo.Drainers = append(topo.Drainers, ins.(meta.DrainerSpec))
		}
	}

	// TODO: node_exporter and blackbox_exporter on custom port is not supported yet
	// if it is set on a host line. Global values in group_vars/all.yml will be
	// correctly parsed.

	return clsName, clsMeta, nil
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
