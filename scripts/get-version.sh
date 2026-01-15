#!/bin/bash
# Copyright New Relic, Inc. All rights reserved.
# SPDX-License-Identifier: Apache-2.0


# This script validates that all manifest.yaml files in the distributions directory have the same version number.
# If they do, it prints the version number to stdout and exits with status 0.
# If they don't, it prints an error message to stderr and exits with status 1.

set -e

version=""

for manifest in ./distributions/*/manifest.yaml; do
  if [ -f "$manifest" ]; then
    current_version=$(grep -E '^  version:' "$manifest" | awk '{print $2}')
    if [ -z "$version" ]; then
      version="$current_version"
    elif [ "$version" != "$current_version" ]; then
      echo "Version mismatch detected in $manifest"
      exit 1
    fi
  fi
done

if [ -z "$version" ]; then
  echo "No version found in any manifest.yaml"
  exit 1
fi

if ! grep -q "collector_version=\"${version}\"" ./distributions/README.md; then
  echo "README was not updated to use ${version} for install instructions"
  exit 1
fi

echo $version
