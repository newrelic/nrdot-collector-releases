# Collector Distributions

This README covers topics that apply to all distributions. For distribution-specific information please refer to:
- [nrdot-collector-host](./nrdot-collector-host/README.md)
- [nrdot-collector-k8s](./nrdot-collector-k8s/README.md)
- [nrdot-collector](./nrdot-collector/README.md)
- [nrdot-collector-experimental](./nrdot-collector-experimental/README.md)

We also provide [FIPS-compliant docker images](../fips/README.md).

## Installation

### Docker
Each distribution is available as a Docker image under the [newrelic](https://hub.docker.com/u/newrelic?page=1&search=nrdot-collector) organization on Docker Hub.

In order to run the collector via docker, you'll have to supply the required environment variables, see also [Configuration](#configuration). Using the `host` distribution as an example:
```bash
docker run -e NEW_RELIC_LICENSE_KEY='your-ingest-license-key' newrelic/nrdot-collector-host
```

### Linux packages and Archives 
All linux packages and archives are available under [Releases](https://github.com/newrelic/opentelemetry-collector-releases/releases), including checksums and signatures.

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

#### Packages
If a distribution provides linux packages (refer to its README), you can follow the instructions below to install them (using the `host` distribution as an example).

> Note: `systemd` is required for automatic service configuration.

##### DEB Installation
```bash
export collector_distro="nrdot-collector-host"
export collector_version="1.7.0"
export collector_arch="amd64" # or arm64
export license_key="YOUR_LICENSE_KEY"

curl "https://github.com/newrelic/nrdot-collector-releases/releases/download/${collector_version}/${collector_distro}_${collector_version}_linux_${collector_arch}.deb" --location --output collector.deb
sudo dpkg -i collector.deb
echo "NEW_RELIC_LICENSE_KEY=${license_key}" | sudo tee -a /etc/${collector_distro}/${collector_distro}.conf > /dev/null
sudo systemctl reload-or-restart "${collector_distro}.service"
```

### RPM Installation
```bash
export collector_distro="nrdot-collector-host"
export collector_version="1.7.0"
export collector_arch="x86_64" # or arm64
export license_key="YOUR_LICENSE_KEY"

curl "https://github.com/newrelic/nrdot-collector-releases/releases/download/${collector_version}/${collector_distro}_${collector_version}_linux_${collector_arch}.rpm" --location --output collector.rpm
sudo rpm -i collector.rpm
echo "NEW_RELIC_LICENSE_KEY=${license_key}" | sudo tee -a /etc/${collector_distro}/${collector_distro}.conf > /dev/null
sudo systemctl reload-or-restart "${collector_distro}.service"
```

### Archives
Archives contain the binary and the default configuration which is usually `config.yaml` unless the distro packages multiple defaults, e.g. `nrdot-collector-k8s`.
```bash
export collector_distro="nrdot-collector-host"
export collector_version="1.7.0"
export collector_arch="amd64" # or arm64
export license_key="YOUR_LICENSE_KEY"
curl "https://github.com/newrelic/nrdot-collector-releases/releases/download/${collector_version}/${collector_distro}_${collector_version}_linux_${collector_arch}.tar.gz" --location --output collector.tar.gz
tar -xzf collector.tar.gz
NEW_RELIC_LICENSE_KEY="${license_key}" ./nrdot-collector-host --config ./config.yaml 
```

## Configuration

### Customize Default Configuration
If the distribution provides a default configuration, some options are exposed via environment variables:

| Environment Variable | Description | Default |
|---|---|---|
| `NEW_RELIC_LICENSE_KEY` | New Relic ingest key | N/A - Required |
| `NEW_RELIC_MEMORY_LIMIT_MIB` | Maximum amount of memory to be used | 100 |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | New Relic OTLP endpoint to export metrics to, see [official docs](https://docs.newrelic.com/docs/opentelemetry/best-practices/opentelemetry-otlp/) | `https://otlp.nr-data.net` |

### Advanced: Using your own collector configuration

We recommend using the default configuration, but you can always supply your own via the `--config` [flag](https://opentelemetry.io/docs/collector/configuration/). The full list of components available for configuration is available in the respective `manifest.yaml`.

## Additional Notes

### Healthcheck
Each NRDOT distribution enables the [healthcheckextension](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/extension/healthcheckextension/README.md) which the default configuration exposes at `localhost:13133`:
```
# curl localhost:13133
{"status":"Server available","upSince":"2025-01-01T23:46:56.653164293Z","uptime":"15.241689299s"}
```

## Troubleshooting

Please refer to our [troubleshooting guide](./TROUBLESHOOTING.md).
