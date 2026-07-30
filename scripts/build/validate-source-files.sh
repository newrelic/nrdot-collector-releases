#!/bin/bash
# Copyright New Relic, Inc. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

# Simple script to validate whether a distribution's OCB-generated source files exist.
set -e

# default values
fips=false
distributions=""

while getopts d:f: flag
do
    case "${flag}" in
        d) distributions=${OPTARG};;
        f) fips=${OPTARG};;
        *) exit 1;;
    esac
done

if [ -z "${distributions}" ]; then
    echo "Distribution not provided. Please provide a distribution with -d."
    exit 1
fi

files=(
    "components.go" "go.mod" "go.sum" "main_others.go"
    "main_windows.go" "main.go" # build.log excluded as it is not cached
)
if [ ${fips} = true ]; then
    files+=("fips.go")
fi

overall_exit=0

for distro in $(echo "$distributions" | tr "," "\n"); do
    path="distributions/${distro}/_build"
    if [ ${fips} = true ]; then
        path="${path}-fips"
    fi
    if [ ! -d "$path" ]; then
        echo "❌ $path not found!"
        overall_exit=1
        continue
    fi

    missing_files=()
    for file in "${files[@]}"; do
        if [ ! -f "${path}/${file}" ]; then
            missing_files+=("$file")
        else
            echo "Found: ${path}/${file}"
        fi
    done

    if [ ${#missing_files[@]} -eq 0 ]; then
        echo "✅ ${distro}: All source files found!"
    else
        echo "❌ ${distro}: files not found: ${missing_files[*]}"
        overall_exit=1
    fi
done

exit ${overall_exit}
