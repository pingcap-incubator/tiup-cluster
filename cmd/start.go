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
	"github.com/pingcap-incubator/tiops/pkg/meta"
	operator "github.com/pingcap-incubator/tiops/pkg/operation"
	"github.com/pingcap-incubator/tiops/pkg/task"
	"github.com/spf13/cobra"
)

func newStartCmd() *cobra.Command {
	var (
		clusterName string
		options     operator.Options
	)

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start a TiDB cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			metadata, err := meta.ClusterMetadata(clusterName)
			if err != nil {
				return err
			}

			t := task.NewBuilder().
				SSHKeySet(
					meta.ClusterPath(clusterName, "ssh", "id_rsa"),
					meta.ClusterPath(clusterName, "ssh", "id_rsa.pub")).
				ClusterSSH(metadata.Topology, metadata.User).
				ClusterOperate(metadata.Topology, operator.StartOperation, options).
				Build()

			return t.Execute(task.NewContext())

		},
	}

	cmd.Flags().StringVar(&clusterName, "cluster", "", "cluster name")
	cmd.Flags().StringVar(&options.Role, "role", "", "role name")
	cmd.Flags().StringVar(&options.Node, "node-id", "", "node id")
	return cmd
}
