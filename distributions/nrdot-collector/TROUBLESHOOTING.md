# Troubleshooting for nrdot-collector

For general NRDOT troubleshooting, see [this guide](../TROUBLESHOOTING.md). This document assumes you are familiar with
the troubleshooting tools mentioned.

## Known issues

### Missing host entity in New Relic UI due to missing `host.id`
If you are [seeing telemetry getting ingested into New Relic](../TROUBLESHOOTING.md#user-content-stablelink-telemetry-not-reaching-new-relic) but even after a few minutes of waiting the Host UI does not show any host entities, you might be running into the limitations of the [resourcedetectionprocessor](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/processor/resourcedetectionprocessor/README.md) NRDOT uses to determine the `host.id` attribute required to [synthesize a host entity](https://github.com/newrelic/entity-definitions/blob/main/entity-types/infra-host/definition.yml#L62-L63). An example log message indicating this issue looks like this:
```
# example 1
2025-01-01T22:49:09.110Z        warn    system/system.go:143    failed to get host ID   {"otelcol.component.id": "resourcedetection", "otelcol.component.kind" : "Processor", "otelcol.pipeline.id": "logs/host", "otelcol.signal": "logs", "error": "failed to obtain \"host.id\": error detecting resource: host id not found in: /etc/machine-id or /var/lib/dbus/machine-id"}
# example 2
2025-01-01T23:07:27.866Z        warn    system/system.go:143    failed to get host ID   {"otelcol.component.id": "resourcedetection", "otelcol.component.kind": "Processor", "otelcol.pipeline.id": "metrics/host", "otelcol.signal": "metrics", "error": "empty \"host.id\""}
```
In order to resolve this, you can set the `host.id` attributes manually via the [environment variable](./README.md#configuration) `OTEL_RESOURCE_ATTRIBUTES`, e.g. `export OTEL_RESOURCE_ATTRIBUTES='host.id=my-custom-host-id'`.

### No `root_path` in containerized environments
The `hostmetricsreceiver` auto-detects the files to scrape system metrics from. When running in a container, this causes issues as the receiver would then scrape metrics of the container instead of the host system which you most likely want to monitor. In order to bridge this gap, the receiver provides the `root_path` option which allows you to specify the path where the host file system is available to the collector, most commonly by mounting it into the container. The warning indicating this issue looks like this:
```
2025-01-01T21:08:21.097Z	warn	filesystemscraper/factory.go:48	No `root_path` config set when running in docker environment, will report container filesystem stats. See https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/hostmetricsreceiver#collecting-host-metrics-from-inside-a-container-linux-only	{"otelcol.component.id": "hostmetrics", "otelcol.component.kind": "Receiver", "otelcol.signal": "metrics"}
```
In order to resolve this, make sure to follow the [receiver's docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/hostmetricsreceiver/README.md#collecting-host-metrics-from-inside-a-container-linux-only) to mount the host file system into the container at the `root_path` and configure the `root_path` accordingly, e.g.
```bash
docker run -v /:/hostfs \
-e NEW_RELIC_LICENSE_KEY='license-key' newrelic/nrdot-collector \
--config /etc/nrdot-collector/config.yaml \
--config 'yaml:receivers::hostmetrics::root_path: /hostfs'
```
