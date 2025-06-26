#!/bin/bash
# Simple script to bump versions for new release to document all places that need to be updated
set -e
old_version=1.1.1
new_version=1.2.0

REPO_DIR="$( cd "$(dirname "$( dirname "${BASH_SOURCE[0]}" )")" &> /dev/null && pwd )"

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


for manifest in ${REPO_DIR}/distributions/*/manifest.yaml; do
  if [ ! -f "${manifest}" ]; then
    echo "File missing: ${manifest}"
    exit 1
  fi
  current_version=$(yq ".dist.version" "${manifest}")
  if [[ ${current_version} != "${old_version}" ]]; then
    echo "Unexpected: Found ${current_version} instead of ${old_version} in ${manifest}"
    exit 1
  fi
  sed_inplace "s/version: ${old_version}/version: ${new_version}/g" ${manifest}
done

if ! grep -q "${old_version}" ${REPO_DIR}/distributions/README.md; then
  echo "Didn't find old version in README.md - does the script need updating?"
  exit 1
else
  sed_inplace "s/${old_version}/${new_version}/g" ${REPO_DIR}/distributions/README.md
fi

