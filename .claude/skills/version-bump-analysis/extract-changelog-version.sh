#!/bin/bash

# Extract a specific version section from an OpenTelemetry changelog
# Usage: extract-changelog-version.sh <url> <version>
# Example: extract-changelog-version.sh "https://raw.githubusercontent.com/open-telemetry/opentelemetry-collector/main/CHANGELOG.md" "v0.144.0"

set -euo pipefail

if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <changelog_url> <version>" >&2
    echo "Example: $0 'https://raw.githubusercontent.com/open-telemetry/opentelemetry-collector/main/CHANGELOG.md' 'v0.144.0'" >&2
    exit 1
fi

CHANGELOG_URL="$1"
VERSION="$2"

# Fetch changelog and extract the specific version section
# The version appears in headers like "## v1.50.0/v0.144.0"
# We extract from that header until the next "## " header or "<!-- previous-version -->"
OUTPUT=$(curl -s "$CHANGELOG_URL" | awk -v version="$VERSION" \
'BEGIN { found=0; printing=0 } \
/^## / { if (found) { exit } if ($0 ~ version) { found=1; printing=1; print $0; next } } \
/^<!-- previous-version -->/ { if (printing) { exit } } \
printing { print $0 }')

# Check if version was found
if [ -z "$OUTPUT" ]; then
    echo "Error: Version $VERSION not found in changelog at $CHANGELOG_URL" >&2
    exit 1
fi

echo "$OUTPUT"
exit 0
