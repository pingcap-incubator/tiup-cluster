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
	"path"

	"github.com/fatih/color"
	"github.com/pingcap-incubator/tiops/pkg/executor"
	"github.com/pingcap-incubator/tiops/pkg/flags"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/pingcap-incubator/tiops/pkg/meta"
	"github.com/pingcap-incubator/tiops/pkg/version"
	tiupmeta "github.com/pingcap-incubator/tiup/pkg/meta"
	"github.com/pingcap-incubator/tiup/pkg/repository"
	"github.com/spf13/cobra"
)

var rootCmd *cobra.Command

func init() {
	cobra.EnableCommandSorting = false

	rootCmd = &cobra.Command{
		Use:           "tiops",
		Short:         "Deploy a TiDB cluster for production",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version.NewTiOpsVersion().FullInfo(),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := meta.Initialize(); err != nil {
				return err
			}
			// Must after meta.Initialize for setting profile directory.
			initExecuteLog()
			return tiupmeta.InitRepository(repository.Options{
				GOOS:   "linux",
				GOARCH: "amd64",
			})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			return tiupmeta.Repository().Mirror().Close()
		},
	}

	rootCmd.AddCommand(
		newDeploy(),
		newStartCmd(),
		newStopCmd(),
		newRestartCmd(),
		newScaleInCmd(),
		newScaleOutCmd(),
		newDestroyCmd(),
		newUpgradeCmd(),
		newReloadCmd(),
		newExecCmd(),
		newDisplayCmd(),
		newListCmd(),
		newImportCmd(),
	)
}

// Execute executes the root command
func Execute() {
	flags.ShowBacktrace = len(os.Getenv("TIUP_BACKTRACE")) > 0
	if err := rootCmd.Execute(); err != nil {
		if flags.ShowBacktrace {
			fmt.Println(color.RedString("Error: %+v", err))
		} else {
			fmt.Println(color.RedString("Error: %v", err))
		}
		os.Exit(1)
	}
}

func initExecuteLog() {
	fname := path.Join(meta.ProfilePath(meta.TiOpsLogDir), "execute.log")
	fmt.Println(fname)
	var w = &lumberjack.Logger{
		Filename:   fname,
		MaxSize:    10, // megabytes
		MaxBackups: 3,
		MaxAge:     28, //days
	}

	executor.SetLogger(w)
}
