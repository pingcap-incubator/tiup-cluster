#!/bin/sh

function run_test() {
  docker exec -i tiops-control bash -c "/tmp/test.sh $1"
}

function check() {
  local result=$(docker exec -i tiops-control bash -c "/tmp/test.sh check-$1")
  if [ ! "$result" == "SUCCESS" ]; then
    echo $result
    exit 1
  fi
}

cat run.sh | docker exec -i tiops-control bash -c "cat > /tmp/test.sh"
docker exec -i tiops-control bash -c "chmod +x /tmp/test.sh"

run_test "prepare"
run_test "deploy"
run_test "start"
check "prepare"
check "old-version"
check "select"

run_test "scale-out-tidb"
check "select"

run_test "scale-out-tikv"
check "select"

run_test "scale-out-pd"
check "select"

run_test "restart-tidb"
check "select"

run_test "restart-pd"
check "select"

run_test "restart-tikv"
check "select"

run_test "stop-tidb"
run_test "stop-tikv"
run_test "stop-pd"

run_test "restart"
check "select"

run_test "upgrade"
check "select"

run_test "scale-in-pd"
run_test "restart"
check "select"

run_test "scale-in-tikv"
run_test "restart"
check "select"

run_test "scale-in-tidb"
check "select"

run_test "destroy"
