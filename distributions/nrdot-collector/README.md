# nrdot-collector

| Status | |
|-----------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Distro    | `nrdot-collector`                                                                                                                                                                                      |
| Stability | defined by use case, see ['Use Cases' below](#use-cases)                                                                                                                                                                                                    |
| Artifacts | [Docker images on DockerHub](https://hub.docker.com/r/newrelic/nrdot-collector)<br> [Linux packages and archives under GitHub Releases](https://github.com/newrelic/nrdot-collector-releases/releases) |

The core NRDOT collector distribution with components for various monitoring needs replacing existing distributions, see [Use Cases](#use-cases).

Note: See [general README](../README.md) for information that applies to all distributions.

## Use Cases

| Use Case              | Stability | Replaces                  | Documentation |
|-----------------------|-----------|---------------------------|---------------|
| Host Monitoring (default)      | `public`  | `nrdot-collector-host`    | [See 'Host Monitoring' below](#host-monitoring) |
| ATP    | `alpha`  |  N/A (new)   | [Docs](https://docs.newrelic.com/docs/opentelemetry/nrdot/atp/overview) |
| Gateway Mode          | `alpha`   | N/A (new)                 | [See 'Gateway Mode' below](#gateway-mode) |
| On-Host Integrations (OHI) | `alpha` | N/A (new)                 | [Kafka](https://docs.newrelic.com/docs/opentelemetry/integrations/kafka/overview/), [NGINX](https://docs.newrelic.com/docs/opentelemetry/integrations/nginx/nginx-otel-overview/), [ElasticSearch](https://docs.newrelic.com/docs/opentelemetry/integrations/elasticsearch/elasticsearch-otel-integration-overview/), [RabbitMQ](https://docs.newrelic.com/docs/opentelemetry/integrations/rabbitmq/overview/) |

Note: While it's technically possible to have a single collector serve multiple use cases at the same time, we generally do not recommend or support this pattern due to the operational complexity that comes with it (configuration, deployment, scaling, ...). Instead we recommend deploying one collector per use case and chain them as necessary. Please note that when we say 'one collector' we refer to a logical service, not a single instance, i.e. you should still employ common scaling practices to ensure your architecture is resilient.  
 
## Host Monitoring

Monitor the host the collector is deployed on via `hostmetricsreceiver` and `filelogreceiver`, and enrich OTLP data with host metadata via the `otlpreceiver` and `resourcedetectionprocessor`.

This is the default use case for `nrdot-collector`, i.e. the published artifacts set the collector's configuration to the packaged `config.yaml` which enables this use case. Other use cases will provide instructions on how to deploy the collector to override this behavior.

### Installation

The following instructions assume you have read and understood the [general installation instructions](../README.md#installation).

#### Linux Packages and NRDOT_MODE=ROOT
Our linux packages (deb, rpm) install the collector as a `systemd` service. By default the collector is installed as a non-root user to prevent unintended access by accident. While this is usually sufficient for the `hostmetricsreceiver` to scrape host metrics, the `filelogreceiver` is likely to run into permission issues reading the default files from the `/var/log` directory and report errors for the affected files.
Your options to address this issue are:
- adjust the list of files to avoid accessing privileged files, either by providing your own complete config or by overwriting the list, e.g. `--config 'yaml:receivers::filelog::include: [/var/log/dpkg.log, /var/log/messages]'`
- provide access to those files for the collector user, see `nrdot-collector.service` for the exact IDs
- install the collector in root mode by setting the environment variable `NRDOT_MODE=ROOT` before calling `dpkg`/`rpm` to install the service

#### Containerized Environments
If you're deploying the collector as a container, make sure to configure the [root_path](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/hostmetricsreceiver/README.md#collecting-host-metrics-from-inside-a-container-linux-only) and mount the host's file system accordingly, otherwise NRDOT will not be able to collect host metrics. See also [our troubleshooting guide](./TROUBLESHOOTING.md) for more details.

### Configuration

See [general README](../README.md) for configuration that applies to all distributions.

#### Use-case specific configuration

| Environment Variable | Description | Default |
|---|---|---|
| `OTEL_RESOURCE_ATTRIBUTES` | Key-value pairs to be used as resource attributes, see [OTel Docs](https://opentelemetry.io/docs/languages/sdk-configuration/general/#otel_resource_attributes) | N/A |

#### Enable process metrics
Process metrics are disabled by default due to their high cardinality. To enable them, reconfigure the `hostmetricsreceiver` per the [receiver docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/hostmetricsreceiver#getting-started). Note the distinction between [processesscraper (`system.processes.*` metrics)](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/hostmetricsreceiver/internal/scraper/processesscraper/documentation.md) and [processscraper (`process.*` metrics)](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/hostmetricsreceiver/internal/scraper/processscraper/documentation.md).

Example configuration:
```shell
newrelic/nrdot-collector --config /etc/nrdot-collector/config.yaml \
--config='yaml:receivers::hostmetrics::scrapers::processes: ' \
--config='yaml:receivers::hostmetrics::scrapers::process: { metrics: { process.cpu.utilization: { enabled: true }, process.cpu.time: { enabled: false } } }'
```

---

## Gateway Mode

Centralized telemetry collection and processing for environments with multiple sources.

### Overview

Gateway mode deploys the collector as a central aggregation point for telemetry data. This mode is useful for:
- Reducing direct connections to backend services
- Centralizing telemetry processing and transformation
- Implementing sampling and filtering policies
- Buffering and batching telemetry data

Gateway mode includes additional components beyond host monitoring capabilities to support centralized collection and processing.

---

## Troubleshooting

Refer to the [troubleshooting guide](./TROUBLESHOOTING.md).
