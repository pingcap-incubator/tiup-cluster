package task

import (
	"context"
	"encoding/json"
	"fmt"

	"go.etcd.io/etcd/clientv3"

	"github.com/pingcap-incubator/tiup/pkg/set"

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
	client, err := u.metadata.Topology.GetEtcdClient()
	if err != nil {
		return err
	}
	txn := client.Txn(context.Background())

	topo := u.metadata.Topology

	instances := (&meta.MonitorComponent{ClusterSpecification: topo}).Instances()
	instances = append(instances, (&meta.GrafanaComponent{ClusterSpecification: topo}).Instances()...)
	instances = append(instances, (&meta.AlertManagerComponent{ClusterSpecification: topo}).Instances()...)

	deleted := set.NewStringSet(u.deletedNodesID...)

	ops := []clientv3.Op{
		clientv3.OpDelete("/topology/prometheus"),
		clientv3.OpDelete("/topology/grafana"),
		clientv3.OpDelete("/topology/alertmanager"),
	}

	for _, instance := range (&meta.TiDBComponent{ClusterSpecification: topo}).Instances() {
		if deleted.Exist(instance.ID()) {
			ops = append(ops, clientv3.OpDelete(fmt.Sprintf("/topology/tidb/%s:%d", instance.GetHost(), instance.GetPort()), clientv3.WithPrefix()))
		}
	}

	for _, ins := range instances {
		op, err := updateTopologyOp(ins)
		if err != nil {
			return err
		}
		ops = append(ops, *op)
	}

	_, err = txn.Then(ops...).Commit()
	return err
}

// componentTopology represent the topology info for alertmanager, prometheus and grafana.
type componentTopology struct {
	IP         string `json:"ip"`
	Port       int    `json:"port"`
	DeployPath string `json:"deploy_path"`
}

func updateTopologyOp(instance meta.Instance) (*clientv3.Op, error) {
	switch instance.ComponentName() {
	case meta.ComponentAlertManager, meta.ComponentPrometheus, meta.ComponentGrafana:
		topology := componentTopology{
			IP:         instance.GetHost(),
			Port:       instance.GetPort(),
			DeployPath: instance.DeployDir(),
		}
		data, err := json.Marshal(topology)
		if err != nil {
			return nil, err
		}
		op := clientv3.OpPut("/topology/"+instance.ComponentName(), string(data))
		return &op, nil
	default:
		panic("updateTopologyOp receive wrong arguments, logic error!")
	}
}

// Rollback implements the Task interface
func (u *UpdateTopology) Rollback(ctx *Context) error {
	return nil
}
