#!/bin/bash

REPO_DIR="$( cd "$(dirname "$( dirname "${BASH_SOURCE[0]}" )")" &> /dev/null && pwd )"
DIRECTORY="_build"

# default values
validate=true

while getopts d:s:b:f: flag
do
    case "${flag}" in
        d) distributions=${OPTARG};;
        l) validate=${OPTARG};;
        *) exit 1;;
    esac
done

if [[ -z $distributions ]]; then
    echo "List of distributions to build not provided. Use '-d' to specify the names of the distributions to build. Ex.:"
    echo "$0 -d nrdot-collector-k8s"
    exit 1
fi

echo "Skipping the compilation, we'll only generate the sources."

for distribution in $(echo "$distributions" | tr "," "\n")
do
    pushd "${REPO_DIR}/distributions/${distribution}" > /dev/null || exit
    mkdir -p _build-fips

    echo "Building: $distribution-fips"

    if [ -d "$DIRECTORY" ]; then
        echo "Copying _build into _build-fips."
        cp -R _build/. ./_build-fips

        echo "Copying fips.go into _build-fips."
        cp ../../fips/fips.go ./_build-fips
        echo "Compiling binary."
        echo "✅ SUCCESS: distribution '${distribution}-fips' built."
    else
        echo "❌ ERROR: failed to build the distribution '${distribution}-fips'."
        echo "🪵 Build logs for '${distribution}-fips'"
        echo "----------------------"
        cat _build-fips/build.log
        echo "----------------------"
        exit 1
    fi

    popd > /dev/null || exit
done