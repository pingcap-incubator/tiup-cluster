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
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/joomcode/errorx"
	"github.com/pingcap-incubator/tiup-cluster/pkg/cliutil"
	"github.com/pingcap-incubator/tiup-cluster/pkg/colorutil"
	"github.com/pingcap-incubator/tiup-cluster/pkg/errutil"
	"github.com/pingcap-incubator/tiup-cluster/pkg/flags"
	"github.com/pingcap-incubator/tiup-cluster/pkg/logger"
	"github.com/pingcap-incubator/tiup-cluster/pkg/meta"
	"github.com/pingcap-incubator/tiup-cluster/pkg/version"
	"github.com/pingcap-incubator/tiup/pkg/localdata"
	tiupmeta "github.com/pingcap-incubator/tiup/pkg/meta"
	"github.com/pingcap-incubator/tiup/pkg/repository"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var rootCmd *cobra.Command

var (
	errNS      = errorx.NewNamespace("cmd")
	sshTimeout int64 // timeout in seconds when connecting an SSH server
)

func init() {
	logger.InitGlobalLogger()

	colorutil.AddColorFunctionsForCobra()

	// Initialize the global variables
	flags.ShowBacktrace = len(os.Getenv("TIUP_BACKTRACE")) > 0
	cobra.EnableCommandSorting = false

	rootCmd = &cobra.Command{
		Use:           filepath.Base(os.Args[0]),
		Short:         "Deploy a TiDB cluster for production",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version.NewTiOpsVersion().FullInfo(),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := meta.Initialize(); err != nil {
				return err
			}
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

	beautifyCobraUsageAndHelp(rootCmd)

	rootCmd.PersistentFlags().Int64Var(&sshTimeout, "ssh-timeout", 5, "Timeout in seconds to connect host via SSH, ignored for operations that don't need an SSH connection.")

	rootCmd.AddCommand(
		newDeploy(),
		newStartCmd(),
		newStopCmd(),
		newRestartCmd(),
		newScaleInCmd(),
		newScaleOutCmd(),
		newDestroyCmd(),
		newUpgradeCmd(),
		newExecCmd(),
		newDisplayCmd(),
		newListCmd(),
		newAuditCmd(),
		newImportCmd(),
		newEditConfigCmd(),
		newReloadCmd(),
		newPatchCmd(),
	)
}

func printErrorMessageForNormalError(err error) {
	_, _ = colorutil.ColorErrorMsg.Fprintf(os.Stderr, "\nError: %s\n", err.Error())
}

func printErrorMessageForErrorX(err *errorx.Error) {
	msg := ""
	ident := 0
	causeErrX := err
	for causeErrX != nil {
		if ident > 0 {
			msg += strings.Repeat("  ", ident) + "caused by: "
		}
		currentErrMsg := causeErrX.Message()
		if len(currentErrMsg) > 0 {
			if ident == 0 {
				// Print error code only for top level error
				msg += fmt.Sprintf("%s (%s)\n", currentErrMsg, causeErrX.Type().FullName())
			} else {
				msg += fmt.Sprintf("%s\n", currentErrMsg)
			}
			ident++
		}
		cause := causeErrX.Cause()
		if c := errorx.Cast(cause); c != nil {
			causeErrX = c
		} else if cause != nil {
			if ident > 0 {
				// Out most error may have empty message. In this case we treat it as a transparent error.
				// Thus `ident == 0` can be possible.
				msg += strings.Repeat("  ", ident) + "caused by: "
			}
			msg += fmt.Sprintf("%s\n", cause.Error())
			break
		} else {
			break
		}
	}
	_, _ = colorutil.ColorErrorMsg.Fprintf(os.Stderr, "\nError: %s", msg)
}

func extractSuggestionFromErrorX(err *errorx.Error) string {
	cause := err
	for cause != nil {
		v, ok := cause.Property(errutil.ErrPropSuggestion)
		if ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
		cause = errorx.Cast(cause.Cause())
	}

	return ""
}

// Execute executes the root command
func Execute() {
	zap.L().Info("Execute command", zap.String("command", cliutil.OsArgs()))
	zap.L().Debug("Environment variables", zap.Strings("env", os.Environ()))

	// Switch current work directory if running in TiUP component mode
	if wd := os.Getenv(localdata.EnvNameWorkDir); wd != "" {
		if err := os.Chdir(wd); err != nil {
			zap.L().Warn("Failed to switch work directory", zap.String("working_dir", wd), zap.Error(err))
		}
	}

	code := 0
	err := rootCmd.Execute()
	if err != nil {
		code = 1
	}

	zap.L().Info("Execute command finished", zap.Int("code", code), zap.Error(err))

	if err != nil {
		if errx := errorx.Cast(err); errx != nil {
			printErrorMessageForErrorX(errx)
		} else {
			printErrorMessageForNormalError(err)
		}

		if !errorx.HasTrait(err, errutil.ErrTraitPreCheck) {
			logger.OutputDebugLog()
		}

		if errx := errorx.Cast(err); errx != nil {
			if suggestion := extractSuggestionFromErrorX(errx); len(suggestion) > 0 {
				_, _ = fmt.Fprintf(os.Stderr, "\n%s\n", suggestion)
			}
		}
	}

	logger.OutputAuditLogIfEnabled()

	color.Unset()

	os.Exit(code)
}

func beautifyCobraUsageAndHelp(rootCmd *cobra.Command) {
	rootCmd.SetUsageTemplate(`Usage:{{if .Runnable}}
  {{ColorCommand}}{{.UseLine}}{{ColorReset}}{{end}}{{if .HasAvailableSubCommands}}
  {{ColorCommand}}{{.CommandPath}} [command]{{ColorReset}}{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{ColorCommand}}{{.NameAndAliases}}{{ColorReset}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{ColorCommand}}{{.CommandPath}} [command] --help{{ColorReset}}" for more information about a command.{{end}}
`)
}
