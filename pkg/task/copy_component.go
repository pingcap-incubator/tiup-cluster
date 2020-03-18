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

import "github.com/pingcap-incubator/tiup/pkg/repository"

// CopyComponent is used to copy all files related the specific version a component
// to the target directory of path
type CopyComponent struct {
	component string
	version   repository.Version
	host      string
	path      string
}

// Execute implements the Task interface
func (c *CopyComponent) Execute(ctx *Context) error {
	panic("implement me")
}

// Rollback implements the Task interface
func (c *CopyComponent) Rollback(ctx *Context) error {
	panic("implement me")
}
