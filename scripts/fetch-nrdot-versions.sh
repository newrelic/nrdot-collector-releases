#!/bin/bash
# Copyright New Relic, Inc. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

set -e

# Allow GO override (e.g. -g /path/to/go); default to 'go'
while getopts "g:" opt; do
    case $opt in
        g) GO=${OPTARG};;
    esac
done
[[ -n "$GO" ]] || GO='go'

# Days the nrdot-collector-components and newrelic-forks repos may drift before CI fails / notice is sent
DRIFT_GRACE_PERIOD_DAYS=14

get_latest_version() {
    local module=$1
    local version=$(${GO} list -m -versions "$module" 2>/dev/null | tr ' ' '\n' | sort -V | tail -1)
    # Go list can return module name if module not found
    if [[ -z "$version" || $version == $module ]]; then
        echo "Warning: No versions found for $module" >&2
        return 1
    fi
    echo $version
}

get_release_time_elapsed_days() {
    local module=$1
    local version=$2
    local release_time
    release_time=$(${GO} list -m -json "${module}@${version}" 2>/dev/null | jq -r '.Time')
    if [[ "$OSTYPE" == "darwin"* ]]; then
        awk "BEGIN {printf \"%.2f\", ( $(date -u +%s) - $(date -u -j -f "%Y-%m-%dT%H:%M:%SZ" "$release_time" +%s) ) / 86400}"
    else
        awk "BEGIN {printf \"%.2f\", ( $(date -u +%s) - $(date -u -d "$release_time" +%s) ) / 86400}"
    fi
}

# Fetch latest release version version of nrdot-collector-components repo
nrdot_module="github.com/newrelic/nrdot-collector-components/exporter/nopexporter"
echo "Fetching latest version of nrdot-collector-components..." >&2
nrdot_latest=$(get_latest_version "$nrdot_module")
echo "Latest nrdot-collector-components version: ${nrdot_latest}" >&2
nrdot_minor=$(echo "$nrdot_latest" | awk -F'.' '{print $2}')

# Fetch latest release version of newrelic-forks contrib repo
nr_fork_module="github.com/newrelic-forks/opentelemetry-collector-contrib/receiver/nrsqlserverreceiver"
echo "Fetching latest version of newrelic-forks contrib..." >&2
nr_fork_latest=$(get_latest_version "$nr_fork_module")
echo "Latest newrelic-forks contrib version: $nr_fork_latest" >&2
nr_fork_latest_minor=$(echo "$nr_fork_latest" | awk -F'.' '{print $2}')

# Calculate bi-directional version drift
echo "Fetching version drift between nrdot-collector-components and newrelic-forks contrib..." >&2
if [[ "$nrdot_minor" -eq "$nr_fork_latest_minor" ]]; then
    days_drifted=0.00
elif [[ "$nrdot_minor" -gt "$nr_fork_latest_minor" ]]; then
    days_drifted=$(get_release_time_elapsed_days "$nrdot_module" "v0.$((nr_fork_latest_minor + 1)).0")
else
    days_drifted=$(get_release_time_elapsed_days "$nr_fork_module" "v0.$((nrdot_minor + 1)).0")
fi
echo "Days drifted: $days_drifted" >&2
if (( $(echo "$days_drifted > $DRIFT_GRACE_PERIOD_DAYS" | bc -l) )); then
    drift_grace_period_exceeded=true
    echo "⚠️ Warning: Grace period exceeded!" >&2
else
    drift_grace_period_exceeded=false
fi

# Fetch latest newrelic-forks contrib patch version matching latest NRDOT minor
echo "Fetching latest patch version of newrelic-forks contrib matching nrdot-collector-components..." >&2
nr_fork_matching=$(${GO} list -m -versions \
    "$nr_fork_module" \
    2>/dev/null | tr ' ' '\n' | grep "^v0\.${nrdot_minor}\." | sort -V | tail -1)
if [[ -z "$nr_fork_matching" ]]; then
    echo "⚠️ Warning: No nr-forks contrib version found for minor v0.$nrdot_minor." >&2
else
    echo "Matching minor version found for nr-forks contrib: $nr_fork_matching" >&2
fi

# Download the specific version and get its dependencies
echo "Downloading nrdot@${nrdot_latest} and extracting dependencies..." >&2

# Build a temporary module to resolve the full dependency graph
# Store the current directory
TEMP_DIR=$(mktemp -d)
trap "rm -rf '$TEMP_DIR'" EXIT
pushd "$TEMP_DIR" > /dev/null || {
    echo "Warning: Could not create temp directory" >&2
    exit 1
}
$GO mod init temp 2>/dev/null
$GO get ${nrdot_module}@${nrdot_latest} 2>/dev/null

# Extract collector core stable version (v1.x.x)
core_stable=$(${GO} list -m all 2>/dev/null | \
    grep "^go.opentelemetry.io/collector/" | \
    awk '{print $2}' | \
    grep "^v1\." | \
    sort -V | tail -1)

# Extract collector core beta version (v0.x.x)
core_beta=$(${GO} list -m all 2>/dev/null | \
    grep "^go.opentelemetry.io/collector/" | \
    awk '{print $2}' | \
    grep "^v0\." | \
    sort -V | tail -1)

popd > /dev/null || exit 1

# Find the highest contrib patch whose minor version matches core_beta.
# Contrib modules track the same minor as core beta (e.g., v0.147.x).
contrib_beta=""
if [[ -n "$core_beta" ]]; then
    core_minor=$(echo "$core_beta" | awk -F'.' '{print $1"."$2}')
    contrib_beta=$(${GO} list -m -versions \
        "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/filelogreceiver" \
        2>/dev/null | tr ' ' '\n' | grep "^${core_minor}\." | sort -V | tail -1)
fi

# Output as $GITHUB_OUTPUT-friendly text
if [[ -n "$core_stable" ]] || [[ -n "$core_beta" ]]; then
    echo "nrdot_latest=${nrdot_latest}"
    echo "days_drifted=${days_drifted}"
    echo "drift_grace_period_exceeded=${drift_grace_period_exceeded}"
    echo "nr_fork_matching=${nr_fork_matching:-none}"
    echo "nr_fork_latest=${nr_fork_latest:-none}"
    echo "core_stable=${core_stable:-none}"
    echo "core_beta=${core_beta:-none}"
    echo "contrib_beta=${contrib_beta:-none}"
else
    echo "⚠️ Warning: Could not extract collector versions from nrdot dependencies" >&2
    exit 1
fi