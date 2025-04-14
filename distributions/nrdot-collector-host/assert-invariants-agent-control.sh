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

# TODO: assert invariants within config

# assert binary name
goreleaser_yamls=('.goreleaser.yaml' '.goreleaser-nightly.yaml')
for goreleaser_yaml in "${goreleaser_yamls[@]}"; do
  echo "Checking ${goreleaser_yaml}"
  yq -e '.builds[].binary == "nrdot-collector-host"' "${host_distro_dir}/${goreleaser_yaml}" ||
    { echo "expected binary name 'nrdot-collector-host' in ${goreleaser_yaml}"; exit 1; }
  yq -e '.nfpms[].package_name == "nrdot-collector-host"' "${host_distro_dir}/${goreleaser_yaml}" ||
    { echo "expected package_name 'nrdot-collector-host' in ${goreleaser_yaml}"; exit 1; }
  yq -e '.nfpms[].contents[] | select(.src == "nrdot-collector-host.conf") | .dst == "/etc/nrdot-collector-host/nrdot-collector-host.conf"' "${host_distro_dir}/${goreleaser_yaml}" ||
    { echo "expected file 'nrdot-collector-host' in ${goreleaser_yaml}"; exit 1; }
done




