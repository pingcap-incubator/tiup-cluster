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
	"io/ioutil"

	"github.com/pingcap-incubator/tiops/pkg/module"
	"github.com/pingcap/errors"
)

// EnvInit is used to initialize the remote environment, e.g:
// 1. Generate SSH key
// 2. ssh-copy-id
type EnvInit struct {
	host       string
	deployUser string
}

// Execute implements the Task interface
func (e *EnvInit) Execute(ctx *Context) error {
	exec, found := ctx.GetExecutor(e.host)
	if !found {
		return ErrNoExecutor
	}

	um := module.NewUserModule(module.UserModuleConfig{
		Action: module.UserActionAdd,
		Name:   e.deployUser,
		Sudoer: true,
	})

	stdout, stderr, err := um.Execute(exec)
	if err != nil {
		return errors.Trace(err)
	}

	fmt.Println("Create user stdout: ", string(stdout))
	fmt.Println("Create user stderr: ", string(stderr))

	pubKey, err := ioutil.ReadFile(ctx.PublicKeyPath)
	if err != nil {
		return errors.Trace(err)
	}

	// Authorize
	cmd := `su - ` + e.deployUser + ` -c 'test -d ~/.ssh || mkdir -p ~/.ssh && chmod 700 ~/.ssh'`
	stdout, stderr, err = exec.Execute(cmd, false)
	if err != nil {
		return errors.Annotatef(err, "cmd: %s", cmd)
	}

	fmt.Println("Create ssh directory stdout: ", string(stdout))
	fmt.Println("Create ssh directory stderr: ", string(stderr))

	// TODO: don't append pubkey if exists
	cmd = `su - ` + e.deployUser + ` -c 'echo "` + string(pubKey) + `" >> .ssh/authorized_keys && chmod 700 ~/.ssh/authorized_keys'`
	stdout, stderr, err = exec.Execute(cmd, false)
	if err != nil {
		return errors.Trace(err)
	}

	fmt.Println("Add pubkey to `.ssh/authorized_keys` stdout: ", string(stdout))
	fmt.Println("Add pubkey to `.ssh/authorized_keys` stderr: ", string(stderr))

	return nil
}

// Rollback implements the Task interface
func (e *EnvInit) Rollback(ctx *Context) error {
	return ErrUnsupportRollback
}
