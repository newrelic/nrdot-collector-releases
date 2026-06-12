#!/usr/bin/env bash
# Copyright New Relic, Inc. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

# Validate that distributions/<distro>/component-inventory.yaml lists exactly
# the components present in distributions/<distro>/manifest.yaml.
#
# Usage: scripts/validate-component-inventory.sh <distribution>
#
# Exits 0 if the inventory file does not exist (the check is opt-in per distro).
# Exits 0 if every component category matches in both directions.
# Exits 1 with a per-category diff when there is drift.

set -euo pipefail

if [[ $# -ne 1 ]]; then
    echo "usage: $0 <distribution>" >&2
    exit 2
fi

DISTRO="$1"
SRC_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
DIR="${SRC_ROOT}/distributions/${DISTRO}"
MANIFEST="${DIR}/manifest.yaml"
INVENTORY="${DIR}/component-inventory.yaml"

if [[ ! -f "$INVENTORY" ]]; then
    echo "ℹ️  ${DISTRO}: no component-inventory.yaml, skipping check"
    exit 0
fi

if [[ ! -f "$MANIFEST" ]]; then
    echo "❌ ${DISTRO}: manifest.yaml not found at ${MANIFEST}" >&2
    exit 1
fi

# plural manifest section -> singular path segment used in gomod URLs
CATEGORIES=(
    "receivers:receiver"
    "processors:processor"
    "exporters:exporter"
    "connectors:connector"
    "extensions:extension"
    "providers:provider"
)

failed=0
for entry in "${CATEGORIES[@]}"; do
    plural="${entry%%:*}"
    singular="${entry##*:}"

    manifest_components=$(yq -r "
        .${plural}[]?.gomod
        | sub(\" v[0-9].*$\"; \"\")
        | sub(\".*/${singular}/\"; \"\")
    " "$MANIFEST" | sort -u)

    inventory_components=$(yq -r "
        .${plural} // {} | keys | .[]
    " "$INVENTORY" | sort -u)

    if ! diff_output=$(diff <(echo "$manifest_components") <(echo "$inventory_components")); then
        echo "❌ ${DISTRO} ${plural}: manifest and inventory disagree"
        echo "$diff_output" | sed 's/^/   /'
        failed=1
    fi
done

if [[ $failed -eq 1 ]]; then
    echo
    echo "   Manifest:  ${MANIFEST}"
    echo "   Inventory: ${INVENTORY}"
    echo "   '<' lines are only in the manifest; '>' lines are only in the inventory."
    exit 1
fi

echo "✅ ${DISTRO}: component inventory matches manifest"
