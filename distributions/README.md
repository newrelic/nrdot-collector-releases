# Collector Distributions

## Installation

### Docker

Each distribution is available as a Docker image under the [newrelic](https://hub.docker.com/u/newrelic?page=1&search=nrdot-collector) organization on Docker Hub.

### OS-specific packages
For certain distributions, signed OS-specific packages are also available under [Releases](https://github.com/newrelic/opentelemetry-collector-releases/releases) on GitHub.

#### Verifying Signatures

```bash
#!/bin/bash

set -e

# Verify that gpg, jq, and curl are installed
for cmd in gpg jq curl; do
    if ! command -v $cmd &> /dev/null; then
        echo "$cmd could not be found. Please install $cmd."
        exit 1
    fi
done

# Get the most recent release version from GitHub
RELEASE=$(curl -s https://api.github.com/repos/newrelic/nrdot-collector-releases/releases/latest | jq -r '.tag_name')

echo "Verifying release: $RELEASE"

# Download and import public gpg key
curl -s "https://raw.githubusercontent.com/newrelic/nrdot-collector-releases/refs/tags/${RELEASE}/nrdot.gpg" | gpg --import

# (optional) To remove the trust signature warning you'll need to manually trust the key
# gpg --edit-key 8ECAA86AB2C1904FAAC12E34B0EE4ACC08A81CD2

# Store artifacts in temp folder
ARTIFACTS_DIR=$(mktemp -d -t artifacts.XXXXXXXX)

trap cleanup exit
cleanup () {
    echo "cleaning up"
    rm -rf "$ARTIFACTS_DIR"
}

ASSETS_URL="https://api.github.com/repos/newrelic/nrdot-collector-releases/releases/tags/${RELEASE}"
ASSETS=$(curl -s $ASSETS_URL | jq -r '.assets[] | .browser_download_url')

# Download each asset
for ASSET_URL in $ASSETS; do
    echo "Downloading $ASSET_URL"
    curl -L --output-dir "$ARTIFACTS_DIR" -O $ASSET_URL
done

echo "Downloaded artifacts:"
ls -la $ARTIFACTS_DIR

for file in $ARTIFACTS_DIR/*.asc; do
    echo "Verifying $file"
    gpg --verify $file
done
```

#### Example: Installing the host distribution on Ubuntu with systemd
```bash
#!/bin/bash

set -e

for cmd in curl tee; do
    if ! command -v $cmd &> /dev/null; then
        echo "$cmd could not be found. Please install $cmd."
        exit 1
    fi
done

export collector_distro="nrdot-collector-host"
export collector_version="1.0.0"
export collector_arch="amd64"
curl "https://github.com/newrelic/nrdot-collector-releases/releases/download/${collector_version}/${collector_distro}_${collector_version}_linux_${collector_arch}.deb" --location --output collector.deb
# This automatically starts the collector as a systemd service
sudo dpkg -i collector.deb

# Add your New Relic ingest key
echo 'NEW_RELIC_LICENSE_KEY=INSERT_YOUR_INGEST_KEY' | sudo tee -a /etc/${collector_distro}/${collector_distro}.conf > /dev/null

# Restart to use new license key
sudo systemctl reload-or-restart "${collector_distro}.service"
# Data should now be flowing to New Relic and be available within a few minutes
```

## Configuration

### Components

The full list of components is available in the respective `manifest.yaml`

### Customize Default Configuration

The default configuration exposes some options via environment variables:

| Environment Variable | Description | Default |
|---|---|---|
| `NEW_RELIC_LICENSE_KEY` | New Relic ingest key | N/A - Required |
| `NEW_RELIC_MEMORY_LIMIT_MIB` | Maximum amount of memory to be used | 100 |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | New Relic OTLP endpoint to export metrics to, see [official docs](https://docs.newrelic.com/docs/opentelemetry/best-practices/opentelemetry-otlp/) | `https://otlp.nr-data.net` |

