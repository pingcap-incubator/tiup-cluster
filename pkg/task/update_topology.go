package task

import (
	"context"
	"encoding/json"
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

	topo := u.metadata.Topology

	instances := (&meta.MonitorComponent{ClusterSpecification: topo}).Instances()
	instances = append(instances, (&meta.GrafanaComponent{ClusterSpecification: topo}).Instances()...)
	instances = append(instances, (&meta.AlertManagerComponent{ClusterSpecification: topo}).Instances()...)

	client, err := u.metadata.Topology.GetEtcdClient()
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
			err := updateTopology(ins, client)
			if err != nil {
				return errors.AddStack(err)
			}
			return nil
		})
	}

	return errg.Wait()
}

// componentTopology represent the topology info for alertmanager, prometheus and grafana.
type componentTopology struct {
	IP         string `json:"ip"`
	Port       int    `json:"port"`
	DeployPath string `json:"deploy_path"`
}

// updateTopology write component topology to "/topology".
func updateTopology(instance meta.Instance, etcdClient *clientv3.Client) error {
	switch instance.ComponentName() {
	case meta.ComponentAlertManager, meta.ComponentPrometheus, meta.ComponentGrafana:
		topology := componentTopology{
			IP:         instance.GetHost(),
			Port:       instance.GetPort(),
			DeployPath: instance.DeployDir(),
		}
		data, err := json.Marshal(topology)
		if err != nil {
			return err
		}
		_, err = etcdClient.KV.Put(context.Background(), "/topology/"+instance.ComponentName(), string(data))
		return err
	default:
		return nil
	}
}

// Rollback implements the Task interface
func (u *UpdateTopology) Rollback(ctx *Context) error {
	return nil
}
