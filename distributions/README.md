# Collector Distributions

## Installation

### Docker

Each distribution is available as a Docker image under the [newrelic](https://hub.docker.com/u/newrelic?page=1&search=nrdot-collector) organization on Docker Hub.

### OS-specific packages
For certain distributions, signed OS-specific packages are also available under [Releases](https://github.com/newrelic/opentelemetry-collector-releases/releases) on GitHub.

#### Verifying Signatures

TODO: add gpg verification instructions using https://github.com/newrelic/nrdot-collector-releases/blob/main/nrdot.gpg

#### Example: Installing the host distribution on Ubuntu with systemd
```
export collector_distro="nrdot-collector-host"
export collector_version="nrdot-collector-host"
export collector_arch="amd64"
curl "https://github.com/newrelic/nrdot-collector-releases/releases/download/${collector_version}/${collector_distro}_${collector_version}_linux_${collector_arch}.deb" --location --output collector.deb
# This automatically starts the collector as a systemd service
sudo dpkg -i collector.deb
echo 'NEW_RELIC_LICENSE_KEY=INSERT_YOUR_INGEST_KEY' | sudo tee -a /etc/${collector_distro}/${collector_distro}.conf > /dev/null
# Restart to use new license key
sudo systemctl reload-or-restart "${collector_distro}.service"
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

