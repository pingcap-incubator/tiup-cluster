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

package topology

import (
	"strconv"

	"github.com/pingcap-incubator/tiops/pkg/executor"
)

// TiDBSpec represents the TiDB topology specification in topology.yml
type TiDBSpec struct {
	IP         string `yml:"ip"`
	Port       int    `yml:"port"`
	StatusPort int    `yml:"status_port"`
	DeployDir  string `yml:"deploy_dir"`
}

// TiKVSpec represents the TiKV topology specification in topology.yml
type TiKVSpec struct {
	IP         string   `yml:"ip"`
	Port       int      `yml:"port"`
	StatusPort int      `yml:"status_port"`
	DeployDir  string   `yml:"deploy_dir"`
	Labels     []string `yml:"labels"`
}

// PDSpec represents the PD topology specification in topology.yml
type PDSpec struct {
	IP         string `yml:"ip"`
	ClientPort int    `yml:"client_port"`
	PeerPort   int    `yml:"peer_port"`
	DataDir    string `yml:"data_dir"`
	DeployDir  string `yml:"deploy_dir"`
}

// PumpSpec represents the Pump topology specification in topology.yml
type PumpSpec struct {
	IP        string `yml:"ip"`
	Port      int    `yml:"port"`
	DataDir   string `yml:"data_dir"`
	DeployDir string `yml:"deploy_dir"`
}

// DrainerSpec represents the Drainer topology specification in topology.yml
type DrainerSpec struct {
	IP        string `yml:"ip"`
	Port      int    `yml:"port"`
	DataDir   string `yml:"data_dir"`
	DeployDir string `yml:"deploy_dir"`
}

// MonitorSpec represents the Monitor topology specification in topology.yml
type MonitorSpec struct {
	IP              string `yml:"ip"`
	PrometheusPort  int    `yml:"prometheus_port"`
	PushGatewayPort int    `yml:"pushgateway_port"`
	DeployDir       string `yml:"deploy_dir"`
}

// GrafanaSpec represents the Grafana topology specification in topology.yml
type GrafanaSpec struct {
	IP        string `yml:"ip"`
	Port      int    `yml:"port"`
	DeployDir string `yml:"deploy_dir"`
}

// AlertManagerSpec represents the AlertManager topology specification in topology.yml
type AlertManagerSpec struct {
	IP          string `yml:"ip"`
	Port        int    `yml:"port"`
	ClusterPort int    `yml:"cluster_port"`
	DeployDir   string `yml:"deploy_dir"`
	DataDir     string `yml:"data_dir"`
}

// Specification represents the specification of topology.yml
type Specification struct {
	TiDBServers  []TiDBSpec
	TiKVServers  []TiKVSpec
	PDServers    []PDSpec
	PumpServers  []PumpSpec
	Drainers     []DrainerSpec
	MonitorSpec  []MonitorSpec
	Grafana      []GrafanaSpec
	Alertmanager []AlertManagerSpec
}

// implements instance.
//
var _ Instance = &TiDBSpec{}

// Ready implements Instance interface.
func (s *TiDBSpec) Ready(e executor.TiOpsExecutor) error {
	return nil
}

// ComponentName implements Instance interface.
func (s *TiDBSpec) ComponentName() string {
	return "tidb"
}

// ServiceName implements Instance interface.
func (s *TiDBSpec) ServiceName() string {
	return "tidb-" + strconv.Itoa(s.Port) + ".service"
}

// GetIP implements Instance interface.
func (s *TiDBSpec) GetIP() string {
	return s.IP
}

var _ Instance = &TiKVSpec{}

// Ready implements Instance interface.
func (s *TiKVSpec) Ready(e executor.TiOpsExecutor) error {
	return nil
}

// ComponentName implements Instance interface.
func (s *TiKVSpec) ComponentName() string {
	return "tikv"
}

// ServiceName implements Instance interface.
func (s *TiKVSpec) ServiceName() string {
	return "tikv-" + s.IP + ".service"
}

// GetIP implements Instance interface.
func (s *TiKVSpec) GetIP() string {
	return s.IP
}

var _ Instance = &PDSpec{}

// Ready implements Instance interface.
func (s *PDSpec) Ready(e executor.TiOpsExecutor) error {
	return nil
}

// ComponentName implements Instance interface.
func (s *PDSpec) ComponentName() string {
	return "pd"
}

// ServiceName implements Instance interface.
func (s *PDSpec) ServiceName() string {
	return "pd-" + s.IP + ".service"
}

// GetIP implements Instance interface.
func (s *PDSpec) GetIP() string {
	return s.IP
}

var _ Instance = &PumpSpec{}

// Ready implements Instance interface.
func (s *PumpSpec) Ready(e executor.TiOpsExecutor) error {
	return nil
}

// ComponentName implements Instance interface.
func (s *PumpSpec) ComponentName() string {
	return "pump"
}

// ServiceName implements Instance interface.
func (s *PumpSpec) ServiceName() string {
	return "pump-" + s.IP + ".service"
}

// GetIP implements Instance interface.
func (s *PumpSpec) GetIP() string {
	return s.IP
}

var _ Instance = &DrainerSpec{}

// Ready implements Instance interface.
func (s *DrainerSpec) Ready(e executor.TiOpsExecutor) error {
	return nil
}

// ComponentName implements Instance interface.
func (s *DrainerSpec) ComponentName() string {
	return "drainer"
}

// ServiceName implements Instance interface.
func (s *DrainerSpec) ServiceName() string {
	return "drainer-" + s.IP + ".service"
}

// GetIP implements Instance interface.
func (s *DrainerSpec) GetIP() string {
	return s.IP
}

