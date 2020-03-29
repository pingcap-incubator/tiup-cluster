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
	"github.com/pingcap-incubator/tiops/pkg/ansible"
	"github.com/pingcap-incubator/tiops/pkg/meta"
	"github.com/spf13/cobra"
)

func newImportCmd() *cobra.Command {
	var (
		ansibleDir string
	)

	cmd := &cobra.Command{
		Use:    "import [OPTIONS]",
		Short:  "Import a TiDB cluster from tidb-ansible",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			clsName, clsMeta, err := ansible.ImportAnsible(ansibleDir)
			if err != nil {
				return err
			}
			if err := ansible.ImportConfig(clsName, clsMeta); err != nil {
				return err
			}

			return meta.SaveClusterMeta(clsName, clsMeta)
		},
	}

	cmd.Flags().StringVarP(&ansibleDir, "dir", "d", "", "The path to TiDB-Ansible's directory")

	return cmd
}
