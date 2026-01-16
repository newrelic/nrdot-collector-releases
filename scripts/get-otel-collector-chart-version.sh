#!/bin/bash
# Copyright New Relic, Inc. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

REPO_DIR="$( cd "$(dirname "$( dirname "${BASH_SOURCE[0]}" )")" &> /dev/null && pwd )"

collector_component_version=$(${REPO_DIR}/scripts/get-core-component-version.sh)

helm repo list | grep -q open-telemetry || helm repo add open-telemetry https://open-telemetry.github.io/opentelemetry-helm-charts >&2

chart_version=$(helm search repo open-telemetry/opentelemetry-collector --output=json --versions | jq -r ".[] | select (.app_version==\"${collector_component_version}\") | .version" | head -n1 )
if [[ -z "${chart_version}" ]]; then
  chart_version="$(helm search repo open-telemetry/opentelemetry-collector --output=json | jq -r '.[0].version')"
  echo "Chart matching collector component version ${collector_component_version} wasn't released (yet?). Using latest (${chart_version}) instead." >&2
fi

echo "${chart_version}"
