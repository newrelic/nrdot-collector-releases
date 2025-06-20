#!/bin/bash
set -e
yq --version | grep -E -q '4\.[0-9]+\.[0-9]+' || { echo $? && echo "yq version 4.x.x is expected, but was '$(yq --version)'"; exit 1; }

helm repo add newrelic https://helm-charts.newrelic.com
helm repo update
rendered_helm_chart=$(helm template "test-render-$(date +%s)" newrelic/nr-k8s-otel-collector \
  --set 'cluster=${env:K8S_CLUSTER_NAME:-cluster-name-placeholder}' \
  --set 'lowDataMode=false' \
  --set licenseKey=license-key-placeholder \
  )


script_name=$(basename "$0")
script_dir=$(realpath "$(dirname "${BASH_SOURCE[0]}")")
k8s_distro_dir="${script_dir}"

function extract_config_with_overwritable_defaults() {
  local helm_deploy_type="${1}"
  local output_file="${k8s_distro_dir}/config-${helm_deploy_type}.yaml"
  echo "$rendered_helm_chart" |
    {
      yq "(select(.kind == \"ConfigMap\" and (.metadata.name | contains(\"${helm_deploy_type}-config\"))) | .data[\"${helm_deploy_type}-config.yaml\"])"
    } | {
      # strip away all comments
      yq '... comments=""'
    } | {
      # remove references to helm chart
      yq 'del(.processors[] | select(has("attributes")) | .[][] | select(.key == "newrelic.chart.version"))'
    } | {
      # remove configuration requiring mounted host filesystem
      yq 'del(.receivers.hostmetrics.root_path)'
    } | {
      # expose ingest endpoint via env var to align with other distros
      sed 's/"https:\/\/otlp.nr-data.net"/${env:OTEL_EXPORTER_OTLP_ENDPOINT:-https:\/\/otlp.nr-data.net}/g'
    } | {
      # normalize env var names
      sed 's/env:NR_LICENSE_KEY/env:NEW_RELIC_LICENSE_KEY/g'
    } | {
      # add healthcheck
      yq '. + {"extensions": {"health_check":{}}}' |
      yq '.service += {"extensions": ["health_check"]}'
    } > "${output_file}"
    echo "Config '${helm_deploy_type}' written to ${output_file}"
}

extract_config_with_overwritable_defaults 'daemonset'
extract_config_with_overwritable_defaults 'deployment'

# Document last synced version in the following line
# last synced version: 0.8.26
chart_version=$(helm show chart newrelic/nr-k8s-otel-collector | yq .version)
# Determine the OS and set the sed -i command accordingly
if [[ "$OSTYPE" == "darwin"* ]]; then
  # macOS
  function sed_inplace {
  	sed -i '' "$@"
  }
else
  function sed_inplace {
    	sed -i'' "$@"
  }
fi
sed_inplace -E "s/^# last synced version: .*/# last synced version: ${chart_version}/" "$0"
