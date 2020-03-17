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

package module

import (
	"fmt"
	"strings"

	"github.com/pingcap-incubator/tiops/pkg/executor"
)

// TiOpsModuleSystemdConfig is the configurations used to initialize a TiOpsModuleSystemd
type TiOpsModuleSystemdConfig struct {
	Unit         string // the name of systemd unit(s)
	Action       string // the action to perform with the unit
	Enabled      bool   // enable the unit or not
	ReloadDaemon bool   // Run daemon-reload before other actions

	// TODO: support more systemd functionalities
	//Scope string // user or system
	//Force bool // add the `--force` arg to systemctl command
}

// TiOpsModuleSystemd is the module used to control systemd units
type TiOpsModuleSystemd struct {
	cmd string // the built command
}

// NewTiOpsModuleSystemd builds and returns a TiOpsModuleSystemd object base on
// given config.
func NewTiOpsModuleSystemd(config TiOpsModuleSystemdConfig) *TiOpsModuleSystemd {
	systemctl := "/usr/bin/systemctl" // TODO: find binary in $PATH

	cmd := fmt.Sprintf("%s %s %s",
		systemctl, strings.ToLower(config.Action), config.Unit)

	if config.Enabled {
		cmd = fmt.Sprintf("%s && %s enable %s",
			cmd, systemctl, config.Unit)
	}

	if config.ReloadDaemon {
		cmd = fmt.Sprintf("%s daemon-reload && %s",
			systemctl, cmd)
	}

	return &TiOpsModuleSystemd{
		cmd: cmd,
	}
}

// Execute passes the command to executor and returns its results, the executor
// should be already initialized.
func (mod *TiOpsModuleSystemd) Execute(exec executor.TiOpsExecutor) ([]byte, []byte, error) {
	return exec.Execute(mod.cmd, true)
}
