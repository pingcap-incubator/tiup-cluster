#!/bin/sh 

eval $(ssh-agent) &> /dev/null
ssh-add /root/.ssh/id_rsa &> /dev/null

export PATH=$PATH:/usr/local/go/bin
export PATH=$PATH:/tiops/bin

cat <<EOF >> /tmp/topology.yaml
tidb_servers:
  - host: 172.19.0.101

pd_servers:
  - host: 172.19.0.102

tikv_servers:
  - host: 172.19.0.103
EOF

cat <<EOF >> /tmp/scale-out-tidb.yaml
tidb_servers:
  - host: 172.19.0.104
EOF

cat <<EOF >> /tmp/scale-out-tikv.yaml
tikv_servers:
  - host: 172.19.0.104
  - host: 172.19.0.105
  - host: 172.19.0.102
EOF

cat <<EOF >> /tmp/scale-out-pd.yaml
pd_servers:
  - host: 172.19.0.105
EOF

echo "Deploy and start"
tiops deploy --key ~/.ssh/id_rsa test-cluster v3.0.12 /tmp/topology.yaml
tiops start test-cluster
sleep 5
tiops display test-cluster

echo "Scaling out TiDB"
tiops scale-out --key ~/.ssh/id_rsa test-cluster /tmp/scale-out-tidb.yaml
sleep 5
tiops display test-cluster
echo "Scaling out TiKV"
tiops scale-out --key ~/.ssh/id_rsa test-cluster /tmp/scale-out-tikv.yaml
sleep 5
tiops display test-cluster
echo "Scaling out PD"
tiops scale-out --key ~/.ssh/id_rsa test-cluster /tmp/scale-out-pd.yaml
sleep 5
tiops display test-cluster

echo "Restarting TiDB"
tiops restart test-cluster --role tidb
sleep 5
tiops display test-cluster
echo "Restarting PD"
tiops restart test-cluster --role pd
sleep 5
tiops display test-cluster
echo "Restarting TiKV"
tiops restart test-cluster --role tikv
sleep 5
tiops display test-cluster

echo "Stopping TiDB"
tiops stop test-cluster --role tidb
sleep 5
tiops display test-cluster
echo "Stopping TiKV"
tiops stop test-cluster --role tikv
sleep 5
tiops display test-cluster
echo "Stopping PD"
tiops stop test-cluster --role pd
sleep 5
tiops display test-cluster

echo "Starting"
tiops start test-cluster
sleep 5
tiops display test-cluster 
echo "Restarting"
tiops restart test-cluster
sleep 5
tiops display test-cluster

echo "Upgrading"
tiops upgrade test-cluster v4.0.0-beta.1
sleep 5
tiops display test-cluster

echo "Scaling in PD"
tiops scale-in test-cluster --node 172.19.0.102:2379
sleep 5
tiops display test-cluster
tiops restart test-cluster
sleep 5
tiops display test-cluster
echo "Scaling in TiKV"
tiops scale-in test-cluster --node 172.19.0.104:20160
sleep 5
tiops display test-cluster
echo "Scaling in TiDB"
tiops scale-in test-cluster --node 172.19.0.101:4000
sleep 5
tiops display test-cluster

tiops destroy test-cluster
