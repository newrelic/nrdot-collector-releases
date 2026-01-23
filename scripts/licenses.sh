#!/bin/bash
# Copyright New Relic, Inc. All rights reserved.
# SPDX-License-Identifier: Apache-2.0


REPO_DIR="$( cd "$(dirname "$( dirname "${BASH_SOURCE[0]}" )")" &> /dev/null && pwd )"

GO_LICENCE_DETECTOR=''
NOTICE_FILE=''

while getopts d:b:n:g: flag
do
  case "${flag}" in
    d) distributions=${OPTARG};;
    b) GO_LICENCE_DETECTOR=${OPTARG};;
    n) NOTICE_FILE=${OPTARG};;
    g) GO=${OPTARG};;
    *) exit 1;;
  esac
done

[[ -n "$NOTICE_FILE" ]] || NOTICE_FILE='THIRD_PARTY_NOTICES.md'

[[ -n "$GO_LICENCE_DETECTOR" ]] || GO_LICENCE_DETECTOR='go-licence-detector'

if [[ -z $distributions ]]; then
  echo "List of distributions to build not provided. Use '-d' to specify the names of the distributions to build. Ex.:"
  echo "$0 -d nrdot-collector-k8s"
  exit 1
fi

for distribution in $(echo "$distributions" | tr "," "\n")
do
  pushd "${REPO_DIR}/distributions/${distribution}/_build" > /dev/null || exit

  echo "📜 Building notice for ${distribution}..."

  ${GO} list -mod=mod -m -json all | ${GO_LICENCE_DETECTOR} \
    -rules "${REPO_DIR}/internal/assets/license/rules.json" \
    -noticeTemplate "${REPO_DIR}/internal/assets/license/THIRD_PARTY_NOTICES.md.tmpl" \
    -noticeOut "${REPO_DIR}/distributions/${distribution}/${NOTICE_FILE}"

  popd > /dev/null || exit
done
