#!/bin/bash
set -e
ulimit -n 1000000

# WARNING: This file was auto-generated. Do not edit!
#          All your edit might be overwritten!
cd "{{.DeployDir}}" || exit 1

export RUST_BACKTRACE=1

export TZ=${TZ:-/etc/localtime}
export LD_LIBRARY_PATH=/data1/deploy-test-tiflash-ansible/bin/tiflash:$LD_LIBRARY_PATH

echo -n 'sync ... '
stat=$(time sync || sync)
echo ok
echo $stat

{{- if .NumaNode}}
exec numactl --cpunodebind={{.NumaNode}} --membind={{.NumaNode}} bin/tiflash/tiflash \
{{- else}}
exec bin/tiflash/tiflash \
{{- end}}
    --config-file conf/tiflash.toml
