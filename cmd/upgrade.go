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
	"golang.org/x/tools/go/ssa/interp/testdata/src/fmt"
	"os"
	"regexp"
	"strings"

	"github.com/pingcap-incubator/tiops/pkg/meta"
	"github.com/spf13/cobra"
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
			if len(os.Args) != 1 {
				return cmd.Help()
			}
			metaInfo, err := meta.ClusterMetadata(os.Args[1])
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
	curVersionNotRelease, _ := regexp.MatchString("[0-9]+.*-", curVersion)
	newVersionNotRelease, _ := regexp.MatchString("[0-9]+.*-", newVersion)
	curVersionRelease, _ := regexp.MatchString("[0-9]+", curVersion)
	newVersionRelease, _ := regexp.MatchString("[0-9]+", newVersion)

	if !newVersionNotRelease && !newVersionRelease {
		return nil
	} else if newVersionRelease {
		switch {
		case curVersionRelease && newVersion > curVersion:
			return nil
		case curVersionNotRelease:
			if newVersion > curVersion || strings.Contains(curVersion, newVersion) {
				return nil
			}
		}
	} else {
		switch {
		case curVersionRelease && (!strings.Contains(newVersion, curVersion) && newVersion > curVersion):
			return nil
		case curVersionNotRelease && newVersion > curVersion:
			return nil
		}
	}
	return errors.New(fmt.Sprintf("unsupport upgrade from %s to %s", curVersion, newVersion))
}

// TODO
func upgrade() error {
	return nil
}