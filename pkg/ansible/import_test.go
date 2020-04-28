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
	"bytes"
	"testing"

	. "github.com/pingcap/check"
	"gopkg.in/yaml.v2"
)

type ansSuite struct {
}

var _ = Suite(&ansSuite{})

func TestAnsible(t *testing.T) {
	TestingT(t)
}

func (s *ansSuite) TestParseInventoryFile(c *C) {
	invData := bytes.NewReader([]byte(`## TiDB Cluster Part
[tidb_servers]
tidb218 ansible_host=172.16.5.218 tidb_status_port=3399
172.16.5.219 tidb_port=3397

[tikv_servers]
172.16.5.219
172.16.5.220 tikv_port=20166
172.16.5.221

[pd_servers]
172.16.5.218
172.16.5.219 pd_status_port=2381
172.16.5.220 deploy_dir=/data-path/custom_deploy/pd220

[spark_master]

[spark_slaves]

[lightning_server]

[importer_server]

## Monitoring Part
# prometheus and pushgateway servers
[monitoring_servers]
172.16.5.221

[grafana_servers]
172.16.5.221

# node_exporter and blackbox_exporter servers
[monitored_servers]
172.16.5.218
172.16.5.219
172.16.5.220
172.16.5.221

[alertmanager_servers]
172.16.5.221

[kafka_exporter_servers]

## Binlog Part
[pump_servers]

[drainer_servers]

## Group variables
[pd_servers:vars]
# location_labels = ["zone","rack","host"]

## Global variables
[all:vars]
deploy_dir = /home/tiops/ansible-deploy

## Connection
# ssh via normal user
ansible_user = tiops

cluster_name = ansible-cluster

tidb_version = v3.0.12

# process supervision, [systemd, supervise]
process_supervision = systemd

timezone = Asia/Shanghai

enable_firewalld = False
# check NTP service
enable_ntpd = True
set_hostname = True

## binlog trigger
enable_binlog = False

# kafka cluster address for monitoring, example:
# kafka_addrs = "192.168.0.11:9092,192.168.0.12:9092,192.168.0.13:9092"
kafka_addrs = ""

# zookeeper address of kafka cluster for monitoring, example:
# zookeeper_addrs = "192.168.0.11:2181,192.168.0.12:2181,192.168.0.13:2181"
zookeeper_addrs = ""

# enable TLS authentication in the TiDB cluster
enable_tls = False

# KV mode
deploy_without_tidb = False

# wait for region replication complete before start tidb-server.
wait_replication = True

# Optional: Set if you already have a alertmanager server.
# Format: alertmanager_host:alertmanager_port
alertmanager_target = ""

grafana_admin_user = "admin"
grafana_admin_password = "admin"


### Collect diagnosis
collect_log_recent_hours = 2

enable_bandwidth_limit = False
# default: 10Mb/s, unit: Kbit/s
collect_bandwidth_limit = 10000`))
	clsName, clsMeta, inv, err := parseInventoryFile(invData)
	c.Assert(err, IsNil)
	c.Assert(inv, NotNil)
	c.Assert(clsName, Equals, "ansible-cluster")
	c.Assert(clsMeta, NotNil)
	c.Assert(clsMeta.Version, Equals, "v3.0.12")
	c.Assert(clsMeta.User, Equals, "tiops")

	expected := []byte(`global:
  user: tiops
  resource_control:
    memory_limit: ""
    cpu_quota: ""
    io_read_bandwidth_max: ""
    io_write_bandwidth_max: ""
tidb_servers: []
tikv_servers: []
tiflash_servers: []
pd_servers: []
monitoring_servers: []
`)

	topo, err := yaml.Marshal(clsMeta.Topology)
	c.Assert(err, IsNil)
	c.Assert(topo, DeepEquals, expected)
}
