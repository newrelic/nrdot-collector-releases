#!/bin/bash
# Copyright New Relic, Inc. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

set -eo pipefail
yq --version | grep -E -q '4\.[0-9]+\.[0-9]+' ||
  { echo $? && echo "yq version 4.x.x is expected, but was '$(yq --version)'"; exit 1; }

script_dir=$(realpath "$(dirname "${BASH_SOURCE[0]}")")
repo_root="${script_dir}/.."
distros_dir="${repo_root}/distributions"
allowlist_file="${repo_root}/internal/assets/license/otel-library-allowlist.json"

gomod_query='.receivers[].gomod, .processors[].gomod, .exporters[].gomod, .connectors[].gomod, .extensions[].gomod, .providers[].gomod'

allowlist_raw=$(yq -p json -o yaml '.allowlist[]' "${allowlist_file}")
mapfile -t allowlist_patterns <<< "${allowlist_raw}"

for core_distro_dir in "${distros_dir}"/*/; do
  manifests=()
  for manifest in manifest.yaml manifest-fips.yaml; do
    test -f "${core_distro_dir}${manifest}" && manifests+=("${core_distro_dir}${manifest}")
  done

  if [[ ${#manifests[@]} -eq 0 ]]; then
    echo "expected manifest.yaml or manifest-fips.yaml in ${core_distro_dir}"
    exit 1
  fi

  otel_libs=()
  for manifest in "${manifests[@]}"; do
    libs_raw=$(yq -e "${gomod_query}" "${manifest}" | awk '{print $1}')
    [[ -n "${libs_raw}" ]] && mapfile -t -O "${#otel_libs[@]}" otel_libs <<< "${libs_raw}"
  done

  unique_libs=()
  [[ ${#otel_libs[@]} -gt 0 ]] && mapfile -t unique_libs < <(printf '%s\n' "${otel_libs[@]}" | sort -u)

  disallowed_libs=()
  for lib in "${unique_libs[@]}"; do
    allowed=false
    for pattern in "${allowlist_patterns[@]}"; do
      if [[ "${lib}" == ${pattern//%/*} ]]; then
        allowed=true
        break
      fi
    done
    [[ "${allowed}" == true ]] || disallowed_libs+=("${lib}")
  done

  if [[ ${#disallowed_libs[@]} -gt 0 ]]; then
    echo "Found gomod libraries in ${core_distro_dir} not covered by the otel library allowlist:"
    printf '  %s\n' "${disallowed_libs[@]}"
    exit 1
  fi
done