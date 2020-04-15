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
	"os"

	"github.com/fatih/color"
	"github.com/pingcap-incubator/tiup-cluster/pkg/ansible"
	"github.com/pingcap-incubator/tiup-cluster/pkg/cliutil"
	"github.com/pingcap-incubator/tiup-cluster/pkg/log"
	"github.com/pingcap-incubator/tiup-cluster/pkg/meta"
	"github.com/pingcap-incubator/tiup-cluster/pkg/utils"
	tiuputils "github.com/pingcap-incubator/tiup/pkg/utils"
	"github.com/spf13/cobra"
)

func newImportCmd() *cobra.Command {
	var (
		ansibleDir        string
		inventoryFileName string
		rename            string
		noBackup          bool
	)

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import an exist TiDB cluster from TiDB-Ansible",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Use current directory as ansibleDir by default
			if ansibleDir == "" {
				cwd, err := os.Getwd()
				if err != nil {
					return err
				}
				ansibleDir = cwd
			}

			// migrate cluster metadata from Ansible inventory
			clsName, clsMeta, err := ansible.ReadInventory(ansibleDir, inventoryFileName, sshTimeout)
			if err != nil {
				return err
			}

			// Rename the imported cluster
			if rename != "" {
				clsName = rename
			}
			if clsName == "" {
				return fmt.Errorf("cluster name should not be empty")
			}
			if tiuputils.IsExist(meta.ClusterPath(clsName, meta.MetaFileName)) {
				return errDeployNameDuplicate.
					New("Cluster name '%s' is duplicated", clsName).
					WithProperty(cliutil.SuggestionFromFormat("Please use --rename `NAME` to specify another name (You can use `tiup cluster list` to see all clusters)"))
			}

			// prompt for backups
			backupDir := meta.ClusterPath(clsName, "ansible-backup")
			backupFile := fmt.Sprintf("tiup-%s.bak", inventoryFileName)
			prompt := fmt.Sprintf("The ansible directory will be moved to %s after import.", backupDir)
			if noBackup {
				log.Infof("The '--no-backup' flag is set, the ansible directory will be kept at its current location.")
				prompt = fmt.Sprintf(
					"The inventory file %s will be renamed to %s after import.",
					inventoryFileName,
					backupFile,
				)
			}
			log.Warnf("TiDB-Ansible and TiUP Cluster can NOT be used together, please DO NOT try to use ansible to manage the imported cluster anymore to avoid metadata conflict.")
			log.Infof(prompt)
			if err := cliutil.PromptForConfirmOrAbortError("Do you want to continue? [y/N]: "); err != nil {
				return err
			}

			// copy SSH key to TiOps profile directory
			if err = utils.CreateDir(meta.ClusterPath(clsName, "ssh")); err != nil {
				return err
			}
			srcKeyPathPriv := ansible.SSHKeyPath()
			srcKeyPathPub := srcKeyPathPriv + ".pub"
			dstKeyPathPriv := meta.ClusterPath(clsName, "ssh", "id_rsa")
			dstKeyPathPub := dstKeyPathPriv + ".pub"
			if err = utils.CopyFile(srcKeyPathPriv, dstKeyPathPriv); err != nil {
				return err
			}
			if err = utils.CopyFile(srcKeyPathPub, dstKeyPathPub); err != nil {
				return err
			}

			// copy config files form deployment servers
			if err = ansible.ImportConfig(clsName, clsMeta, sshTimeout); err != nil {
				return err
			}

			if err = meta.SaveClusterMeta(clsName, clsMeta); err != nil {
				return err
			}

			if noBackup {
				// rename original TiDB-Ansible inventory file
				if err = utils.Move(inventoryFileName, backupFile); err != nil {
					return err
				}
				log.Infof("Ansible inventory renamed to %s.", color.HiCyanString(backupFile))
			} else {
				// move original TiDB-Ansible directory to a staged location
				if err = utils.Move(ansibleDir, backupDir); err != nil {
					return err
				}
				log.Infof("Ansible inventory saved in %s.", color.HiCyanString(backupDir))
			}

			log.Infof("Cluster %s imported.", clsName)
			fmt.Printf("Try `%s` to see the cluster.\n",
				color.HiYellowString("%s display %s", cmd.Parent().Use, clsName))
			return nil
		},
	}

	cmd.Flags().StringVarP(&ansibleDir, "dir", "d", "", "The path to TiDB-Ansible directory")
	cmd.Flags().StringVar(&inventoryFileName, "inventory", ansible.AnsibleInventoryFile, "The name of inventory file")
	cmd.Flags().StringVarP(&rename, "rename", "r", "", "Rename the imported cluster to `NAME`")
	cmd.Flags().BoolVar(&noBackup, "no-backup", false, "Don't backup ansible dir, useful when there're multiple inventory files")

	return cmd
}
