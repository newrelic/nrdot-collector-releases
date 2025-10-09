#!/bin/bash
# Simple script to validate whether a distribution's OCB-generated source files exist.
set -e

# default values
fips=false

while getopts d:f: flag
do
    case "${flag}" in
        d) distro=${OPTARG};;
        f) fips=${OPTARG};;
        *) exit 1;;
    esac
done

if [ -z ${distro} ]; then
    echo "Distribution not provided. Please provide a distribution with -d."
    exit 1
fi

path="distributions/${distro}/_build"
if [ ${fips} = true ]; then
    path="${path}-fips"
fi
if [ ! -d "$path" ]; then
    echo "❌ $path not found!"
    exit 1
fi
cd "$path"

files=(
    "build.log" "components.go" "go.mod" "go.sum"
    "main_others.go" "main_windows.go" "main.go"
)
if [ ${fips} = true ]; then
    files+=("fips.go")
fi
missing_files=()

for file in "${files[@]}"; do
    if [ ! -f "$file" ]; then
        missing_files+=("$file")
    else
        echo "Found: $file"
    fi
done

if [ ${#missing_files[@]} -eq 0 ]; then
    echo "✅ All source files found!"
else
    echo "❌ files not found: ${missing_files[*]}"
    exit 1
fi