#!/bin/bash
set -e
ulimit -n 1000000

# WARNING: This file was auto-generated. Do not edit!
#          All your edit might be overwritten!
DEPLOY_DIR={{.DeployDir}}

cd "${DEPLOY_DIR}" || exit 1

export RUST_BACKTRACE=1

export LD_LIBRARY_PATH=/usr/local/lib::/usr/local/lib::/usr/local/lib64:/usr/local/lib:/usr/lib64:/usr/lib:{{.DeployDir}}:{{.DeployDir}}/tiflash_lib:$LD_LIBRARY_PATH

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