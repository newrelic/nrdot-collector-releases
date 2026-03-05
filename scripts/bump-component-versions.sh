#!/bin/bash
# Copyright New Relic, Inc. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

set -e

GO=''

while getopts d:g: flag
do
    case "${flag}" in
        g) GO=${OPTARG};;
        *) exit 1;;
    esac
done

[[ -n "$GO" ]] || GO='go'

# Store the current directory
ORIGINAL_DIR=$(pwd)

# Fetch the latest nrdot-collector-components version and its OTel dependency versions.
fetch_nrdot_versions() {
    local nrdot_module="github.com/newrelic/nrdot-collector-components/exporter/nopexporter"

    echo "Fetching latest version of nrdot-collector-components nrdot..." >&2

    local latest_version
    latest_version=$(${GO} list -m -versions "$nrdot_module" 2>/dev/null | awk '{print $NF}')

    if [[ -z "$latest_version" ]]; then
        echo "Warning: No versions found for $nrdot_module" >&2
        return 1
    fi

    echo "Latest nrdot version: ${latest_version}" >&2

    # Download the specific version and get its dependencies
    echo "Downloading nrdot@${latest_version} and extracting dependencies..." >&2
    ${GO} get "${nrdot_module}@${latest_version}" >/dev/null 2>&1

    # Build a temporary module to resolve the full dependency graph
    TEMP_DIR=$(mktemp -d)
    cd "$TEMP_DIR" || {
        echo "Warning: Could not create temp directory" >&2
        return 1
    }
    $GO mod init temp 2>/dev/null
    $GO get ${nrdot_module}@${latest_version} 2>/dev/null

    # Extract collector core stable version (v1.x.x)
    local core_stable
    core_stable=$(${GO} list -m all 2>/dev/null | \
        grep "^go.opentelemetry.io/collector/" | \
        awk '{print $2}' | \
        grep "^v1\." | \
        sort -V | tail -1)

    # Extract collector core beta version (v0.x.x)
    local core_beta
    core_beta=$(${GO} list -m all 2>/dev/null | \
        grep "^go.opentelemetry.io/collector/" | \
        awk '{print $2}' | \
        grep "^v0\." | \
        sort -V | tail -1)

    # Clean up temp directory and return to original directory
    cd "$ORIGINAL_DIR" || exit 1
    rm -rf "$TEMP_DIR"

    # Find the highest contrib patch whose minor version matches core_beta.
    # Contrib modules track the same minor as core beta (e.g., v0.147.x).
    local contrib_beta=""
    if [[ -n "$core_beta" ]]; then
        local core_minor
        core_minor=$(echo "$core_beta" | awk -F'.' '{print $1"."$2}')
        contrib_beta=$(${GO} list -m -versions \
            "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/filelogreceiver" \
            2>/dev/null | tr ' ' '\n' | grep "^${core_minor}\." | sort -V | tail -1)
    fi

    # Output as JSON
    if [[ -n "$core_stable" ]] || [[ -n "$core_beta" ]]; then
        echo "{"
        echo "  \"nrdotVersion\": \"$latest_version\","
        echo "  \"coreStable\": \"${core_stable:-none}\","
        echo "  \"coreBeta\": \"${core_beta:-none}\","
        echo "  \"contribBeta\": \"${contrib_beta:-none}\""
        echo "}"
    else
        echo "Warning: Could not extract collector versions from nrdot dependencies" >&2
        return 1
    fi
}

# Fetch nrdot versions and store them
NRDOT_VERSIONS=$(fetch_nrdot_versions)
NRDOT_STATUS=$?

# Extract individual values from the nrdot JSON if successful
NRDOT_FLAGS=()
if [[ $NRDOT_STATUS -eq 0 ]]; then
    NRDOT_VERSION=$(echo "$NRDOT_VERSIONS" | jq -r '.nrdotVersion // ""')
    CORE_STABLE=$(echo "$NRDOT_VERSIONS" | jq -r '.coreStable // ""')
    CORE_BETA=$(echo "$NRDOT_VERSIONS" | jq -r '.coreBeta // ""')
    CONTRIB_BETA=$(echo "$NRDOT_VERSIONS" | jq -r '.contribBeta // ""')

    [[ -n "$NRDOT_VERSION" ]] && NRDOT_FLAGS+=(--nrdot-version "$NRDOT_VERSION")
    [[ -n "$CORE_STABLE" ]]   && NRDOT_FLAGS+=(--core-stable "$CORE_STABLE")
    [[ -n "$CORE_BETA" ]]     && NRDOT_FLAGS+=(--core-beta "$CORE_BETA")
    [[ -n "$CONTRIB_BETA" ]]  && NRDOT_FLAGS+=(--contrib-beta "$CONTRIB_BETA")
fi

# Change to the CLI tool directory
cd "$(dirname "$0")/../cmd/nrdot-collector-builder" || exit 1

# Run the manifest update with nrdot flags
OUTPUT=$(${GO} run main.go manifest update --json --config "../../distributions/*/manifest.yaml" "${NRDOT_FLAGS[@]}")

# Return to the original directory
cd "$ORIGINAL_DIR" || exit 1

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

# Extract the current beta core version
current_beta_core=$(echo "$OUTPUT" | jq -r '.currentVersions.betaCoreVersion')
current_beta_core=${current_beta_core#v}
escaped_current_beta_core=${current_beta_core//./\\.}
next_beta_core=$(echo "$OUTPUT" | jq -r '.nextVersions.betaCoreVersion')
next_beta_core=${next_beta_core#v}

# If the current beta core version is not equal to the next beta core version, update the Makefile
if [[ "$current_beta_core" != "$next_beta_core" ]]; then
  # Update Makefile OCB version
  sed_inplace "s/OTELCOL_BUILDER_VERSION ?= $escaped_current_beta_core/OTELCOL_BUILDER_VERSION ?= $next_beta_core/" Makefile
fi

# Output the result (nrdot info is already included if it was fetched successfully)
echo "$OUTPUT"
