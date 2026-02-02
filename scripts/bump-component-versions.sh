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

# Function to fetch the latest version of nrdot-collector-components nrdot
fetch_nrdot_versions() {
    local nrdot_module="github.com/newrelic/nrdot-collector-components/exporter/nrdot"

    echo "Fetching latest version of nrdot-collector-components nrdot..." >&2

    local latest_version
    latest_version=$(${GO} list -m -versions "$nrdot_module" 2>/dev/null | awk '{print $NF}')

    if [[ -z "$latest_version" ]]; then
        echo "Warning: No versions found for $nrdot_module" >&2
        return 1
    fi

    echo "Latest nrdot version: $latest_version" >&2

    # Download the specific version and get its dependencies
    echo "Downloading nrdot@$latest_version and extracting dependencies..." >&2
    ${GO} get "${nrdot_module}@${latest_version}" >/dev/null 2>&1

    nrdot_info=$($GO list -m -json github.com/newrelic/nrdot-collector-components/exporter/nrdot@${latest_version} 2>/dev/null)

    # Get the dependency graph for nrdot
    TEMP_DIR=$(mktemp -d)
    cd "$TEMP_DIR" || {
        echo "Warning: Could not create temp directory" >&2
        return 1
    }
    $GO mod init temp 2>/dev/null
    $GO get github.com/newrelic/nrdot-collector-components/exporter/nopexporter@${latest_version} 2>/dev/null

    # Extract collector core version (stable v1.x.x)
    local core_stable
    core_stable=$(${GO} list -m -json all 2>/dev/null | \
        grep "go.opentelemetry.io/collector " | \
        awk '{print $2}' | \
        grep "^v1\." | \
        head -1)

    # Extract collector core version (beta v0.x.x)
    local contrib_beta
    contrib_beta=$(${GO} list -m -json all 2>/dev/null | \
        grep "github.com/open-telemetry/opentelemetry-collector-contrib " | \
        awk '{print $2}' | \
        grep "^v0\." | \
        head -1)

    # Clean up temp directory and return to original directory
    cd "$ORIGINAL_DIR" || exit 1
    rm -rf "$TEMP_DIR"

    # Output as JSON
    if [[ -n "$core_stable" ]] || [[ -n "$contrib_beta" ]]; then
        echo "{"
        echo "  \"nrdotVersion\": \"$latest_version\","
        echo "  \"coreStable\": \"${core_stable:-none}\","
        echo "  \"contribBeta\": \"${contrib_beta:-none}\""
        echo "}"
    else
        echo "Warning: Could not extract collector versions from nrdot dependencies" >&2
        return 1
    fi

    return 0
}

# Fetch nrdot versions and store them
NRDOT_VERSIONS=$(fetch_nrdot_versions)
nrdot_STATUS=$?

# Extract individual values from the nrdot JSON if successful
NRDOT_FLAGS=""
if [[ $NRDOT_STATUS -eq 0 ]]; then
    NRDOT_VERSION=$(echo "$NRDOT_VERSIONS" | jq -r '.nrdotVersion // ""')
    CORE_STABLE=$(echo "$NRDOT_VERSIONS" | jq -r '.coreStable // ""')
    CONTRIB_BETA=$(echo "$NRDOT_VERSIONS" | jq -r '.contribBeta // ""')

    # Build flags string
    [[ -n "$NRDOT_VERSION" ]] && NRDOT_FLAGS="$NRDOT_FLAGS --nrdot-version=\"$NRDOT_VERSION\""
    [[ -n "$CORE_STABLE" ]] && NRDOT_FLAGS="$NRDOT_FLAGS --core-stable=\"$CORE_STABLE\""
    [[ -n "$CONTRIB_BETA" ]] && NRDOT_FLAGS="$NRDOT_FLAGS --contrib-beta=\"$CONTRIB_BETA\""
fi

# Change to the CLI tool directory
cd "$(dirname "$0")/../cmd/nrdot-collector-builder" || exit 1

# Run the manifest update with nrdot flags
OUTPUT=$(${GO} run main.go manifest update --json --config "../../distributions/*/manifest.yaml" $NRDOT_FLAGS)

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

#  If the current beta core version is not equal to the next beta core version, update the Makefile
if [[ "$current_beta_core" != "$next_beta_core" ]]; then
  # Update Makefile OCB version
  sed_inplace "s/OTELCOL_BUILDER_VERSION ?= $escaped_current_beta_core/OTELCOL_BUILDER_VERSION ?= $next_beta_core/" Makefile
fi

# Output the result (nrdot info is already included if it was fetched successfully)
echo "$OUTPUT"
