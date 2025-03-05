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

### Distribution-specific configuration

Note: See [general README](../README.md) for information that applies to all distributions.

There is currently no distribution-specific configuration.