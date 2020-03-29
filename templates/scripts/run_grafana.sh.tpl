#!/bin/bash
set -e

# WARNING: This file was auto-generated. Do not edit!
#          All your edit might be overwritten!
DEPLOY_DIR={{.DeployDir}}
cd "${DEPLOY_DIR}" || exit 1

mkdir -p {{.DeployDir}}/plugins
mkdir -p {{.DeployDir}}/dashboards
mkdir -p {{.DeployDir}}/provisioning/dashboards
mkdir -p {{.DeployDir}}/provisioning/datasources

cp {{.DeployDir}}/bin/*.json {{.DeployDir}}/dashboards/
cp {{.DeployDir}}/config/datasource.yml {{.DeployDir}}/provisioning/datasources
cp {{.DeployDir}}/config/dashboard.yml {{.DeployDir}}/provisioning/dashboards

LANG=en_US.UTF-8 \
{{- if .NumaNode}}
exec numactl --cpunodebind={{.NumaNode}} --membind={{.NumaNode}} bin/bin/grafana-server \
{{- else}}
exec bin/bin/grafana-server \
{{- end}}
    --homepath="{{.DeployDir}}/bin" \
    --config="{{.DeployDir}}/config/grafana.ini"