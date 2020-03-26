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
	"path/filepath"
	"strings"

	"github.com/pingcap-incubator/tiops/pkg/meta"
	operator "github.com/pingcap-incubator/tiops/pkg/operation"
	"github.com/pingcap-incubator/tiops/pkg/task"
	"github.com/pingcap-incubator/tiup/pkg/set"
	"github.com/pingcap/errors"
	"github.com/spf13/cobra"
)

func newScaleInCmd() *cobra.Command {
	var nodes []string
	cmd := &cobra.Command{
		Use:   "scale-in <cluster-name>",
		Short: "Scale in a TiDB cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(nodes) < 1 || len(args) != 1 {
				return cmd.Help()
			}
			return scaleIn(args[0], nodes)
		},
	}
	cmd.Flags().StringSliceVar(&nodes, "node-id", nil, "Specify the node ids")
	return cmd
}

func scaleIn(cluster string, nodeIds []string) error {
	metadata, err := meta.ClusterMetadata(cluster)
	if err != nil {
		return err
	}

	// instances by uuid
	instances := map[string]meta.Instance{}

	// make sure all nodeIds exists in topology
	for _, component := range metadata.Topology.ComponentsByStartOrder() {
		for _, instance := range component.Instances() {
			instances[instance.UUID()] = instance
		}
	}

	// Clean components
	var cleanComponentTasks []task.Task
	for _, nodeID := range nodeIds {
		inst, found := instances[nodeID]
		if !found {
			return errors.Errorf("cannot find node id '%s' in topology", nodeID)
		}

		deployDir := inst.DeployDir()
		if !strings.HasPrefix(deployDir, "/") {
			deployDir = filepath.Join("/home/"+metadata.User+"/deploy", deployDir)
		}
		// Deploy component
		t := task.NewBuilder().Rmdir(inst.GetHost(), deployDir).Build()
		cleanComponentTasks = append(cleanComponentTasks, t)
	}

	// Regenerate configuration
	deleted := set.NewStringSet(nodeIds...)
	var regenConfigTasks []task.Task
	for _, component := range metadata.Topology.ComponentsByStartOrder() {
		for _, instance := range component.Instances() {
			if deleted.Exist(instance.GetHost()) {
				continue
			}
			deployDir := instance.DeployDir()
			if !strings.HasPrefix(deployDir, "/") {
				deployDir = filepath.Join("/home/"+metadata.User+"/deploy", deployDir)
			}
			t := task.NewBuilder().InitConfig(cluster, instance, deployDir).Build()
			regenConfigTasks = append(regenConfigTasks, t)
		}
	}

	t := task.NewBuilder().
		SSHKeySet(
			meta.ClusterPath(cluster, "ssh", "id_rsa"),
			meta.ClusterPath(cluster, "ssh", "id_rsa.pub")).
		ClusterSSH(metadata.Topology, metadata.User).
		ClusterOperate(metadata.Topology, operator.ScaleInOperation, operator.Options{
			DeletedNodes: nodeIds,
		}).
		UpdateMeta(cluster, metadata, nodeIds).
		Parallel(cleanComponentTasks...).
		Parallel(regenConfigTasks...).
		Build()

	return t.Execute(task.NewContext())
}
