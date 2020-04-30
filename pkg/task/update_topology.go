package task

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pingcap-incubator/tiup/pkg/set"

	"github.com/pingcap/errors"
	"go.etcd.io/etcd/clientv3"
	"golang.org/x/sync/errgroup"

	"github.com/pingcap-incubator/tiup-cluster/pkg/meta"
)

// UpdateTopology is used to maintain the cluster meta information
type UpdateTopology struct {
	cluster        string
	metadata       *meta.ClusterMeta
	deletedNodesID []string
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

	deleted := set.NewStringSet(u.deletedNodesID...)

	var deletedDBInstance []meta.TiDBSpec
	for i, instance := range (&meta.TiDBComponent{ClusterSpecification: topo}).Instances() {
		if deleted.Exist(instance.ID()) {
			deletedDBInstance = append(deletedDBInstance, topo.TiDBServers[i])
		}
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

	for _, dbIns := range deletedDBInstance {
		dbIns := dbIns
		errg.Go(func() error {
			err := deleteOfflineTiDBTopology(dbIns, client)
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

func deleteOfflineTiDBTopology(instance meta.TiDBSpec, etcdClient *clientv3.Client) error {
	_, err := etcdClient.KV.Delete(context.Background(), fmt.Sprintf("/topology/tidb/%s:%d", instance.Host, instance.Port), clientv3.WithPrefix())
	return err
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
