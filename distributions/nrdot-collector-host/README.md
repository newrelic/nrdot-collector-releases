# nrdot-collector-host

| Status    |                                                                     |
|-----------|---------------------------------------------------------------------|
| Distro    | `nrdot-collector-host`                                              |
| Stability | `public`                                                            |
| Images    | [DockerHub](https://hub.docker.com/r/newrelic/nrdot-collector-host) |

A distribution of the NRDOT collector focused on
- monitoring the host the collector is deployed on via `hostmetricsreceiver` and `filelogreceiver`
- support piping other telemetry through it via the `otlpreceiver`

Distribution is available as docker image and as OS-specific package.

## Installation

The following instructions assume you have read and understood the [general installation instructions](../README.md#installation).

### Containerized Environments
If you're deploying the `host` distribution in a containerized environment like docker and want to scrape metrics of a linux host machine,
make sure to configure the [root_path](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/hostmetricsreceiver/README.md#collecting-host-metrics-from-inside-a-container-linux-only) and mount the host's file system accordingly.
See also [our troubleshooting guide](./TROUBLESHOOTING.md) for more details.


## Configuration

Note: See [general README](../README.md) for information that applies to all distributions.

There is currently no distribution-specific configuration.

## Troubleshooting

Please refer to our [troubleshooting guide](./TROUBLESHOOTING.md).
