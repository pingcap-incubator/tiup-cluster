#!/bin/bash
set -e
ulimit -n 1000000

# WARNING: This file was auto-generated. Do not edit!
#          All your edit might be overwritten!
DEPLOY_DIR={{.DeployDir}}
cd "${DEPLOY_DIR}" || exit 1

exec > >(tee -i -a "log/blackbox_exporter.log")
exec 2>&1

{{- if .NumaNode}}
exec numactl --cpunodebind={{.NumaNode}} --membind={{.NumaNode}} bin/blackbox_exporter \
{{- else}}
exec bin/blackbox_exporter \
{{- end}}
    --web.listen-address=":{{.Port}}" \
    --log.level="info" \
    --config.file="conf/blackbox.yml"