#!/bin/bash
# Copyright New Relic, Inc. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

# Script to check if / how long nrdot-collector-components and newrelic-forks repos have been out-of-sync.
set -e

GO=''

while getopts g: flag
do
    case "${flag}" in
        g) GO=${OPTARG};;
        *) exit 1;;
    esac
done

[[ -n "$GO" ]] || GO='go'

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

nrdot_module="github.com/newrelic/nrdot-collector-components/exporter/nopexporter"
nrdot_version=$(${GO} list -m -versions "$nrdot_module" 2>/dev/null | tr ' ' '\n' | sort -V | tail -1)
if [[ -z "$nrdot_version" ]]; then
    echo "Warning: No versions found for $nrdot_module" >&2
    exit 1
fi

nr_forks_module="github.com/newrelic-forks/opentelemetry-collector-contrib/receiver/nrsqlserverreceiver"
nr_forks_version=$(${GO} list -m -versions "$nr_forks_module" 2>/dev/null | tr ' ' '\n' | sort -V | tail -1)
if [[ -z "$nr_forks_version" ]]; then
    echo "Warning: No versions found for $nr_forks_module" >&2
    exit 1
fi

nrdot_minor=$(echo "$nrdot_version" | awk -F'.' '{print $2}')
nr_forks_minor=$(echo "$nr_forks_version" | awk -F'.' '{print $2}')

# Get days out of sync
if [[ "$nrdot_minor" -eq "$nr_forks_minor" ]]; then
    days_drifted="0.00"
elif [[ "$nr_forks_minor" -gt "$nrdot_minor" ]]; then
    days_drifted=$(get_release_time_elapsed_days "$nrdot_module" "$nrdot_version")
else
    days_drifted=$(get_release_time_elapsed_days "$nr_forks_module" "$nr_forks_version")
fi

echo "Latest NRDOT version: ${nrdot_version}" >&2
echo "Latest NR fork contrib version: ${nr_forks_version}" >&2
echo "NR fork contrib days drifted: ${days_drifted}" >&2

echo "{"
echo "  \"nrdotVersion\": \"${nrdot_version}\","
echo "  \"nrForkContribVersion\": \"${nr_forks_version}\","
echo "  \"daysDrifted\": ${days_drifted}"
echo "}"