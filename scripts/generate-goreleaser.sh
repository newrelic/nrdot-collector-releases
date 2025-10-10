#!/bin/bash

GO=''

while getopts d:g: flag
do
    case "${flag}" in
        d) distributions=${OPTARG};;
        g) GO=${OPTARG};;
        *) exit 1;;
    esac
done

[[ -n "$GO" ]] || GO='go'

if [[ -z $distributions ]]; then
    echo "List of distributions to generate the goreleaser not provided. Use '-d' to specify the names of the distributions use. Ex.:"
    echo "$0 -d nrdot-collector-host"
    exit 1
fi

echo "Generating goreleaser files for distributions: $distributions";

for distribution in $(echo "$distributions" | tr "," "\n")
do
    ${GO} run cmd/goreleaser/main.go -d "${distribution}" > "./distributions/${distribution}/.goreleaser.yaml"
    ${GO} run cmd/goreleaser/main.go -d "${distribution}" -f > "./distributions/${distribution}/.goreleaser-fips.yaml"
done
