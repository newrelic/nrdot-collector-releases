# nrdot-collector

| Status    |                                                                                                                                                                                                             |
|-----------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Distro    | `nrdot-collector`                                                                                                                                                                                      |
| Stability | `alpha`                                                                                                                                                                                                    |

A distribution of the NRDOT collector focused on
- monitoring the host the collector is deployed on via `hostmetricsreceiver` and `filelogreceiver`
- enriching other OTLP data with host metadata via the `otlpreceiver` and `resourcedetectionprocessor`
- facilitating gateway mode deployments with additional components for centralized telemetry collection and processing

This distribution includes all the capabilities of `nrdot-collector-host` plus additional components to support gateway mode deployments, allowing it to act as a central collection point for telemetry data from multiple sources.

Note: See [general README](../README.md) for information that applies to all distributions.

## Installation

The following instructions assume you have read and understood the [general installation instructions](../README.md#installation).

### Containerized Environments
If you're deploying the `nrdot-collector` distribution as a container, make sure to configure the [root_path](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/hostmetricsreceiver/README.md#collecting-host-metrics-from-inside-a-container-linux-only) and mount the host's file system accordingly, otherwise NRDOT will not be able to collect host metrics properly.
See also [our troubleshooting guide](./TROUBLESHOOTING.md) for more details.

### Gateway Mode Deployment
When deploying in gateway mode, the collector acts as a central aggregation point for telemetry data. This mode is particularly useful for:
- Reducing the number of direct connections to backend services
- Centralizing telemetry processing and transformation
- Implementing sampling and filtering policies
- Buffering and batching telemetry data

## Configuration

Note: See [general README](../README.md) for information that applies to all distributions.

### Distribution-specific configuration

| Environment Variable | Description | Default |
|---|---|---|
| `OTEL_RESOURCE_ATTRIBUTES` | Key-value pairs to be used as resource attributes, see [OTel Docs](https://opentelemetry.io/docs/languages/sdk-configuration/general/#otel_resource_attributes) | N/A |

#### Enable process metrics
Process metrics are disabled by default as they are quite noisy. If you want to enable them, you can do so by reconfiguring the `hostmetricsreceiver`, see also [receiver docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/hostmetricsreceiver#getting-started). Note that there is a [processesscraper (`system.processes.*` metrics)](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/hostmetricsreceiver/internal/scraper/processesscraper/documentation.md) and a [processscraper (`process.*` metrics)](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/hostmetricsreceiver/internal/scraper/processscraper/documentation.md) with separate options. An example configuration would look like this:
```shell
newrelic/nrdot-collector --config /etc/nrdot-collector/config.yaml \
--config='yaml:receivers::hostmetrics::scrapers::processes: ' \
--config='yaml:receivers::hostmetrics::scrapers::process: { metrics: { process.cpu.utilization: { enabled: true }, process.cpu.time: { enabled: false } } }'
```

## Troubleshooting

Please refer to our [troubleshooting guide](./TROUBLESHOOTING.md).
