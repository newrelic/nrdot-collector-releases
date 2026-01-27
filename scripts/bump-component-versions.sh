#!/bin/bash
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

# Function to fetch the latest version of nrdot-collector-components nopexporter
fetch_nopexporter_versions() {
    local nopexporter_module="github.com/newrelic/nrdot-collector-components/exporter/nopexporter"

    echo "Fetching latest version of nrdot-collector-components nopexporter..." >&2

    local latest_version
    latest_version=$(${GO} list -m -versions "$nopexporter_module" 2>/dev/null | awk '{print $NF}')

    if [[ -z "$latest_version" ]]; then
        echo "Warning: No versions found for $nopexporter_module" >&2
        return 1
    fi

    echo "Latest nopexporter version: $latest_version" >&2

    # Download the specific version and get its dependencies
    echo "Downloading nopexporter@$latest_version and extracting dependencies..." >&2
    ${GO} get "${nopexporter_module}@${latest_version}" >/dev/null 2>&1 

    nopexporter_info=$($GO list -m -json github.com/newrelic/nrdot-collector-components/exporter/nopexporter@${latest_version} 2>/dev/null)

    # Get the dependency graph for nopexporter
    TEMP_DIR=$(mktemp -d)
    cd "$TEMP_DIR" || {
        echo "Warning: Could not create temp directory" >&2
        return 1
    }
    $GO mod init temp 2>/dev/null
    $GO get github.com/newrelic/nrdot-collector-components/exporter/nopexporter@${latest_version} 2>/dev/null

    # Extract collector core version (stable v1.x.x)
    local collector_core_stable
    collector_core_stable=$(${GO} list -m -json all 2>/dev/null | \
        grep "go.opentelemetry.io/collector " | \
        awk '{print $2}' | \
        grep "^v1\." | \
        head -1)

    # Extract collector core version (beta v0.x.x)
    local collector_contrib_beta
    collector_contrib_beta=$(${GO} list -m -json all 2>/dev/null | \
        grep "github.com/open-telemetry/opentelemetry-collector-contrib " | \
        awk '{print $2}' | \
        grep "^v0\." | \
        head -1)

    # Clean up temp directory and return to original directory
    cd "$ORIGINAL_DIR" || exit 1
    rm -rf "$TEMP_DIR"

    # Output as JSON
    if [[ -n "$collector_core_stable" ]] || [[ -n "$collector_contrib_beta" ]]; then
        echo "{"
        echo "  \"nopexporterVersion\": \"$latest_version\","
        echo "  \"collectorCoreStable\": \"${collector_core_stable:-none}\","
        echo "  \"collectorContribBeta\": \"${collector_contrib_beta:-none}\""
        echo "}"
    else
        echo "Warning: Could not extract collector versions from nopexporter dependencies" >&2
        return 1
    fi

    return 0
}

# Fetch nopexporter versions and store them
NOPEXPORTER_VERSIONS=$(fetch_nopexporter_versions)
NOPEXPORTER_STATUS=$?

# Extract individual values from the nopexporter JSON if successful
NOPEXPORTER_FLAGS=""
if [[ $NOPEXPORTER_STATUS -eq 0 ]]; then
    NOPEXPORTER_VERSION=$(echo "$NOPEXPORTER_VERSIONS" | jq -r '.nopexporterVersion // ""')
    COLLECTOR_CORE_STABLE=$(echo "$NOPEXPORTER_VERSIONS" | jq -r '.collectorCoreStable // ""')
    COLLECTOR_CONTRIB_BETA=$(echo "$NOPEXPORTER_VERSIONS" | jq -r '.collectorContribBeta // ""')

    # Build flags string
    [[ -n "$NOPEXPORTER_VERSION" ]] && NOPEXPORTER_FLAGS="$NOPEXPORTER_FLAGS --nopexporter-version=\"$NOPEXPORTER_VERSION\""
    [[ -n "$COLLECTOR_CORE_STABLE" ]] && NOPEXPORTER_FLAGS="$NOPEXPORTER_FLAGS --collector-core-stable=\"$COLLECTOR_CORE_STABLE\""
    [[ -n "$COLLECTOR_CONTRIB_BETA" ]] && NOPEXPORTER_FLAGS="$NOPEXPORTER_FLAGS --collector-contrib-beta=\"$COLLECTOR_CONTRIB_BETA\""
fi

# Change to the CLI tool directory
cd "$(dirname "$0")/../cmd/nrdot-collector-builder" || exit 1

# Run the manifest update with nopexporter flags
OUTPUT=$(${GO} run main.go manifest update --json --config "../../distributions/*/manifest.yaml" $NOPEXPORTER_FLAGS)

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

# Output the result (nopexporter info is already included if it was fetched successfully)
echo "$OUTPUT"
