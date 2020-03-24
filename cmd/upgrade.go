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
	"errors"
	"fmt"

	"github.com/pingcap-incubator/tiops/pkg/meta"
	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
)

type upgradeOptions struct {
	version string
	force   bool
	node    string
	role    string
}

func newUpgradeCmd() *cobra.Command {
	opt := upgradeOptions{}
	cmd := &cobra.Command{
		Use:   "upgrade <cluster-name>",
		Short: "Upgrade a specified TiDB cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return cmd.Help()
			}
			metaInfo, err := meta.ClusterMetadata(args[0])
			if err != nil {
				return err
			}
			curVersion := metaInfo.Version
			newVersion := opt.version
			if err := versionCompare(curVersion, newVersion); err != nil {
				return err
			}
			return upgrade() // TODO
		},
	}

	cmd.Flags().StringVar(&opt.version, "version", "", "version of cluster")
	cmd.Flags().BoolVar(&opt.force, "force", false, "upgrade without transfer leader, fast but affects stability (default false)")
	cmd.Flags().StringVar(&opt.node, "node-id", "", "node id")
	cmd.Flags().StringVar(&opt.role, "role", "", "role name")
	return cmd
}

func versionCompare(curVersion, newVersion string) error {
	if curVersion == "nightly" && newVersion == "nightly" { // imperfect
		return nil
	}
	if semver.Compare(curVersion, newVersion) == -1 {
		return nil
	} else {
		return errors.New(fmt.Sprintf("unsupport upgrade from %s to %s", curVersion, newVersion))
	}
}

// TODO
func upgrade() error {
	return nil
}
