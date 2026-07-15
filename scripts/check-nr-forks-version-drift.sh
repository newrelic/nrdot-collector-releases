#!/bin/bash
# Copyright New Relic, Inc. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

# Script to check length of version drift with fork repo

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

nrdot_module="github.com/newrelic/nrdot-collector-components/exporter/nopexporter"
nrdot_version=$(${GO} list -m -versions "$nrdot_module" 2>/dev/null | awk '{print $NF}')
if [[ -z "$nrdot_version" ]]; then
    echo "Warning: No versions found for $nrdot_module" >&2
    exit 1
fi

# Fetch release time (in days to 2-decimal precision)
release_time=$(${GO} list -m -json "${nrdot_module}@${nrdot_version}" 2>/dev/null | jq -r '.Time')
if [[ "$OSTYPE" == "darwin"* ]]; then
    days_drifted=$(awk "BEGIN {printf \"%.2f\", ( $(date -u +%s) - $(date -u -j -f "%Y-%m-%dT%H:%M:%SZ" "$release_time" +%s) ) / 86400}")
else
    days_drifted=$(awk "BEGIN {printf \"%.2f\", ( $(date -u +%s) - $(date -u -d "$release_time" +%s) ) / 86400}")
fi

latest_nr_forks_version=$(${GO} list -m -versions \
    "github.com/newrelic-forks/opentelemetry-collector-contrib/receiver/nrsqlserverreceiver" \
    2>/dev/null | tr ' ' '\n' | sort -V | tail -1)

echo "NRDOT version: ${nrdot_version}" >&2
echo "NR fork contrib version: ${latest_nr_forks_version}" >&2
echo "NR fork contrib days drifted: ${days_drifted}" >&2

echo "{"
echo "  \"nrdotVersion\": \"${nrdot_version}\","
echo "  \"nrForkContribVersion\": \"${latest_nr_forks_version}\","
echo "  \"nrForkContribDaysDrifted\": ${days_drifted}"
echo "}"