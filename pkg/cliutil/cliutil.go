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

package cliutil

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/joomcode/errorx"
	"github.com/spf13/cobra"

	"github.com/pingcap-incubator/tiops/pkg/colorutil"
	"github.com/pingcap-incubator/tiops/pkg/errutil"
)

var (
	errNS           = errorx.NewNamespace("cli")
	errMismatchArgs = errNS.NewType("mismatch_args", errutil.ErrTraitPreCheck)
)

var templateFuncs = template.FuncMap{
	"OsArgs":  osArgs,
	"OsArgs0": osArgs0,
}

func osArgs() string {
	return strings.Join(os.Args, " ")
}

func osArgs0() string {
	return os.Args[0]
}

func init() {
	colorutil.AddColorFunctions(func(name string, f interface{}) {
		templateFuncs[name] = f
	})
}

// CheckCommandArgsAndMayPrintHelp checks whether user passes enough number of arguments.
// If insufficient number of arguments are passed, an error with proper suggestion will be raised.
// When no argument is passed, command help will be printed and no error will be raised.
func CheckCommandArgsAndMayPrintHelp(cmd *cobra.Command, args []string, minArgs int) (shouldContinue bool, err error) {
	if minArgs == 0 {
		return true, nil
	}
	lenArgs := len(args)
	if lenArgs == 0 {
		return false, cmd.Help()
	}
	if lenArgs < minArgs {
		return false, errMismatchArgs.
			New("Expect at least %d arguments, but received %d arguments", minArgs, lenArgs).
			WithProperty(SuggestionFromString(cmd.UsageString()))
	}
	return true, nil
}

func formatSuggestion(templateStr string, data interface{}) string {
	t := template.Must(template.New("suggestion").Funcs(templateFuncs).Parse(templateStr))
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		panic(err)
	}
	return buf.String()
}

// SuggestionFromString creates a suggestion from string.
// Usage: SomeErrorX.WithProperty(SuggestionFromString(..))
func SuggestionFromString(str string) (errorx.Property, string) {
	return errutil.ErrPropSuggestion, strings.TrimSpace(str)
}

// SuggestionFromTemplate creates a suggestion from go template. Colorize function and some other utilities
// are available.
// Usage: SomeErrorX.WithProperty(SuggestionFromTemplate(..))
func SuggestionFromTemplate(templateStr string, data interface{}) (errorx.Property, string) {
	return SuggestionFromString(formatSuggestion(templateStr, data))
}

// SuggestionFromFormat creates a suggestion from a format.
// Usage: SomeErrorX.WithProperty(SuggestionFromFormat(..))
func SuggestionFromFormat(format string, a ...interface{}) (errorx.Property, string) {
	s := fmt.Sprintf(format, a...)
	return SuggestionFromString(s)
}
