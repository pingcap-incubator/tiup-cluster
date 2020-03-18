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

package meta

import (
	"github.com/creasty/defaults"
	//"gopkg.in/yaml.v2"
)

// TiDBSpec represents the TiDB topology specification in topology.yml
type TiDBSpec struct {
	IP         string `yml:"ip"`
	Port       int    `yml:"port" default:"4000"`
	StatusPort int    `yml:"status_port" default:"10080"`
	UUID       string `yml:"uuid,omitempty"`
	SSHPort    int    `yml:"ssh_port,omitempty" default:"22"`
	DeployDir  string `yml:"deploy_dir"`
	NumaNode   bool   `yml:"numa_node,omitempty" default:"false"`
}

// TiKVSpec represents the TiKV topology specification in topology.yml
type TiKVSpec struct {
	IP         string   `yml:"ip"`
	Port       int      `yml:"port" default:"20160"`
	StatusPort int      `yml:"status_port" default:"20180"`
	UUID       string   `yml:"uuid,omitempty"`
	SSHPort    int      `yml:"ssh_port,omitempty" default:"22"`
	DeployDir  string   `yml:"deploy_dir"`
	DataDir    string   `yml:"data_dir"`
	Offline    bool     `yml:"offline" default:"false"`
	Labels     []string `yml:"labels,omitempty"`
	NumaNode   bool     `yml:"numa_node,omitempty" default:"false"`
}

// PDSpec represents the PD topology specification in topology.yml
type PDSpec struct {
	IP         string `yml:"ip"`
	ClientPort int    `yml:"client_port" default:"2379"`
	PeerPort   int    `yml:"peer_port" default:"2380"`
	UUID       string `yml:"uuid,omitempty"`
	SSHPort    int    `yml:"ssh_port,omitempty" default:"22"`
	DeployDir  string `yml:"deploy_dir"`
	DataDir    string `yml:"data_dir"`
	NumaNode   bool   `yml:"numa_node,omitempty" default:"false"`
}

// PumpSpec represents the Pump topology specification in topology.yml
type PumpSpec struct {
	IP        string `yml:"ip"`
	Port      int    `yml:"port" default:"8250"`
	UUID      string `yml:"uuid,omitempty"`
	SSHPort   int    `yml:"ssh_port,omitempty" default:"22"`
	DeployDir string `yml:"deploy_dir"`
	DataDir   string `yml:"data_dir"`
	Offline   bool   `yml:"offline" default:"false"`
	NumaNode  bool   `yml:"numa_node,omitempty" default:"false"`
}

// DrainerSpec represents the Drainer topology specification in topology.yml
type DrainerSpec struct {
	IP        string `yml:"ip"`
	Port      int    `yml:"port" default:"8249"`
	UUID      string `yml:"uuid,omitempty"`
	SSHPort   int    `yml:"ssh_port,omitempty" default:"22"`
	DeployDir string `yml:"deploy_dir"`
	DataDir   string `yml:"data_dir"`
	CommitTS  string `yml:"commit_ts"`
	Offline   bool   `yml:"offline" default:"false"`
	NumaNode  bool   `yml:"numa_node,omitempty" default:"false"`
}

// PrometheusSpec represents the Prometheus Server topology specification in topology.yml
type PrometheusSpec struct {
	IP        string `yml:"ip"`
	Port      int    `yml:"port" default:"9090"`
	UUID      string `yml:"uuid,omitempty"`
	SSHPort   int    `yml:"ssh_port,omitempty" default:"22"`
	DeployDir string `yml:"deploy_dir"`
	DataDir   string `yml:"data_dir"`
}

// GrafanaSpec represents the Grafana topology specification in topology.yml
type GrafanaSpec struct {
	IP        string `yml:"ip"`
	Port      int    `yml:"port" default:"3000"`
	UUID      string `yml:"uuid,omitempty"`
	SSHPort   int    `yml:"ssh_port,omitempty" default:"22"`
	DeployDir string `yml:"deploy_dir"`
}

// AlertManagerSpec represents the AlertManager topology specification in topology.yml
type AlertManagerSpec struct {
	IP          string `yml:"ip"`
	WebPort     int    `yml:"web_port" default:"9093"`
	ClusterPort int    `yml:"cluster_port" default:"9094"`
	UUID        string `yml:"uuid,omitempty"`
	SSHPort     int    `yml:"ssh_port,omitempty" default:"22"`
	DeployDir   string `yml:"deploy_dir"`
	DataDir     string `yml:"data_dir"`
}

/*
// TopologyGlobalOptions represents the global options for all groups in topology
// pecification in topology.yml
type TopologyGlobalOptions struct {
	SSHPort              int    `yml:"ssh_port,omitempty" default:"22"`
	DeployDir            string `yml:"deploy_dir,omitempty"`
	DataDir              string `yml:"data_dir,omitempty"`
	NodeExporterPort     int    `yml:"node_exporter_port,omitempty" default:"9100"`
	BlackboxExporterPort int    `yml:"blackbox_exporter_port,omitempty" default:"9115"`
}
*/

// TopologySpecification represents the specification of topology.yml
type TopologySpecification struct {
	//GlobalOptions TopologyGlobalOptions `yml:"global,omitempty"`
	TiDBServers  []TiDBSpec       `yml:"tidb_servers"`
	TiKVServers  []TiKVSpec       `yml:"tikv_servers"`
	PDServers    []PDSpec         `yml:"pd_servers"`
	PumpServers  []PumpSpec       `yml:"pump_servers,omitempty"`
	Drainers     []DrainerSpec    `yml:"drainer_servers,omitempty"`
	MonitorSpec  []PrometheusSpec `yml:"monitoring_server"`
	Grafana      GrafanaSpec      `yml:"grafana_server,omitempty"`
	Alertmanager AlertManagerSpec `yml:"alertmanager_server,omitempty"`
}

// UnmarshalYAML sets default values when unmarshaling the topology file
func (t *TopologySpecification) UnmarshalYAML(unmarshal func(interface{}) error) error {
	defaults.Set(t)

	if err := unmarshal(t); err != nil {
		return err
	}
	return nil
}
