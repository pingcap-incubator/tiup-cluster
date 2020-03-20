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
	"github.com/pingcap-incubator/tiops/pkg/task"
	"github.com/spf13/cobra"
)

func newDeploy() *cobra.Command {
	// for test
	var (
		host       string // hostname of the SSH server
		port       int    // port of the SSH server
		user       string // username to login to the SSH server
		password   string // password of the user
		keyFile    string // path to the private key file
		passphrase string // passphrase of the private key file
	)
	cmd := &cobra.Command{
		Use:          "deploy",
		Short:        "Deploy a cluster for production",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			t := task.NewBuilder().
				RootSSH(host, port, user, password, keyFile, passphrase).
				SSHKeyGen("ssh/tiops/id_rsa").
				EnvInit(host).
				// Switch the SSH tunnel to the `tidb` user
				UserSSH(host).
				Mkdir(host, "~/deploy/tidb/bin", "~/deploy/tidb/logs", "~/deploy/tidb/data").
				CopyComponent("tidb", "v3.0.10", host, "~/deploy/tidb/bin/").
				Build()
			return t.Execute(task.NewContext())
		},
	}

	cmd.Flags().StringVar(&host, "host", "", "deploy to host")
	cmd.Flags().IntVar(&port, "port", 22, "deploy to host")
	cmd.Flags().StringVar(&user, "user", "root", "system user root")
	cmd.Flags().StringVar(&password, "password", "", "system user root")
	cmd.Flags().StringVar(&keyFile, "key", "", "keypath")
	cmd.Flags().StringVar(&passphrase, "passphrase", "", "passphrase")
	return cmd
}
