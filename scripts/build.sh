#!/bin/bash

REPO_DIR="$( cd "$(dirname "$( dirname "${BASH_SOURCE[0]}" )")" &> /dev/null && pwd )"
BUILDER=''

# default values
skipcompilation=true
validate=true
fips=false
cgo=0

while getopts d:s:l:b:f:c: flag

do
    case "${flag}" in
        d) distributions=${OPTARG};;
        s) skipcompilation=${OPTARG};;
        l) validate=${OPTARG};;
        b) BUILDER=${OPTARG};;
        f) fips=${OPTARG};;
        c) cgo=${OPTARG};;
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
elif [[ "$fips" == true ]]; then
    echo "❌ ERROR: FIPS requires skip compilation."
    echo "Skip Compilation is false."
    exit 1
fi

echo "Distributions to build: $distributions";

for distribution in $(echo "$distributions" | tr "," "\n")
do
    pushd "${REPO_DIR}/distributions/${distribution}" > /dev/null || exit

    manifest_file="manifest.yaml";
    build_folder="_build"

    if [[ "$fips" == true ]]; then
      yq eval '
         .dist.name += "-fips" |
         .dist.description += "-fips" |
         .dist.output_path += "-fips"' manifest.yaml > manifest-fips.yaml
      manifest_file="manifest-fips.yaml"
      build_folder="_build-fips"
      cgo=1
    fi

    mkdir -p $build_folder

    echo "Building: $distribution"
    echo "Using Builder: $(command -v "$BUILDER")"
    echo "Using Go: $(command -v go)"
    echo "Using FIPS: ${fips}"

    if CGO_ENABLED=${cgo} "$BUILDER" --skip-compilation="${skipcompilation}" --config ${manifest_file} > ${build_folder}/build.log 2>&1; then
        if [[ "$fips" == true ]]; then
            echo "Copying fips.go into _build-fips."
            cp ../../fips/fips.go ./$build_folder
        fi
        echo "✅ SUCCESS: distribution '${distribution}' built."
    else
        echo "❌ ERROR: failed to build the distribution '${distribution}'."
        echo "🪵 Build logs for '${distribution}'"
        echo "----------------------"
        cat $build_folder/build.log
        echo "----------------------"
        exit 1
    fi

    popd > /dev/null || exit
done