var _ Instance = &MonitorSpec{}

// Ready implements Instance interface.
func (s *MonitorSpec) Ready(e executor.TiOpsExecutor) error {
	return nil
}

// ComponentName implements Instance interface.
func (s *MonitorSpec) ComponentName() string {
	return "monitor"
}

// ServiceName implements Instance interface.
func (s *MonitorSpec) ServiceName() string {
	return "monitor-" + s.IP + ".service"
}

// GetIP implements Instance interface.
func (s *MonitorSpec) GetIP() string {
	return s.IP
}

var _ Instance = &GrafanaSpec{}

// Ready implements Instance interface.
func (s *GrafanaSpec) Ready(e executor.TiOpsExecutor) error {
	return nil
}

// ComponentName implements Instance interface.
func (s *GrafanaSpec) ComponentName() string {
	return "grafana"
}

// ServiceName implements Instance interface.
func (s *GrafanaSpec) ServiceName() string {
	return "grafana-" + s.IP + ".service"
}

// GetIP implements Instance interface.
func (s *GrafanaSpec) GetIP() string {
	return s.IP
}

var _ Instance = &AlertManagerSpec{}

// Ready implements Instance interface.
func (s *AlertManagerSpec) Ready(e executor.TiOpsExecutor) error {
	return nil
}

// ComponentName implements Instance interface.
func (s *AlertManagerSpec) ComponentName() string {
	return "alertmanager"
}

// ServiceName implements Instance interface.
func (s *AlertManagerSpec) ServiceName() string {
	return "alertmanager-" + s.IP + ".service"
}

// GetIP implements Instance interface.
func (s *AlertManagerSpec) GetIP() string {
	return s.IP
}

// TiDBComponent represents TiDB component.
type TiDBComponent []TiDBSpec

// Name implements Component interface.
func (c TiDBComponent) Name() string {
	return "tidb"
}

// Instances implements Component interface.
func (c TiDBComponent) Instances() (ins []Instance) {
	for _, s := range c {
		ins = append(ins, &s)
	}

	return
}

// TiKVComponent represents TiKV component.
type TiKVComponent []TiKVSpec

// Name implements Component interface.
func (c TiKVComponent) Name() string {
	return "tikv"
}

// Instances implements Component interface.
func (c TiKVComponent) Instances() (ins []Instance) {
	for _, s := range c {
		ins = append(ins, &s)
	}

	return
}

// PDComponent represents PD component.
type PDComponent []PDSpec

// Name implements Component interface.
func (c PDComponent) Name() string {
	return "pd"
}

// Instances implements Component interface.
func (c PDComponent) Instances() (ins []Instance) {
	for _, s := range c {
		ins = append(ins, &s)
	}

	return
}

// PumpComponent represents Pump component.
type PumpComponent []PumpSpec

// Name implements Component interface.
func (c PumpComponent) Name() string {
	return "pd"
}

// Instances implements Component interface.
func (c PumpComponent) Instances() (ins []Instance) {
	for _, s := range c {
		ins = append(ins, &s)
	}

	return
}

// DrainerComponent represents Drainer component.
type DrainerComponent []DrainerSpec

// Name implements Component interface.
func (c DrainerComponent) Name() string {
	return "drainer"
}

// Instances implements Component interface.
func (c DrainerComponent) Instances() (ins []Instance) {
	for _, s := range c {
		ins = append(ins, &s)
	}

	return
}

// MonitorComponent represents Monitor component.
type MonitorComponent []MonitorSpec

// Name implements Component interface.
func (c MonitorComponent) Name() string {
	return "monitor"
}

// Instances implements Component interface.
func (c MonitorComponent) Instances() (ins []Instance) {
	for _, s := range c {
		ins = append(ins, &s)
	}

	return
}

// GrafanaComponent represents Grafana component.
type GrafanaComponent []GrafanaSpec

// Name implements Component interface.
func (c GrafanaComponent) Name() string {
	return "grafana"
}

// Instances implements Component interface.
func (c GrafanaComponent) Instances() (ins []Instance) {
	for _, s := range c {
		ins = append(ins, &s)
	}

	return
}

// AlertmanagerComponent represents Alertmanager component.
type AlertmanagerComponent []AlertManagerSpec

// Name implements Component interface.
func (c AlertmanagerComponent) Name() string {
	return "alertmanager"
}

// Instances implements Component interface.
func (c AlertmanagerComponent) Instances() (ins []Instance) {
	for _, s := range c {
		ins = append(ins, &s)
	}

	return
}

// ComponentsByStartOrder return component in the order need to start.
func (s *Specification) ComponentsByStartOrder() (comps []Component) {
	// "pd", "tikv", "pump", "tidb", "drainer", "prometheus", "grafana", "alertmanager"

	comps = append(comps, PDComponent(s.PDServers))
	comps = append(comps, TiKVComponent(s.TiKVServers))
	comps = append(comps, PumpComponent(s.PumpServers))
	comps = append(comps, TiDBComponent(s.TiDBServers))
	comps = append(comps, DrainerComponent(s.Drainers))
	comps = append(comps, MonitorComponent(s.MonitorSpec))
	comps = append(comps, GrafanaComponent(s.Grafana))
	comps = append(comps, AlertmanagerComponent(s.Alertmanager))

	return
}

// Component represents a component of the cluster.
type Component interface {
	Name() string
	Instances() []Instance
}

// pd may need check this
// url="http://{{ ansible_host }}:{{ client_port }}/health"
// other just check pont is listen

// Instance represents the instance.
type Instance interface {
	Ready(executor.TiOpsExecutor) error
	ComponentName() string
	ServiceName() string
	GetIP() string
}
