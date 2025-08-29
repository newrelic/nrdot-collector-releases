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
    echo "List of distributions to generate systemd .service and .conf files not provided. Use '-d' to specify the names of the distributions use. Ex.:"
    echo "$0 -d nrdot-collector-host"
    exit 1
fi

echo "Distributions to generate: $distributions";

for distribution in $(echo "$distributions" | tr "," "\n")
do
    if [[ "$distribution" = "nrdot-collector-k8s" ]]; then
        echo "Skipping $distribution: dist does not require systemd files"
        continue
    else
        echo "Generating .conf and .service systemd files for $distribution"
    fi
    ${GO} run cmd/systemd/main.go -d "${distribution}" -o env > "./distributions/${distribution}/${distribution}.conf"
    ${GO} run cmd/systemd/main.go -d "${distribution}" -o service > "./distributions/${distribution}/${distribution}.service"

    ${GO} run cmd/systemd/main.go -d "${distribution}" -o env -f > "./distributions/${distribution}/${distribution}-fips.conf"
    ${GO} run cmd/systemd/main.go -d "${distribution}" -o service -f > "./distributions/${distribution}/${distribution}-fips.service"
done