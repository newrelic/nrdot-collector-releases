#!/bin/bash
# Copyright New Relic, Inc. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

set -e

REPO_DIR="$( cd "$(dirname "$( dirname "${BASH_SOURCE[0]}" )")" &> /dev/null && pwd )"
ALLOWLIST_FILE="${REPO_DIR}/internal/assets/license/otel-library-allowlist.json"
SPEC_FILES=(
  "${REPO_DIR}/distributions/nrdot-collector/test/spec-local.yaml"
  "${REPO_DIR}/distributions/nrdot-collector/test/spec-nightly-kind.yaml"
)

clauses=$(jq -r --arg q "'" '.allowlist[] | "AND otel.library.name NOT LIKE " + $q + . + $q' "$ALLOWLIST_FILE" | paste -sd' ' -)
query="FROM Metric SELECT filter(uniqueCount(otel.library.name), WHERE otel.library.name IS NOT NULL ${clauses}) as c2_non_allowlist"

for f in "${SPEC_FILES[@]}"; do
  echo "Updating generated otel-library-allowlist query in ${f}..."

  if ! grep -q 'as c2_non_allowlist$' "$f"; then
    echo "No 'c2_non_allowlist' query found in ${f}, skipping." >&2
    exit 1
  fi

  awk -v q="  - query: ${query}" '
    /as c2_non_allowlist$/ { print q; next }
    { print }
  ' "$f" > "$f.tmp" && mv "$f.tmp" "$f"
done
