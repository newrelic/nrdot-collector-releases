#!/bin/bash
set -e


# Function to validate semantic version and strip leading 'v'
validate_and_strip_version() {
  local var_name=$1
  local version=${!var_name}
  # Strip leading 'v' if present
  version=${version#v}
  if [[ ! $version =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Invalid version: $version. Must be a semantic version (e.g., 1.2.3)."
    exit 1
  fi
  eval "$var_name='$version'"
}


# Get the most recent release tag from the open-telemetry/opentelemetry-collector-contrib repo
get_latest_otel_release_tag() {
    local repo="open-telemetry/opentelemetry-collector-releases"
    local latest_tag=$(curl --silent "https://api.github.com/repos/$repo/releases/latest" | jq -r .tag_name)
    if [[ -z $latest_tag ]]; then
        echo "Failed to fetch the latest release tag from $repo."
        exit 1
    fi
    
    validate_and_strip_version latest_tag
    echo "$latest_tag"
}


otel_version=$(get_latest_otel_release_tag)
manifest_url="https://raw.githubusercontent.com/open-telemetry/opentelemetry-collector-releases/refs/tags/v${otel_version}/distributions/otelcol/manifest.yaml"
manifest_content=$(curl --silent "$manifest_url")

# Extract the distribution version from the manifest content
next_beta_core=$(echo "$manifest_content" | awk '/^.*go\.opentelemetry\.io\/collector\/.* v0/ {print $4; exit}')
next_beta_contrib=$(echo "$manifest_content" | awk '/^.*github\.com\/open-telemetry\/opentelemetry-collector-contrib\/.* v0/ {print $4; exit}')
next_stable=$(echo "$manifest_content" | awk '/^.*go\.opentelemetry\.io\/collector\/.* v1/ {print $4; exit}')

validate_and_strip_version next_beta_core
validate_and_strip_version next_beta_contrib
validate_and_strip_version next_stable

echo "Next beta core version: $next_beta_core"
echo "Next beta contrib version: $next_beta_contrib"
echo "Next stable version: $next_stable"

# Get the current versions from the manifest.yaml files
current_beta_core=$(awk '/^.*go\.opentelemetry\.io\/collector\/.* v0/ {print $4; exit}' distributions/nrdot-collector-host/manifest.yaml)
current_beta_contrib=$(awk '/^.*github\.com\/open-telemetry\/opentelemetry-collector-contrib\/.* v0/ {print $4; exit}' distributions/nrdot-collector-host/manifest.yaml)
current_stable=$(awk '/^.*go\.opentelemetry\.io\/collector\/.* v1/ {print $4; exit}' distributions/nrdot-collector-host/manifest.yaml)

validate_and_strip_version current_beta_core
validate_and_strip_version current_beta_contrib
validate_and_strip_version current_stable

echo "Current beta core version: $current_beta_core"
echo "Current beta contrib version: $current_beta_contrib"
echo "Current stable version: $current_stable"


# add escape characters to the current versions to work with sed
escaped_current_beta_core=${current_beta_core//./\\.}
escaped_current_beta_contrib=${current_beta_contrib//./\\.}
escaped_current_stable=${current_stable//./\\.}

# Determine the OS and set the sed -i command accordingly
if [[ "$OSTYPE" == "darwin"* ]]; then
  # macOS
  function sed_inplace {
  	sed -i '' "$@"
  }
else
  function sed_inplace {
    	sed -i'' "$@"
  }
fi

# Update versions in each manifest file
echo "Updating core beta version from $current_beta_core to $next_beta_core,"
echo "core stable version from $current_stable to $next_stable,"
echo "contrib beta version from $current_beta_contrib to $next_beta_contrib,"
for file in ./distributions/*/manifest.yaml; do
  if [ -f "$file" ]; then
    sed_inplace "s/\(^.*go\.opentelemetry\.io\/collector\/.*\) v$escaped_current_beta_core/\1 v$next_beta_core/" "$file"
    sed_inplace "s/\(^.*github\.com\/open-telemetry\/opentelemetry-collector-contrib\/.*\) v$escaped_current_beta_contrib/\1 v$next_beta_contrib/" "$file"
    sed_inplace "s/\(^.*go\.opentelemetry\.io\/collector\/.*\) v$escaped_current_stable/\1 v$next_stable/" "$file"
  else
    echo "File $file does not exist"
  fi
done

# Update Makefile OCB version
sed_inplace "s/OTELCOL_BUILDER_VERSION ?= $escaped_current_beta_core/OTELCOL_BUILDER_VERSION ?= $next_beta_core/" Makefile
