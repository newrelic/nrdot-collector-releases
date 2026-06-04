#!/bin/bash
# Copyright New Relic, Inc. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

# Script to validate a goreleaser distribution's dist file
set -e

split=false
goos=""

while getopts d:sg: flag
do
    case "${flag}" in
        d) distro=${OPTARG};;
        s) split=true;;
        g) goos=${OPTARG};;
        *) exit 1;;
    esac
done

if [ -z ${distro} ]; then
    echo "Distribution not provided. Please provide a distribution with -d."
    exit 1
fi

if [ "${split}" = true ] && [ -z "${goos}" ]; then
    echo "Split mode requires GOOS to be provided with -g."
    exit 1
fi

cd "distributions/${distro}"
if [ ! -d "dist" ]; then
    echo "❌ dist directory not found!"
    exit 1
fi

# Set metadata directory based on mode
if [ "${split}" = true ]; then
    # In split mode, metadata files are in dist/<goos>/
    metadata_dir="dist/${goos}"
else
    # In normal/merge mode, metadata files are at dist root
    metadata_dir="dist"
fi

echo "📋 Verifying metadata files exist..."
files=("${metadata_dir}/artifacts.json" "${metadata_dir}/metadata.json")
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
binaries=$( jq -r '.[] | select(.type == "Binary") | .path' "${metadata_dir}/artifacts.json" )
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
artifacts=$( jq -r '.[] | select(.type == "Archive" or .type == "Linux Package") | .path' "${metadata_dir}/artifacts.json" )
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
        # Skip checksum validation in split mode (checksums are generated during merge)
        if [ "${split}" = true ]; then
            echo "  ⏭️ Skipping checksum validation (split mode)"
            continue
        fi
        # Search for the corresponding checksum file and verify it exists
        sum_file=$( jq -r ".[] | select(.type == \"Checksum\" and .extra.ChecksumOf == \"${artifact}\") | .path" "${metadata_dir}/artifacts.json" )
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
