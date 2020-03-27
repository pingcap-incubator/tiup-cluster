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
	"io"
	"time"

	"github.com/pingcap-incubator/tiops/pkg/api"
	"github.com/pingcap-incubator/tiops/pkg/meta"
	"github.com/pingcap-incubator/tiup/pkg/set"
	"github.com/pingcap/errors"
)

// ScaleIn scales in the cluster
func ScaleIn(
	getter ExecutorGetter,
	w io.Writer,
	spec *meta.Specification,
	options Options,
) error {
	// instances by uuid
	instances := map[string]meta.Instance{}

	// make sure all nodeIds exists in topology
	for _, component := range spec.ComponentsByStartOrder() {
		for _, instance := range component.Instances() {
			instances[instance.ID()] = instance
		}
	}

	// Clean components
	deletedDiff := map[string][]meta.Instance{}
	deletedNodes := set.NewStringSet(options.DeletedNodes...)
	for nodeID := range deletedNodes {
		inst, found := instances[nodeID]
		if !found {
			return errors.Errorf("cannot find node id '%s' in topology", nodeID)
		}
		deletedDiff[inst.ComponentName()] = append(deletedDiff[inst.ComponentName()], inst)
	}

	// Cannot delete all PD servers
	if len(deletedDiff[meta.ComponentPD]) == len(spec.PDServers) {
		return errors.New("cannot delete all PD servers")
	}

	// Cannot delete all TiKV servers
	if len(deletedDiff[meta.ComponentTiKV]) == len(spec.TiKVServers) {
		return errors.New("cannot delete all TiKV servers")
	}

	asyncOfflineComps := set.NewStringSet(meta.ComponentPump, meta.ComponentTiKV, meta.ComponentDrainer)

	// At least a PD server exists
	var pdClient *api.PDClient
	for _, instance := range (&meta.PDComponent{Specification: spec}).Instances() {
		if !deletedNodes.Exist(instance.ID()) {
			pdClient = api.NewPDClient(instance.GetHost(), 10*time.Second, nil)
		}
	}

	if pdClient == nil {
		return errors.New("cannot find available PD instance")
	}

	// Delete member from cluster
	for _, component := range spec.ComponentsByStartOrder() {
		for _, instance := range component.Instances() {
			if !deletedNodes.Exist(instance.ID()) {
				continue
			}

			switch component.Name() {
			case meta.ComponentTiKV:
				if err := pdClient.DelStore(instance.GetHost()); err != nil {
					return err
				}
			case meta.ComponentPD:
				if err := pdClient.DelPD(instance.(*meta.PDInstance).Name); err != nil {
					return err
				}
			case meta.ComponentDrainer:
				// TODO: binlog api
			case meta.ComponentPump:
				// TODO: binlog api
			}

			if !asyncOfflineComps.Exist(instance.ComponentName()) {
				if err := StopComponent(getter, w, []meta.Instance{instance}); err != nil {
					return errors.Annotatef(err, "failed to stop %s", component.Name())
				}
				if err := DestroyComponent(getter, w, []meta.Instance{instance}); err != nil {
					return errors.Annotatef(err, "failed to destroy %s", component.Name())
				}
			}
		}
	}

	return nil
}
