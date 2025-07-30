#!/bin/bash

REPO_DIR="$( cd "$(dirname "$( dirname "${BASH_SOURCE[0]}" )")" &> /dev/null && pwd )"
BUILDER=''

# default values
skipcompilation=false
validate=true
fips=false

while getopts d:s:b:f: flag
do
    case "${flag}" in
        d) distributions=${OPTARG};;
        s) skipcompilation=${OPTARG};;
        l) validate=${OPTARG};;
        b) BUILDER=${OPTARG};;
        f) fips=${OPTARG};;
        *) exit 1;;
    esac
done

[[ -n "$BUILDER" ]] || BUILDER='ocb'

if [[ -z $distributions ]]; then
    echo "List of distributions to build not provided. Use '-d' to specify the names of the distributions to build. Ex.:"
    echo "$0 -d nrdot-collector-k8s"
    exit 1
fi

if [[ "$skipcompilation" = true ]]; then
    echo "Skipping the compilation, we'll only generate the sources."
fi

echo "Distributions to build: $distributions";

for distribution in $(echo "$distributions" | tr "," "\n")
do
    pushd "${REPO_DIR}/distributions/${distribution}" > /dev/null || exit
    mkdir -p _build

    echo "Building: $distribution"
    echo "Using Builder: $(command -v "$BUILDER")"
    echo "Using Go: $(command -v go)"

    if "$BUILDER" --skip-compilation="${skipcompilation}" --config manifest.yaml > _build/build.log 2>&1; then
        echo "✅ SUCCESS: distribution '${distribution}' built."
    else
        echo "❌ ERROR: failed to build the distribution '${distribution}'."
        echo "🪵 Build logs for '${distribution}'"
        echo "----------------------"
        cat _build/build.log
        echo "----------------------"
        exit 1
    fi

    if [[ "$skipcompilation" = true && "$fips" = true ]]; then
      echo "Copying fips.go into _build."
      cp ../fips.go ./_build
      echo "Compiling binary."
      goreleaser build --snapshot --clean
    fi

    popd > /dev/null || exit
done