#!/bin/bash
set -e

GO=''

while getopts d:g: flag
do
    case "${flag}" in
        g) GO=${OPTARG};;
        *) exit 1;;
    esac
done

[[ -n "$GO" ]] || GO='go'

# Store the current directory
ORIGINAL_DIR=$(pwd)

# Change to the CLI tool directory
cd "$(dirname "$0")/../cmd/nrdot-collector-builder" || exit 1

OUTPUT=$(${GO} run main.go manifest update --json --config "../../distributions/*/manifest.yaml")

# Return to the original directory
cd "$ORIGINAL_DIR" || exit 1

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

# Extract the current beta core version
current_beta_core=$(echo "$OUTPUT" | jq -r '.currentVersions.betaCoreVersion')
current_beta_core=${current_beta_core#v}
escaped_current_beta_core=${current_beta_core//./\\.}
next_beta_core=$(echo "$OUTPUT" | jq -r '.nextVersions.betaCoreVersion')
next_beta_core=${next_beta_core#v}

#  If the current beta core version is not equal to the next beta core version, update the Makefile
if [[ "$current_beta_core" != "$next_beta_core" ]]; then
  echo "Updating Makefile from $current_beta_core to $next_beta_core"
  # Update Makefile OCB version
  sed_inplace "s/OTELCOL_BUILDER_VERSION ?= $escaped_current_beta_core/OTELCOL_BUILDER_VERSION ?= $next_beta_core/" Makefile
else
  echo "No update needed for the Makefile."
fi
