#!/bin/bash 

eval $(ssh-agent) &> /dev/null
ssh-add /root/.ssh/id_rsa &> /dev/null

export PATH=$PATH:/usr/local/go/bin
export PATH=$PATH:/tiops/bin

if [[ "$1" == "prepare" ]]; then
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
fi

if [[ "$1" == "deploy" ]]; then
echo "Deploy"
yes|tiops deploy -i ~/.ssh/id_rsa test-cluster v3.0.12 /tmp/topology.yaml
sleep 10
tiops display test-cluster
fi

if [[ "$1" == "start" ]]; then
tiops start test-cluster 
sleep 10
tiops display test-cluster
fi

if [[ "$1" == "scale-out-tidb" ]]; then
echo "Scaling out TiDB"
yes|tiops scale-out -i ~/.ssh/id_rsa test-cluster /tmp/scale-out-tidb.yaml
sleep 10
tiops display test-cluster
fi

if [[ "$1" == "scale-out-tikv" ]]; then
echo "Scaling out TiKV"
yes|tiops scale-out -i ~/.ssh/id_rsa test-cluster /tmp/scale-out-tikv.yaml
sleep 10
tiops display test-cluster
fi

if [[ "$1" == "scale-out-pd" ]]; then
echo "Scaling out PD"
yes|tiops scale-out -i ~/.ssh/id_rsa test-cluster /tmp/scale-out-pd.yaml
sleep 10
tiops display test-cluster
fi

if [[ "$1" == "restart-tidb" ]]; then
echo "Restarting TiDB"
tiops restart test-cluster --role tidb
sleep 10
tiops display test-cluster
fi

if [[ "$1" == "restart-pd" ]]; then
echo "Restarting PD"
tiops restart test-cluster --role pd
sleep 10
tiops display test-cluster
fi

if [[ "$1" == "restart-tikv" ]]; then
echo "Restarting TiKV"
tiops restart test-cluster --role tikv
sleep 10
tiops display test-cluster
fi

if [[ "$1" == "stop-tidb" ]]; then
echo "Stopping TiDB"
tiops stop test-cluster --role tidb
sleep 10
tiops display test-cluster
fi

if [[ "$1" == "stop-tikv" ]]; then
echo "Stopping TiKV"
tiops stop test-cluster --role tikv
sleep 10
tiops display test-cluster
fi

if [[ "$1" == "stop-pd" ]]; then
echo "Stopping PD"
tiops stop test-cluster --role pd
sleep 10
tiops display test-cluster
fi

if [[ "$1" == "restart" ]]; then
echo "Restarting"
tiops restart test-cluster
sleep 10
tiops display test-cluster
fi

if [[ "$1" == "upgrade" ]]; then
echo "Upgrading"
tiops upgrade test-cluster v4.0.0-beta.1
sleep 10
tiops display test-cluster
fi

if [[ "$1" == "scale-in-pd" ]]; then
echo "Scaling in PD"
yes|tiops scale-in test-cluster --node 172.19.0.102:2379
sleep 10
tiops display test-cluster
fi

if [[ "$1" == "scale-in-tikv" ]]; then
echo "Scaling in TiKV"
yes|tiops scale-in test-cluster --node 172.19.0.104:20160
sleep 10
tiops display test-cluster
fi

if [[ "$1" == "scale-in-tidb" ]]; then
echo "Scaling in TiDB"
yes|tiops scale-in test-cluster --node 172.19.0.104:4000
sleep 10
tiops display test-cluster
fi

if [[ "$1" == "destroy" ]]; then
yes|tiops destroy test-cluster
fi

if [[ "$1" == "check-old-version" ]]; then
  version=$(echo "SELECT @@version" | mysql -h 172.19.0.101 -u root -P 4000 | tail -n1)
  if [[ ! "$version" == "5.7.25-TiDB-v3.0.12" ]]; then
    echo "FAIL"
    echo "version=$version"
    exit 1
  else
    echo "SUCCESS"
  fi
fi
if [[ "$1" == "check-new-version" ]]; then
  version=$(echo "SELECT @@version" | mysql -h 172.19.0.101 -u root -P 4000 | tail -n1)
  if [[ ! "$version" == "5.7.25-TiDB-v4.0.0-beta.1" ]]; then
    echo "FAIL"
    echo "version=$version"
    exit 1
  else
    echo "SUCCESS"
  fi
fi
if [[ "$1" == "check-prepare" ]]; then
  echo "CREATE TABLE \`test_table\` (  \`col_double\` double DEFAULT NULL,  \`col_binary_8_key\` binary(8) DEFAULT NULL, KEY \`col_binary_8_key\` (\`col_binary_8_key\`) );"|mysql -h 172.19.0.101 -u root -P 4000 -D test
  echo "INSERT INTO \`test_table\` VALUES(1000.0, \"lyashfuw\");" | mysql -h 172.19.0.101 -u root -P 4000 -D test
  echo "SUCCESS"
fi
if [[ "$1" == "check-select" ]]; then
  result=$(echo "SELECT \`col_double\` FROM \`test_table\` WHERE \`col_binary_8_key\`;show warnings;" | mysql -h 172.19.0.101 -u root -P 4000 -D test | tail -n1)
  case $result in
    Warning*1105*Eval*)
      echo "SUCCESS"
      ;;
    Warning*1292*Truncated*)
      echo "SUCCESS"
      ;;
    *)
      echo "FAIL"
      echo $result
      ;;
  esac
fi
