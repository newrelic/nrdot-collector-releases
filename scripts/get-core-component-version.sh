#!/bin/bash

# This script validates that all manifest.yaml files in the distributions directory have the same version number.
# If they do, it prints the version number to stdout and exits with status 0.
# If they don't, it prints an error message to stderr and exits with status 1.

set -e

REPO_DIR="$( cd "$(dirname "$( dirname "${BASH_SOURCE[0]}" )")" &> /dev/null && pwd )"

version=""

for manifest in ${REPO_DIR}/distributions/*/manifest.yaml; do
  if [ -f "$manifest" ]; then
    # batchprocessor as representative of the core component version
    current_version=$(grep batchprocessor "$manifest" | awk '{print $NF}')
    if [ -z "$version" ]; then
      version="$current_version"
    elif [ "$version" != "$current_version" ]; then
      echo "Core component version mismatch detected in $manifest"
      exit 1
    fi
  fi
done

if [ -z "$version" ]; then
  echo "No core component version found in any manifest.yaml"
  exit 1
fi

echo "${version:1}"
