package task

import (
	"context"
	"fmt"

	"github.com/pingcap/errors"
	"go.etcd.io/etcd/clientv3"
	"golang.org/x/sync/errgroup"

	"github.com/pingcap-incubator/tiup-cluster/pkg/meta"
)

// UpdateTopology is used to maintain the cluster meta information
type UpdateTopology struct {
	cluster  string
	metadata *meta.ClusterMeta
}

// String implements the fmt.Stringer interface
func (u *UpdateTopology) String() string {
	return fmt.Sprintf("UpdateTopology: cluster=%s", u.cluster)
}

// Execute implements the Task interface
func (u *UpdateTopology) Execute(ctx *Context) error {
	newMeta := &meta.ClusterMeta{}
	*newMeta = *u.metadata
	newMeta.Topology = &meta.TopologySpecification{
		GlobalOptions:    u.metadata.Topology.GlobalOptions,
		MonitoredOptions: u.metadata.Topology.MonitoredOptions,
		ServerConfigs:    u.metadata.Topology.ServerConfigs,
	}

	topo := u.metadata.Topology

	instances := (&meta.MonitorComponent{Specification: topo}).Instances()
	instances = append(instances, (&meta.GrafanaComponent{Specification: topo}).Instances()...)
	instances = append(instances, (&meta.AlertManagerComponent{Specification: topo}).Instances()...)

	client, err := newMeta.Topology.GetEtcdClient()
	if err != nil {
		return err
	}
	// Remove all keys under topology
	_, err = client.KV.Delete(context.Background(), "/topology", clientv3.WithPrefix())
	if err != nil {
		return err
	}

	errg, _ := errgroup.WithContext(context.Background())

	for _, ins := range instances {
		ins := ins
		errg.Go(func() error {

			err := ins.UpdateTopology()
			if err != nil {
				return errors.AddStack(err)
			}
			return nil
		})
	}

	return errg.Wait()
}

// Rollback implements the Task interface
func (u *UpdateTopology) Rollback(ctx *Context) error {
	panic("implement me")
}
