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

package operator

import (
	"fmt"
	"github.com/pingcap-incubator/tiup/pkg/set"
	"time"

	"github.com/pingcap-incubator/tiops/pkg/executor"
	"github.com/pingcap-incubator/tiops/pkg/meta"
)

// Options represents the operation options
type Options struct {
	Roles []string
	Nodes []string
	Force bool // Option for upgrade subcommand
}

// Operation represents the type of cluster operation
type Operation byte

// Operation represents the kind of cluster operation
const (
	StartOperation Operation = iota
	StopOperation
	RestartOperation
	DestroyOperation
	UpgradeOperation
	ScaleInOperation
	ScaleOutOperation
)

var defaultTimeoutForReady = time.Second * 60

var opStringify = [...]string{
	"StartOperation",
	"StopOperation",
	"RestartOperation",
	"DestroyOperation",
	"UpgradeOperation",
	"ScaleInOperation",
	"ScaleOutOperation",
}

func (op Operation) String() string {
	if op <= ScaleOutOperation {
		return opStringify[op]
	}
	return fmt.Sprintf("unknonw-op(%d)", op)
}

func filterComponent(comps []meta.Component, components set.StringSet) (res []meta.Component) {
	if len(components) == 0 {
		res = comps
		return
	}

	for _, c := range comps {
		if !components.Exist(c.Name()) {
			continue
		}

		res = append(res, c)
	}

	return
}

func filterInstance(instances []meta.Instance, nodes set.StringSet) (res []meta.Instance) {
	if len(nodes) == 0 {
		res = instances
		return
	}

	for _, c := range instances {
		if !nodes.Exist(c.ID()) {
			continue
		}
		res = append(res, c)
	}

	return
}

// ExecutorGetter get the executor by host.
type ExecutorGetter interface {
	Get(host string) (e executor.TiOpsExecutor)
}
