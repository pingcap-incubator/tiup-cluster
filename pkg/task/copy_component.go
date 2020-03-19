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

	"github.com/pingcap-incubator/tiup/pkg/meta"
	"github.com/pingcap-incubator/tiup/pkg/repository"
	"github.com/pingcap/errors"
)

// CopyComponent is used to copy all files related the specific version a component
// to the target directory of path
type CopyComponent struct {
	component string
	version   repository.Version
	host      string
	dstPath   string
}

// Execute implements the Task interface
func (c *CopyComponent) Execute(ctx *Context) error {
	binPath, err := meta.BinaryPath(c.component, c.version)
	if err != nil {
		return err
	}

	exec, found := ctx.GetExecutor(c.host)
	if !found {
		return ErrNoExecutor
	}

	err = exec.Transfer(binPath, c.dstPath)
	if err != nil {
		return errors.Trace(err)
	}

	cmd := fmt.Sprintf(`chmod 755 %s`, c.dstPath)

	stdout, stderr, err := exec.Execute(cmd, false)
	if err != nil {
		return errors.Trace(err)
	}

	fmt.Println("Change binary permission stdout: ", string(stdout))
	fmt.Println("Change binary permission stderr: ", string(stderr))
	return nil
}

// Rollback implements the Task interface
func (c *CopyComponent) Rollback(ctx *Context) error {
	return ErrUnsupportRollback
}
