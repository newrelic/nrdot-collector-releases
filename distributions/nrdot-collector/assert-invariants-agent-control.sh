#!/bin/bash
set -e
yq --version | grep -E -q '4\.[0-9]+\.[0-9]+' ||
  { echo $? && echo "yq version 4.x.x is expected, but was '$(yq --version)'"; exit 1; }

script_name=$(basename "$0")
script_dir=$(realpath "$(dirname "${BASH_SOURCE[0]}")")
host_distro_dir="${script_dir}"

# assert config file location
test -f "${host_distro_dir}/config.yaml" ||
  { echo "expect config at config.yaml"; exit 1; }

# assert unchanged env vars
env_vars=(
  'NEW_RELIC_MEMORY_LIMIT_MIB'
  'OTEL_EXPORTER_OTLP_ENDPOINT'
  'NEW_RELIC_LICENSE_KEY'
)
for env_var in "${env_vars[@]}"; do
  echo "Checking for env var ${env_var}"
  grep -E '\${env:[^}]+}' "${host_distro_dir}/config.yaml" |
  grep "${env_var}" -q ||
    { echo "expected env var '${env_var}' in config.yaml"; exit 1; }
done

echo 'Checking for host.id detection'
yq -e '.processors.resourcedetection.system.resource_attributes["host.id"].enabled == "true"' "${host_distro_dir}/config.yaml" ||
  { echo "expected host.id detection to be enabled"; exit 1; }


# assert binary name
goreleaser_yamls=('.goreleaser.yaml' '.goreleaser-nightly.yaml')
for goreleaser_yaml in "${goreleaser_yamls[@]}"; do
  echo "Checking ${goreleaser_yaml}"
  yq -e '.builds[].binary == "nrdot-collector"' "${host_distro_dir}/${goreleaser_yaml}" ||
    { echo "expected binary name 'nrdot-collector' in ${goreleaser_yaml}"; exit 1; }
  yq -e '.nfpms[].package_name == "nrdot-collector"' "${host_distro_dir}/${goreleaser_yaml}" ||
    { echo "expected package_name 'nrdot-collector' in ${goreleaser_yaml}"; exit 1; }
done




