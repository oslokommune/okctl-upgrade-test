#!/usr/bin/env bash

# This patch enables basic JSON parsing of logs. See https://grafana.com/docs/loki/latest/clients/promtail/stages/docker/

WD=/tmp/okctl-promtail-json-patch

# Create work directory
mkdir -p ${WD}

# Fetch and decode existing config
kubectl -n monitoring get secret promtail -o yaml | yq eval '.data["promtail.yaml"]' - | base64 -d -w 0 > ${WD}/promtail.yaml

# Patch config to change parser to the docker parser which can read JSON
yq eval '.scrape_configs[] |= select(.job_name=="kubernetes-pods-app").pipeline_stages[0]={"docker": {}}' ${WD}/promtail.yaml > ${WD}/promtail-patched.yaml

# Recode patched config
CFG=$(cat ${WD}/promtail-patched.yaml | base64 -w 0)

# Create JSON patch with new config content
read -d '' PATCH <<EOF
[{
	"op": "replace",
	"path": "/data/promtail.yaml",
	"value": "${CFG}"
}]
EOF

# Patch the promtail secret containing the config
kubectl -n monitoring patch secret promtail --type=json --patch="${PATCH}"

# Restart promtail
POD=`kubectl -n monitoring get pods | grep promtail | cut -d' ' -f1`
kubectl -n monitoring delete pod ${POD}
