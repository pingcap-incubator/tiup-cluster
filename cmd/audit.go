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
	"bufio"
	"io/ioutil"
	"os"
	"time"

	"github.com/pingcap-incubator/tiops/pkg/base52"
	"github.com/pingcap-incubator/tiops/pkg/meta"
	"github.com/pingcap-incubator/tiops/pkg/utils"
	tiuputils "github.com/pingcap-incubator/tiup/pkg/utils"
	"github.com/pingcap/errors"
	"github.com/spf13/cobra"
)

func newAuditCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "audit [audit-id]",
		Short: "Show audit log of cluster operation",
		RunE: func(cmd *cobra.Command, args []string) error {
			switch len(args) {
			case 0:
				return showAuditList()
			case 1:
				return showAuditLog(args[0])
			default:
				return cmd.Help()
			}
		},
	}
	return cmd
}

func showAuditList() error {
	firstLine := func(fileName string) (string, error) {
		file, err := os.Open(meta.ProfilePath(meta.TiOpsAuditDir, fileName))
		if err != nil {
			return "", errors.Trace(err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		if scanner.Scan() {
			return scanner.Text(), nil
		}
		return "", errors.New("unknown audit log format")
	}

	auditDir := meta.ProfilePath(meta.TiOpsAuditDir)
	// Header
	clusterTable := [][]string{{"ID", "Time", "Command"}}
	fileInfos, err := ioutil.ReadDir(auditDir)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	for _, fi := range fileInfos {
		if fi.IsDir() {
			continue
		}
		ts, err := base52.Decode(fi.Name())
		if err != nil {
			continue
		}
		t := time.Unix(ts, 0)
		cmd, err := firstLine(fi.Name())
		if err != nil {
			continue
		}
		clusterTable = append(clusterTable, []string{
			fi.Name(),
			t.Format(time.RFC3339),
			cmd,
		})
	}

	utils.PrintTable(clusterTable, true)
	return nil
}

func showAuditLog(auditID string) error {
	path := meta.ProfilePath(meta.TiOpsAuditDir, auditID)
	if tiuputils.IsNotExist(path) {
		return errors.Errorf("cannot find the audit log '%s'", auditID)
	}

	content, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.Trace(err)
	}
	_, _ = os.Stdout.Write(content)
	return nil
}
