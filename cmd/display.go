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
	"reflect"
	"strings"

	//"github.com/pingcap-incubator/tiops/pkg/task"
	"github.com/pingcap-incubator/tiops/pkg/meta"
	"github.com/pingcap-incubator/tiops/pkg/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func newDisplayCmd() *cobra.Command {
	var (
		clusterName  string
		topologyFile string
	)

	cmd := &cobra.Command{
		Use:    "display <cluster> [OPTIONS]",
		Short:  "Display information of a TiDB cluster",
		Hidden: true,
		Args: func(cmd *cobra.Command, args []string) error {
			switch len(args) {
			case 0:
				cmd.Help()
				return fmt.Errorf("cluster name not specified")
			case 1:
				fallthrough
			default:
				if strings.HasPrefix(args[0], "-") {
					cmd.Help()
					return fmt.Errorf("cluster name not specified")
				}
				clusterName = args[0]
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return displayClusterTopology(clusterName, topologyFile)
		},
	}

	cmd.Flags().StringVarP(&topologyFile, "topology", "T", "", "path to the topology file")

	return cmd
}

func displayClusterTopology(name, topoFile string) error {
	/*
		clsMeta, err := meta.ClusterMetadata(name)
		if err != nil {
			return err
		}
		clsTopo, err := meta.ClusterTopology(name)
		if err != nil {
			return err
		}
	*/

	//var clsMeta meta.ClusterMeta
	var clsTopo meta.TopologySpecification

	yamlFile, err := ioutil.ReadFile(topoFile)
	if err != nil {
		return err
	}
	if err = yaml.Unmarshal(yamlFile, &clsTopo); err != nil {
		return err
	}

	var clusterTable [][]string
	clusterTable = append(clusterTable,
		[]string{"ID", "Role", "Host", "Ports", "Data Dir", "Deploy Dir"})

	v := reflect.ValueOf(clsTopo)
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		subTable, err := buildTable(v.Field(i))
		if err != nil {
			continue
		}
		clusterTable = append(clusterTable, subTable...)
	}
	utils.PrintTable(clusterTable, true)

	return nil
}

func buildTable(field reflect.Value) ([][]string, error) {
	var resTable [][]string

	switch field.Kind() {
	case reflect.Slice:
		for i := 0; i < field.Len(); i++ {
			subTable, err := buildTable(field.Index(i))
			if err != nil {
				return nil, err
			}
			resTable = append(resTable, subTable...)
		}
	case reflect.Ptr:
		subTable, err := buildTable(field.Elem())
		if err != nil {
			return nil, err
		}
		resTable = append(resTable, subTable...)
	case reflect.Struct:
		ins := field.Interface().(meta.InstanceSpec)

		dataDir := "-"
		insDirs := ins.GetDir()
		deployDir := insDirs[0]
		if len(insDirs) > 1 {
			dataDir = insDirs[1]
		}

		resTable = append(resTable, []string{
			ins.GetID(),
			ins.Role(),
			ins.GetHost(),
			utils.JoinInt(ins.GetPort(), "/"),
			dataDir,
			deployDir,
		})
	}

	return resTable, nil
}
