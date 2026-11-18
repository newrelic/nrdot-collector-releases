#!/bin/bash
# Copyright New Relic, Inc. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

# Records the pending .chloggen entries as structured YAML under the given version in changelog.yaml.
# A version without entries is recorded with empty buckets so it never blocks a release.
# Usage: record-changelog.sh <version>

set -euo pipefail

VERSION="${1:?Usage: $0 <version>}"
REPO_DIR="$(git rev-parse --show-toplevel)"
CHANGELOG_YAML="$REPO_DIR/changelog.yaml"

[ -f "$CHANGELOG_YAML" ] || echo '{}' > "$CHANGELOG_YAML"

entry_files=$(find "$REPO_DIR/.chloggen" -maxdepth 1 -name '*.yaml' ! -name 'config.yaml' ! -name 'TEMPLATE.yaml')
if [ -n "$entry_files" ]; then
  ENTRIES="$(yq ea -o=json -I=0 '[.]' $entry_files)"
else
  echo "::warning::No .chloggen entries found, recording $VERSION with empty changelog" >&2
  ENTRIES='[]'
fi

# Entries are bucketed by change_type; the description is only kept when subtext is set.
export VERSION ENTRIES
yq -i '
  (env(ENTRIES) | map({"change_type": .change_type, "summary": .note, "description": ((.subtext // "") | sub("\n+$"; "")), "issues": .issues} | del(.description | select(. == "")))) as $e
  | .[strenv(VERSION)] = {
      "features": ($e | map(select(.change_type == "feature") | del(.change_type))),
      "enhancements": ($e | map(select(.change_type == "docs") | del(.change_type))),
      "bug-fixes": ($e | map(select(.change_type == "bug_fix") | del(.change_type))),
      "security": []
    }
  | (.. | select(tag == "!!map" or tag == "!!seq")) style = ""' "$CHANGELOG_YAML"
