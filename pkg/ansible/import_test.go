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
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/creasty/defaults"
	"github.com/pingcap-incubator/tiup-cluster/pkg/meta"
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
	dir := "test-data"
	invData, err := os.Open(filepath.Join(dir, "inventory.ini"))
	c.Assert(err, IsNil)

	clsName, clsMeta, inv, err := parseInventoryFile(invData)
	c.Assert(err, IsNil)
	c.Assert(inv, NotNil)
	c.Assert(clsName, Equals, "ansible-cluster")
	c.Assert(clsMeta, NotNil)
	c.Assert(clsMeta.Version, Equals, "v3.0.12")
	c.Assert(clsMeta.User, Equals, "tiops")

	expected := []byte(`global:
  user: tiops
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

func (s *ansSuite) TestParseGroupVars(c *C) {
	dir := "test-data"
	invData, err := os.Open(filepath.Join(dir, "inventory.ini"))
	c.Assert(err, IsNil)
	_, clsMeta, inv, err := parseInventoryFile(invData)
	c.Assert(err, IsNil)

	err = parseGroupVars(dir, clsMeta, inv)
	c.Assert(err, IsNil)
	err = defaults.Set(clsMeta)
	c.Assert(err, IsNil)

	var expected meta.ClusterMeta
	var metaFull meta.ClusterMeta

	expectedTopo, err := ioutil.ReadFile(filepath.Join(dir, "meta.yaml"))
	c.Assert(err, IsNil)
	err = yaml.Unmarshal(expectedTopo, &expected)
	c.Assert(err, IsNil)

	// marshal and unmarshal the meta to ensure custom defaults are populated
	meta, err := yaml.Marshal(clsMeta)
	c.Assert(err, IsNil)
	err = yaml.Unmarshal(meta, &metaFull)
	c.Assert(err, IsNil)

	sortClusterMeta(&metaFull)
	sortClusterMeta(&expected)

	c.Assert(metaFull, DeepEquals, expected)
}

func sortClusterMeta(clsMeta *meta.ClusterMeta) {
	v := reflect.ValueOf(clsMeta.Topology).Elem()

	for i := 0; i < v.Type().NumField(); i++ {
		switch v.Field(i).Kind() {
		case reflect.Slice:
			field := v.Field(i)
			switch v.Type().Field(i).Name {
			case "TiDBServers":
				lst := field.Interface().([]meta.TiDBSpec)
				sort.Slice(lst, func(i, j int) bool {
					return lst[i].Host < lst[j].Host
				})
				v.Field(i).Set(reflect.ValueOf(lst))
			case "TiKVServers":
				lst := field.Interface().([]meta.TiKVSpec)
				sort.Slice(lst, func(i, j int) bool {
					return lst[i].Host < lst[j].Host
				})
				v.Field(i).Set(reflect.ValueOf(lst))
			case "PDServers":
				lst := field.Interface().([]meta.PDSpec)
				sort.Slice(lst, func(i, j int) bool {
					return lst[i].Host < lst[j].Host
				})
				v.Field(i).Set(reflect.ValueOf(lst))
			case "TiFlashServers":
				lst := field.Interface().([]meta.TiFlashSpec)
				sort.Slice(lst, func(i, j int) bool {
					return lst[i].Host < lst[j].Host
				})
				v.Field(i).Set(reflect.ValueOf(lst))
			case "PumpServers":
				lst := field.Interface().([]meta.PumpSpec)
				sort.Slice(lst, func(i, j int) bool {
					return lst[i].Host < lst[j].Host
				})
				v.Field(i).Set(reflect.ValueOf(lst))
			case "Drainers":
				lst := field.Interface().([]meta.DrainerSpec)
				sort.Slice(lst, func(i, j int) bool {
					return lst[i].Host < lst[j].Host
				})
				v.Field(i).Set(reflect.ValueOf(lst))
			case "Monitors":
				lst := field.Interface().([]meta.PrometheusSpec)
				sort.Slice(lst, func(i, j int) bool {
					return lst[i].Host < lst[j].Host
				})
				v.Field(i).Set(reflect.ValueOf(lst))
			case "Alertmanager":
				lst := field.Interface().([]meta.AlertManagerSpec)
				sort.Slice(lst, func(i, j int) bool {
					return lst[i].Host < lst[j].Host
				})
				v.Field(i).Set(reflect.ValueOf(lst))
			case "Grafana":
				lst := field.Interface().([]meta.GrafanaSpec)
				sort.Slice(lst, func(i, j int) bool {
					return lst[i].Host < lst[j].Host
				})
				v.Field(i).Set(reflect.ValueOf(lst))
			}
		}
	}
}
