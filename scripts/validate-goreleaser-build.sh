#!/bin/bash
# Copyright New Relic, Inc. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

# Script to validate a goreleaser distribution's dist file
set -e

while getopts d: flag
do
    case "${flag}" in
        d) distro=${OPTARG};;
        *) exit 1;;
    esac
done

if [ -z ${distro} ]; then
    echo "Distribution not provided. Please provide a distribution with -d."
    exit 1
fi

cd "distributions/${distro}"
if [ ! -d "dist" ]; then
    echo "❌ dist directory not found!"
    exit 1
fi

echo "📋 Verifying metadata files exist..."
files=("dist/artifacts.json" "dist/metadata.json")
missing_files=()
for file in "${files[@]}"; do
    if [ ! -f "$file" ]; then
        missing_files+=("$file")
    else
        echo "Found: $file"
    fi
done
if [ ${#missing_files[@]} -ne 0 ]; then
    echo "❌ files not found: ${missing_files[*]}"
    exit 1
else
    echo "✅ All common build files found!"
fi

echo "📋 Verifying Binaries exist..."
binaries=$( jq -r '.[] | select(.type == "Binary") | .path' dist/artifacts.json )
if [ -z "${binaries}" ]; then
    echo "❌ No binaries found in artifacts.json"
    exit 1
else
    for binary in $binaries; do
        if [ ! -f "${binary}" ]; then
            echo "❌ ${binary} not found!"
            exit 1
        else
            echo "Found: ${binary}"
        fi
    done
fi
echo "✅ All binaries found!"

echo "📋 Validating Archives and Packages..."
artifacts=$( jq -r '.[] | select(.type == "Archive" or .type == "Linux Package") | .path' dist/artifacts.json )
if [ -z "${artifacts}" ]; then
    echo "⚠️ No archives or packages found in artifacts.json"
else
    for artifact in $artifacts; do
        echo "Validating ${artifact}"
        # Verify the artifact file exists
        if [ ! -f "${artifact}" ]; then
            echo "❌ ${artifact} not found!"
            exit 1
        else
            echo "  Found artifact: ${artifact}"
        fi
        # Search for the corresponding checksum file and verify it exists
        sum_file=$( jq -r ".[] | select(.type == \"Checksum\" and .extra.ChecksumOf == \"${artifact}\") | .path" dist/artifacts.json )
        if [ -z "${sum_file}" ]; then
            echo "❌ Checksum not defined for ${artifact} in artifacts.json"
            exit 1
        fi
        if [ ! -f "${sum_file}" ]; then
            echo "❌ ${sum_file} not found!"
            exit 1
        else
            echo "  Found checksum: ${sum_file}"
        fi
        # Compare checksums to ensure file integrity
        artifact_sum=$(sha256sum ${artifact} | cut -d' ' -f1)
        expected_sum=$(cat ${sum_file})
        if [ "${artifact_sum}" != "${expected_sum}" ]; then
            echo "❌ Checksums do not match!"
            echo "Checksum: ${artifact_sum}"
            echo "Expected: ${expected_sum}"
            exit 1
        else
            echo "  Checksum validated"
        fi
    done
    echo "✅ Archives and Packages validated!"
fi