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

package task

import (
	"fmt"

	"github.com/pingcap-incubator/tiup-cluster/pkg/operation"
)

// the check types
var (
	CheckTypeSystemInfo   = "insight"
	CheckTypeSystemLimits = "limits"
	CheckTypeKernelParam  = "sysctl"
	CheckTypeService      = "service"
)

// CheckSys performs checks of system information
type CheckSys struct {
	host  string
	user  string
	opt   *operator.CheckOptions
	check string
}

// Execute implements the Task interface
func (c *CheckSys) Execute(ctx *Context) error {
	stdout, stderr, _ := ctx.GetOutputs(c.host)
	if len(stderr) > 0 && len(stdout) == 0 {
		return fmt.Errorf("error getting output of %s: %s", c.host, stderr)
	}

	switch c.check {
	case CheckTypeSystemInfo:
		ctx.SetCheckResults(c.host, operator.CheckSystemInfo(c.opt, stdout))
	case CheckTypeSystemLimits:
		ctx.SetCheckResults(c.host, operator.CheckSysLimits(c.opt, c.user, stdout))
	case CheckTypeKernelParam:
		ctx.SetCheckResults(c.host, operator.CheckKernelParameters(c.opt, stdout))
	case CheckTypeService:
		e, ok := ctx.GetExecutor(c.host)
		if !ok {
			return fmt.Errorf("can not get executor for %s", c.host)
		}
		var results []*operator.CheckResult
		results = append(
			results,
			operator.CheckServices(e, c.host, "irqbalance", false),
			// FIXME: set firewalld rules in deploy, and not disabling it anymore
			operator.CheckServices(e, c.host, "firewalld", true),
		)
		ctx.SetCheckResults(c.host, results)
	}

	return nil
}

// Rollback implements the Task interface
func (c *CheckSys) Rollback(ctx *Context) error {
	return ErrUnsupportedRollback
}

// String implements the fmt.Stringer interface
func (c *CheckSys) String() string {
	return fmt.Sprintf("CheckSys: host=%s type=%s", c.host, c.check)
}
