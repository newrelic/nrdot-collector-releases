# nrdot-collector-host

| Status    |                                                                                                                                                                                                             |
|-----------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Distro    | `nrdot-collector-host`                                                                                                                                                                                      |
| Stability | `public` (deprecated in favor of [nrdot-collector](../nrdot-collector/README.md)) as drop-in replacement                                                                                                                                                                                                  |
| Artifacts | [Docker images on DockerHub](https://hub.docker.com/r/newrelic/nrdot-collector-host)<br> [Linux packages and archives under GitHub Releases](https://github.com/newrelic/nrdot-collector-releases/releases) |

A distribution of the NRDOT collector focused on
- monitoring the host the collector is deployed on via `hostmetricsreceiver` and `filelogreceiver`
- enriching other OTLP data with host metadata via the `otlpreceiver` and `resourcedetectionprocessor`

Note: See [general README](../README.md) for information that applies to all distributions.

## Installation

The following instructions assume you have read and understood the [general installation instructions](../README.md#installation).

### Linux Packages and NRDOT_MODE=ROOT
Our linux packages (deb, rpm) install the collector as a `systemd` service. By default the collector is installed as a non-root user to prevent unintended access by accident. While this is usually sufficient for the `hostmetricsreceiver` to scrape host metrics, the `filelogreceiver` is likely to run into permission issues reading the default files from the `/var/log` directory and report errors for the affected files.
Your options to address this issue are:
- adjust the list of files to avoid accessing privileged files, either by providing your own complete config or by overwriting the list, e.g. `--config 'yaml:receivers::filelog::include: [/var/log/dpkg.log, /var/log/messages]'`
- provide access to those files for the collector user, see `nrdot-collector-host.service` for the exact IDs
- install the collector in root mode by setting the environment variable `NRDOT_MODE=ROOT` before calling `dpkg`/`rpm` to install the service

### Containerized Environments
If you're deploying the collector as a container, make sure to configure the [root_path](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/hostmetricsreceiver/README.md#collecting-host-metrics-from-inside-a-container-linux-only) and mount the host's file system accordingly, otherwise NRDOT will not be able to collect host metrics. See also [our troubleshooting guide](./TROUBLESHOOTING.md) for more details.


## Configuration

Note: See [general README](../README.md) for information that applies to all distributions.


### Distribution-specific configuration

| Environment Variable | Description | Default |
|---|---|---|
| `OTEL_RESOURCE_ATTRIBUTES` | Key-value pairs to be used as resource attributes, see [OTel Docs](https://opentelemetry.io/docs/languages/sdk-configuration/general/#otel_resource_attributes) | N/A |

#### Enable process metrics
Process metrics are disabled by default as they are quite noisy. If you want to enable them, you can do so by reconfiguring the `hostmetricsreceiver`, see also [receiver docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/hostmetricsreceiver#getting-started). Note that there is a [processesscraper (`system.processes.*` metrics)](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/hostmetricsreceiver/internal/scraper/processesscraper/documentation.md) and a [processscraper (`process.*` metrics)](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/hostmetricsreceiver/internal/scraper/processscraper/documentation.md) with separate options. An example configuration would look like this:
```shell
newrelic/nrdot-collector-host --config /etc/nrdot-collector-host/config.yaml \
--config='yaml:receivers::hostmetrics::scrapers::processes: ' \
--config='yaml:receivers::hostmetrics::scrapers::process: { metrics: { process.cpu.utilization: { enabled: true }, process.cpu.time: { enabled: false } } }'
```

## Troubleshooting

Please refer to our [troubleshooting guide](./TROUBLESHOOTING.md).
