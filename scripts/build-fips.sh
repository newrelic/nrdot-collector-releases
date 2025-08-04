#!/bin/bash

REPO_DIR="$( cd "$(dirname "$( dirname "${BASH_SOURCE[0]}" )")" &> /dev/null && pwd )"
BUILDER=''

# default values
validate=true

while getopts d:s:b:f: flag
do
    case "${flag}" in
        d) distributions=${OPTARG};;
        l) validate=${OPTARG};;
        b) BUILDER=${OPTARG};;
        *) exit 1;;
    esac
done

[[ -n "$BUILDER" ]] || BUILDER='ocb'

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
    echo "Using Builder: $(command -v "$BUILDER")"
    echo "Using Go: $(command -v go)"

    if "$BUILDER" --skip-compilation="true" --config manifest-fips.yaml > _build-fips/build.log 2>&1; then
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