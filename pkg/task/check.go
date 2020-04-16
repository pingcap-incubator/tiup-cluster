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

// CheckSys performs checks of system information
type CheckSys struct {
	host string
	opt  *operator.CheckOptions
}

// Execute implements the Task interface
func (c *CheckSys) Execute(ctx *Context) error {
	stdout, stderr, _ := ctx.GetOutputs(c.host)
	if len(stderr) > 0 {
		return fmt.Errorf("error getting system info of %s: %s", c.host, stderr)
	}
	if err := operator.CheckSystemInfo(c.opt, stdout); err != nil {
		return fmt.Errorf("check fails for %s: %s", c.host, err)
	}
	return nil
}

// Rollback implements the Task interface
func (c *CheckSys) Rollback(ctx *Context) error {
	return ErrUnsupportedRollback
}

// String implements the fmt.Stringer interface
func (c *CheckSys) String() string {
	return fmt.Sprintf("CheckSys: host=%s", c.host)
}
