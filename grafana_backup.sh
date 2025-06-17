#!/bin/sh

set -a
. ./.env
set +a

docker run --user $(id -u):$(id -g) --rm --name grafana-backup-tool \
           -e GRAFANA_TOKEN=${GRAFANA_TOKEN} \
           -e GRAFANA_URL=${GRAFANA_URL} \
           -e GRAFANA_ADMIN_ACCOUNT=${GRAFANA_ADMIN_ACCOUNT} \
           -e GRAFANA_ADMIN_PASSWORD=${GRAFANA_ADMIN_PASSWORD} \
           -e VERIFY_SSL=${GRAFANA_BT_VERIFY_SSL} \
           --network host \
           -v ./grafana/backup:/opt/grafana-backup-tool/_OUTPUT_  \
           ysde/docker-grafana-backup-tool
