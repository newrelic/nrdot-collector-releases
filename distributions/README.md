# Collector Distributions

This README covers topics that apply to all distributions. For distribution-specific information please refer to:
- [nrdot-collector-host](./nrdot-collector-host/README.md)
- [nrdot-collector-k8s](./nrdot-collector-k8s/README.md)

## Installation

### Docker
Each distribution is available as a Docker image under the [newrelic](https://hub.docker.com/u/newrelic?page=1&search=nrdot-collector) organization on Docker Hub.

In order to run the collector via docker, you'll have to supply the required environment variables, see also [Configuration](#configuration). Using the `host` distribution as an example:
```bash
docker run -e NEW_RELIC_LICENSE_KEY='your-ingest-license-key' newrelic/nrdot-collector-host
```

### OS-specific packages
For certain distributions (refer to its README), signed OS-specific packages are also available under [Releases](https://github.com/newrelic/opentelemetry-collector-releases/releases) on GitHub.

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

#### Installation on Ubuntu with systemd
We will use the `host` distribution as an example.
```bash
#!/bin/bash

set -e

# Step 1: choose your distro
export collector_distro="nrdot-collector-host"
# Step 2: choose the version
export collector_version="1.0.0"
# Step 3: Choose the arch - you can check https://github.com/newrelic/nrdot-collector-releases/releases/tag/${collector_version} for available options
export collector_arch="amd64"
# Step 4: Configure your license key
export license_key="YOUR_LICENSE_KEY"



if command -v dpkg &> /dev/null; then
    export package_installer_cmd="dpkg -i"
    export package_extension="deb"
elif command -v rpm &> /dev/null; then
    export package_installer_cmd="rpm -i"
    export package_extension="rpm"
else
    echo "Didn't detect supported package managers: dpkg, rpm - please install manually by using the tar.gz"
    exit 1
fi
for cmd in curl tee; do
    if ! command -v $cmd &> /dev/null; then
        echo "$cmd could not be found. Please install $cmd."
        exit 1
    fi
done

# Download the package
curl "https://github.com/newrelic/nrdot-collector-releases/releases/download/${collector_version}/${collector_distro}_${collector_version}_linux_${collector_arch}.${package_extension}" --location --output "collector.${package_extension}"
# This automatically starts the collector as a systemd service
sudo ${package_installer_cmd} "collector.${package_extension}"

# Add your New Relic ingest key to the systemd config
echo "NEW_RELIC_LICENSE_KEY=${license_key}" | sudo tee -a /etc/${collector_distro}/${collector_distro}.conf > /dev/null

# Restart to use new license key
sudo systemctl reload-or-restart "${collector_distro}.service"
# Data should now be flowing to New Relic and be available within a few minutes
```

## Configuration

### Customize Default Configuration
The default configuration exposes some options via environment variables:

| Environment Variable | Description | Default |
|---|---|---|
| `NEW_RELIC_LICENSE_KEY` | New Relic ingest key | N/A - Required |
| `NEW_RELIC_MEMORY_LIMIT_MIB` | Maximum amount of memory to be used | 100 |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | New Relic OTLP endpoint to export metrics to, see [official docs](https://docs.newrelic.com/docs/opentelemetry/best-practices/opentelemetry-otlp/) | `https://otlp.nr-data.net` |

### Advanced: Using your own collector configuration

We recommend using the default configuration, but you can always supply your own via the `--config` [flag](https://opentelemetry.io/docs/collector/configuration/). The full list of components available for configuration is available in the respective `manifest.yaml`.

## Troubleshooting

Please refer to our [troubleshooting guide](./TROUBLESHOOTING.md).
