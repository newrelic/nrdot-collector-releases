#!/bin/bash
# Copyright New Relic, Inc. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

# Writes the docs-website release notes .mdx file for a version: frontmatter built from changelog.yaml,
# followed by the GitHub release body. Missing versions or buckets yield empty lists.
# Usage: generate-docs-release-notes.sh <version> <release_date> <output_file>
# Env: RELEASE_BODY (release notes body)

set -euo pipefail

VERSION="${1:?Usage: $0 <version> <release_date> <output_file>}"
RELEASE_DATE="${2:?Usage: $0 <version> <release_date> <output_file>}"
OUTPUT_FILE="${3:?Usage: $0 <version> <release_date> <output_file>}"
CHANGELOG_YAML="$(dirname "$0")/../../changelog.yaml"

# Prints the summaries of one bucket as a JSON array (a valid YAML flow sequence).
summaries() {
  BUCKET="$1" yq -o=json -I=0 '(.[strenv(VERSION)][strenv(BUCKET)] // []) | map(.summary)' "$CHANGELOG_YAML"
}

export VERSION

cat > "$OUTPUT_FILE" <<MDX
---
subject: NRDOT
releaseDate: '$RELEASE_DATE'
version: $VERSION
features: $(summaries features)
enhancements: $(summaries enhancements)
bugs: $(summaries bug-fixes)
security: $(summaries security)
metaDescription: Release notes for NRDOT Collector version $VERSION
category: Collector
title: NRDOT Collector v$VERSION
---

${RELEASE_BODY:-}
MDX
