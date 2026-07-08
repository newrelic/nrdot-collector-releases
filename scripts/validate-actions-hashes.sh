#!/bin/bash
# Copyright New Relic, Inc. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

set -e

# From all workflows, extract every unique action-hash pair.
repo_pins=$(grep -rhoE 'uses: +[^ /]+/[^@ ]+@[a-f0-9]{40}' .github/workflows/ \
    | sed -E 's/uses: +//; s/@/ /')

# Any action which is represented more than once in repo_pins is misaligned
misaligned=$(echo "$repo_pins" | sort -u | awk '{print $1}' | uniq -d)

# If any actions are misaligned, print each divergent hash and where it's pinned.
if [ -n "$misaligned" ]; then
    echo -e "❌ Some actions have misaligned commit hashes!\n"
    for action in $misaligned; do
        grep -rnoE "${action}@[a-f0-9]{40}" .github/workflows/ \
            | sed -E "s|:${action}@| |" \
            | awk '{print $2, $1}' \
            | sort \
            | awk -v action="$action" '{ if ($1 != prev) { print "⚠️  " action "@" $1; prev=$1 } print "  - " $2 }'
        echo
    done
    exit 1
else
    echo "✅ All action commit hashes match!"
fi
