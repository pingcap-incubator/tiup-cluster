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

package ansible

import (
	"fmt"
	"os"
	"strings"

	"github.com/pingcap-incubator/tiops/pkg/executor"
	"github.com/pingcap-incubator/tiops/pkg/meta"
	"github.com/relex/aini"
)

var (
	systemdUnitPath = "/etc/systemd/system"
)

// parseDirs sets values of directories of component
func parseDirs(host *aini.Host, ins meta.InstanceSpec) (meta.InstanceSpec, error) {
	hostName, sshPort := ins.SSH()
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ins, err
	}
	e, err := executor.NewSSHExecutor(executor.SSHConfig{
		Host:    hostName,
		Port:    sshPort,
		User:    host.Vars["ansible_user"],
		KeyFile: fmt.Sprintf("%s/.ssh/id_rsa", homeDir), // ansible generated keyfile
	})
	if err != nil {
		return ins, err
	}

	switch ins.Role() {
	case meta.RoleTiDB:
		serviceFile := fmt.Sprintf("%s/%s-%d.service",
			systemdUnitPath,
			meta.ComponentTiDB,
			ins.GetMainPort())
		cmd := fmt.Sprintf("cat `grep 'ExecStart' %s | sed 's/ExecStart=//'`", serviceFile)
		stdout, _, err := e.Execute(cmd, false)
		if err != nil {
			return ins, nil
		}

		// parse dirs
		newIns := ins.(meta.TiDBSpec)
		for _, line := range strings.Split(string(stdout), "\n") {
			if strings.HasPrefix(line, "DEPLOY_DIR=") {
				newIns.DeployDir = strings.TrimPrefix(line, "DEPLOY_DIR=")
				continue
			}
			if strings.Contains(line, "--log-file=") {
				fullLog := strings.Split(line, " ")[4] // 4 whitespaces ahead
				logDir := strings.TrimSuffix(strings.TrimPrefix(fullLog,
					"--log-file=\""), "/tidb.log\"")
				newIns.LogDir = logDir
				continue
			}
		}
		return newIns, nil
	case meta.RoleTiKV:
		serviceFile := fmt.Sprintf("%s/%s-%d.service",
			systemdUnitPath,
			meta.ComponentTiKV,
			ins.GetMainPort())
		cmd := fmt.Sprintf("cat `grep 'ExecStart' %s | sed 's/ExecStart=//'`", serviceFile)
		stdout, _, err := e.Execute(cmd, false)
		if err != nil {
			return ins, nil
		}

		// parse dirs
		newIns := ins.(meta.TiKVSpec)
		for _, line := range strings.Split(string(stdout), "\n") {
			if strings.HasPrefix(line, "cd \"") {
				newIns.DeployDir = strings.Trim(strings.Split(line, " ")[1], "\"")
				continue
			}
			if strings.Contains(line, "--data-dir") {
				dataDir := strings.Split(line, " ")[5] // 4 whitespaces ahead
				newIns.DataDir = strings.Trim(dataDir, "\"")
				continue
			}
			if strings.Contains(line, "--log-file") {
				fullLog := strings.Split(line, " ")[5] // 4 whitespaces ahead
				logDir := strings.TrimSuffix(strings.TrimPrefix(fullLog,
					"\""), "/tikv.log\"")
				newIns.LogDir = logDir
				continue
			}
		}
		return newIns, nil
	case meta.RolePD:
		serviceFile := fmt.Sprintf("%s/%s-%d.service",
			systemdUnitPath,
			meta.ComponentPD,
			ins.GetMainPort())
		cmd := fmt.Sprintf("cat `grep 'ExecStart' %s | sed 's/ExecStart=//'`", serviceFile)
		stdout, _, err := e.Execute(cmd, false)
		if err != nil {
			return ins, nil
		}
		//fmt.Printf("%s\n", stdout)

		// parse dirs
		newIns := ins.(meta.PDSpec)
		for _, line := range strings.Split(string(stdout), "\n") {
			if strings.HasPrefix(line, "DEPLOY_DIR=") {
				newIns.DeployDir = strings.TrimPrefix(line, "DEPLOY_DIR=")
				continue
			}
			if strings.Contains(line, "--name") {
				nameArg := strings.Split(line, " ")[4] // 4 whitespaces ahead
				name := strings.TrimPrefix(nameArg, "--name=")
				newIns.Name = strings.Trim(name, "\"")
				continue
			}
			if strings.Contains(line, "--data-dir") {
				dataArg := strings.Split(line, " ")[4] // 4 whitespaces ahead
				dataDir := strings.TrimPrefix(dataArg, "--data-dir=")
				newIns.DataDir = strings.Trim(dataDir, "\"")
				continue
			}
			if strings.Contains(line, "--log-file=") {
				fullLog := strings.Split(line, " ")[4] // 4 whitespaces ahead
				logDir := strings.TrimSuffix(strings.TrimPrefix(fullLog,
					"--log-file=\""), "/pd.log\"")
				newIns.LogDir = logDir
				continue
			}
		}
		return newIns, nil
	case meta.RolePump:
	case meta.RoleDrainer:
	case meta.RoleMonitor:
	case meta.RoleGrafana:
	}
	return ins, nil
}
