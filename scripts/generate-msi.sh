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
    echo "List of distributions to generate .wxs files not provided. Use '-d' to specify the names of the distributions use. Ex.:"
    echo "$0 -d nrdot-collector-host"
    exit 1
fi

echo "Generating .wxs files for distributions: $distributions";

for distribution in $(echo "$distributions" | tr "," "\n")
do
    if [[ $distribution = "nrdot-collector-k8s" ]]; then
        continue
    fi
    ${GO} run cmd/msi/main.go -d "${distribution}" > "./distributions/${distribution}/windows-installer.wxs"
    ${GO} run cmd/msi/main.go -d "${distribution}" -f > "./distributions/${distribution}/windows-installer-fips.wxs"
done
