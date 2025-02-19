# Collector Distributions

## Installation

### Docker

Each distribution is available as a Docker image under the [newrelic](https://hub.docker.com/u/newrelic?page=1&search=nrdot-collector) organization on Docker Hub.

### OS-specific packages
For certain distributions, signed OS-specific packages are also available under [Releases](https://github.com/newrelic/opentelemetry-collector-releases/releases) on GitHub.

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

