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
	"io"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/pingcap-incubator/tiops/pkg/flags"
	"github.com/pingcap-incubator/tiops/pkg/log"
	"github.com/pingcap-incubator/tiops/pkg/meta"
	"github.com/pingcap-incubator/tiops/pkg/utils"
	"github.com/pingcap-incubator/tiops/pkg/version"
	tiupmeta "github.com/pingcap-incubator/tiup/pkg/meta"
	"github.com/pingcap-incubator/tiup/pkg/repository"
	"github.com/pingcap/errors"
	"github.com/spf13/cobra"
)

var rootCmd *cobra.Command
var auditFile *os.File

func init() {
	flags.ShowBacktrace = len(os.Getenv("TIUP_BACKTRACE")) > 0
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
			auditDir := meta.ProfilePath(meta.TiOpsAuditDir)
			if err := utils.CreateDir(auditDir); err != nil {
				return errors.Trace(err)
			}
			auditFilePath := meta.ProfilePath(meta.TiOpsAuditDir, time.Now().Format(time.RFC3339))
			f, err := os.OpenFile(auditFilePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
			if err != nil {
				return errors.Trace(err)
			}
			_, _ = auditFile.WriteString(strings.Join(os.Args, " ") + "\n")
			auditFile = f
			log.SetOutput(io.MultiWriter(os.Stdout, auditFile))
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
	var code int
	err := rootCmd.Execute()
	if err != nil {
		if flags.ShowBacktrace {
			log.Output(color.RedString("Error: %+v", err))
		} else {
			log.Output(color.RedString("Error: %v", err))
		}
		code = 1
	}
	if auditFile != nil {
		_ = auditFile.Close()
	}
	os.Exit(code)
}
