#!/bin/bash
# Copyright New Relic, Inc. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

# This script serves as a wrapper around chloggen so that we can use our own change types.
# Chloggen's default change types are "breaking", "deprecation", "bug_fix", "new_component", and "enhancement".
# These values are hard-coded and not configurable, necessitating a wrapper if we want to use our own custom types.

set -euo pipefail
REPO_DIR="$( cd "$(dirname "$( dirname "${BASH_SOURCE[0]}" )")" &> /dev/null && pwd )"
CHLOGGEN_DIR="$REPO_DIR/.chloggen"

CHLOGGEN=''

set_command() {
    if [ -n "$COMMAND" ]; then
        echo "Error: only one command flag allowed (-n, -v, -p, -u)" >&2
        exit 1
    fi
    COMMAND=$1
}

while getopts nvpub: flag
do
    case "${flag}" in
        b) CHLOGGEN=${OPTARG};;
        n) set_command new;;
        v) set_command validate;;
        p) set_command preview;;
        u) set_command update;;
        *) exit 1;;
    esac
done

if [ -z "$CHLOGGEN" ]; then
    CHLOGGEN=chloggen
fi

if [ -z "$COMMAND" ]; then
    echo "Error: specify a command flag: -n (new), -v (validate), -p (preview), or -u (update)" >&2
    exit 1
fi

# --new command (no need for translation)
if [ "$COMMAND" = new ]; then
    BRANCH_NAME=$(git branch --show-current)
    filepath=$("$CHLOGGEN" new --config "$CHLOGGEN_DIR/config.yaml" --filename "$BRANCH_NAME" | sed -n 's/^Changelog entry template copied to: //p')

    # Pre-populate the issues field with the PR number if one exists for this branch
    if gh auth status >/dev/null 2>&1; then
        PR_NUMBER=$(gh pr view "$BRANCH_NAME" --json number --jq '.number' 2>/dev/null)
        if [ -n "$PR_NUMBER" ]; then
            yq -i ".issues = [$PR_NUMBER] | .issues style=\"flow\"" "$filepath"
        else
            echo "ℹ️ No PR found for branch '$BRANCH_NAME'; issues field not populated." >&2
        fi

        # Infer change_type from the conventional commit prefix of the PR title
        PR_TITLE=$(gh pr view "$BRANCH_NAME" --json title --jq '.title' 2>/dev/null)
        case "$PR_TITLE" in
            feat*) yq -i '.change_type = "feature"' "$filepath";;
            perf*) yq -i '.change_type = "feature"' "$filepath";;
            fix*)  yq -i '.change_type = "bug_fix"' "$filepath";;
            docs*) yq -i '.change_type = "docs"' "$filepath";;
            *)     echo "ℹ️ change_type not inferred; PR title '$PR_TITLE' prefix does not map to a change_type (feat, fix, docs)." >&2;;
        esac
    else
        echo "⚠️ Warning: gh not authenticated; pr# and change type cannot be inferred." >&2
    fi

    echo "$filepath"
    exit 0
fi

# Semi-arbitrarily map our in-house change types to change types allowed by chloggen.
# When the changelog is updated, these will be placed under the proper headers via summary.tmpl
translate_change_type() {
    case "$1" in
        feature) echo enhancement;;
        bug_fix) echo bug_fix;;
        docs) echo new_component;;
        *) echo "Error: invalid change_type '$1'. Specify one of [feature bug_fix docs]" >&2; return 1;;
    esac
}

# --update translates in-place so chloggen consumes and clears the real entries.
# --validate and --preview translate into a temp dir, leaving the real entries untouched.
if [ "$COMMAND" = update ]; then
    dir=$CHLOGGEN_DIR
    CONFIG="$CHLOGGEN_DIR/config.yaml"
else
    dir=$(mktemp -d "$CHLOGGEN_DIR/tmp.XXXXXX")
    trap 'rm -rf "$dir"' EXIT
    # Need to create a new config to point chloggen at the temp dir
    cp "$CHLOGGEN_DIR/config.yaml" "$dir/config.yaml"
    CONFIG="$dir/config.yaml"
    yq -i ".entries_dir = \"$dir\"" "$CONFIG"
fi

# Translate each entry's change_type into a chloggen-native type (in $dir)
for entry in "$CHLOGGEN_DIR"/*.yaml; do
    base=$(basename "$entry")
    
    case "$base" in
        TEMPLATE.yaml|config.yaml) continue;;
    esac
    
    old_type=$(yq '.change_type' "$entry")
    new_type=$(translate_change_type "$old_type") || exit 1

    if [ "$COMMAND" = update ]; then
        # In-place edit of the real entry ($entry and $dir/$base are the same file)
        yq -i ".change_type = \"$new_type\"" "$entry"
    else
        # Read the real entry, write the translated copy into the temp dir
        yq ".change_type = \"$new_type\"" "$entry" > "$dir/$base"
    fi
done

# Run chloggen against the translated entries
case "$COMMAND" in
    validate) "$CHLOGGEN" validate --config "$CONFIG";;
    preview)  "$CHLOGGEN" update --config "$CONFIG" --dry;;
    update)   "$CHLOGGEN" update --config "$CONFIG" --version "$("$REPO_DIR/scripts/get-version.sh")";;
esac
