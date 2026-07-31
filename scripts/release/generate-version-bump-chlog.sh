#!/bin/bash
# Copyright New Relic, Inc. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

# Generates a changelog entry for an otel component version bump.
#
# Must be run after the PR is created so that `make chlog-new` can populate the
# .issues field with the PR number.
#
# Usage: generate-version-bump-chlog.sh -c <current_beta_core> -n <next_beta_core> -f <next_beta_fork>
#   -c  current otel beta core version (e.g. v0.147.0)
#   -n  next otel beta core version    (e.g. v0.148.0)
#   -f  next otel beta fork version    (e.g. v0.148.0-nr1)

set -euo pipefail

CURRENT_BETA_CORE=''
NEXT_BETA_CORE=''
NEXT_BETA_FORK=''

while getopts c:n:f: flag
do
    case "${flag}" in
        c) CURRENT_BETA_CORE=${OPTARG};;
        n) NEXT_BETA_CORE=${OPTARG};;
        f) NEXT_BETA_FORK=${OPTARG};;
        *) exit 1;;
    esac
done

if [[ -z "$CURRENT_BETA_CORE" || -z "$NEXT_BETA_CORE" || -z "$NEXT_BETA_FORK" ]]; then
    echo "Usage: $0 -c <current_beta_core> -n <next_beta_core> -f <next_beta_fork>" >&2
    exit 1
fi

core_filepath=$(make -s chlog-new CHLOG_FILE=core_version_bump)

# The .issues and .change_type fields are automatically populated by the make target if a PR has been created prior
yq -i "
  .note = \"Bump otel component versions from ${CURRENT_BETA_CORE} to ${NEXT_BETA_CORE}\" |
  ... comments=\"\"
" "$core_filepath"

echo "New changelog entry added:"
cat "$core_filepath"

fork_filepath=$(make -s chlog-new CHLOG_FILE=fork_version_bump)

yq -i "
  .note = \"Bump \`nroracledbreceiver\` and \`nrsqlserverreceiver\` to ${NEXT_BETA_FORK}\" |
  .subtext = \"- For the list of changes to these components, refer to [their changelog](https://github.com/newrelic-forks/opentelemetry-collector-contrib/blob/receiver/nrsqlserverreceiver/${NEXT_BETA_FORK}/NR_CHANGELOG.md).\" |
  ... comments=\"\"
" "$fork_filepath"

echo "New changelog entry added:"
cat "$fork_filepath"

make chlog-validate
