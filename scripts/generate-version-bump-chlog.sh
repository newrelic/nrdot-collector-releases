#!/bin/bash
# Copyright New Relic, Inc. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

# Generates a changelog entry for an otel component version bump.
#
# Must be run after the PR is created so that `make chlog-new` can populate the
# .issues field with the PR number.
#
# Usage: generate-version-bump-chlog.sh -c <current_beta_core> -n <next_beta_core>
#   -c  current otel beta core version (e.g. v0.147.0)
#   -n  next otel beta core version    (e.g. v0.148.0)

set -euo pipefail

CURRENT_BETA_CORE=''
NEXT_BETA_CORE=''

while getopts c:n: flag
do
    case "${flag}" in
        c) CURRENT_BETA_CORE=${OPTARG};;
        n) NEXT_BETA_CORE=${OPTARG};;
        *) exit 1;;
    esac
done

if [[ -z "$CURRENT_BETA_CORE" || -z "$NEXT_BETA_CORE" ]]; then
    echo "Usage: $0 -c <current_beta_core> -n <next_beta_core>" >&2
    exit 1
fi

filepath=$(make -s chlog-new)

# The .issues and .change_type fields are automatically populated by the make target if a PR has been created prior
yq -i "
  .component = \"distributions\" |
  .note = \"Bump otel component versions from ${CURRENT_BETA_CORE} to ${NEXT_BETA_CORE}\" |
  ... comments=\"\"
" "$filepath"

echo "New changelog entry added:"
cat "$filepath"

make chlog-validate