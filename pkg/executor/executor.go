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

package executor

// TiOpsExecutor is the executor interface for TiOps, any tasks will in the end
// be passed to a executor and then be actually performed.
type TiOpsExecutor interface {
	// Init builds and initializes an executor
	Init(config interface{}) error
	// Exec run the command, then return it's stdout and stderr
	Exec(cmd string, stdin []byte, sudo bool) (stdout []byte, stderr []byte, err error)
	// Transfer copies files from or to a target
	Transfer(src string, dst string) error
}
