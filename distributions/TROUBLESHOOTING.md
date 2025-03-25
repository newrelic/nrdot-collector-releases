# NRDOT Collector Troubleshooting
Please make sure to consult the [README](./README.md) first to ensure that you have the necessary prerequisites in place.
The collector is extremely powerful but therefore also easily misconfigured. As we cannot cover all the customization possibilities, we recommend consulting the [OTel Collector troubleshooting documentation](https://opentelemetry.io/docs/collector/troubleshooting/) if this guide does not help you resolve your issue. At its core, NRDOT is a [custom distribution](https://opentelemetry.io/docs/collector/custom-collector/), so many of the same steps will apply.

This guide covers general collector troubleshooting tools and common issues. Distribution-specific troubleshooting is available here:
- [nrdot-collector-host](./nrdot-collector-host/TROUBLESHOOTING.md)
- [nrdot-collector-k8s](./nrdot-collector-k8s/TROUBLESHOOTING.md)

However, note that those guides assume you are familiar with the tools mentioned in this guide.

## Helpful Tools

### NrIntegrationError
[NrIntegrationError](https://docs.newrelic.com/docs/data-apis/ingest-apis/metric-api/troubleshoot-nrintegrationerror-events/) are emitted when telemetry was sent to New Relic but the processing pipeline rejected it. Refer to the linked documentation for a root cause analysis.

### Customizing the configuration
Each distribution comes with one or more configurations which are intended to power a specific NR experience. However, for troubleshooting it is sometimes
necessary to tweak the configuration to rule out certain issues. The collector supports configurations from more than one source and merges them into one
final runtime configuration, see also [OTel Collector documentation](https://opentelemetry.io/docs/collector/configuration/). This means you can tweak the
default configuration supplied with each distribution for troubleshooting purposes by supplying additional `--config` arguments to the collector but need to
keep in mind that 'last one wins' when it comes to configuration values.

There are different types of input you can use, each enabled by its own [provider](https://github.com/open-telemetry/opentelemetry-collector/tree/main/confmap/provider). As an example you can change the [log level](https://opentelemetry.io/docs/collector/internal-telemetry/#configure-internal-logs) of our default configuration (supplied via the default [fileprovider](https://github.com/open-telemetry/opentelemetry-collector/tree/main/confmap/provider/fileprovider)) via the [yamlprovider](https://github.com/open-telemetry/opentelemetry-collector/tree/main/confmap/provider/yamlprovider). It is important to note that you will still need to supply the default configuration (or your own) which can be easy to forgot depending on the deployment mechanism.

#### Binary
```bash
/usr/bin/nrdot-collector --config=/etc/nrdot-collector-host/config.yaml --config 'yaml:service::telemetry::logs::level: WARN'"
```

#### Docker
```bash
# docker without override implicitly adding `--config /etc/nrdot-collector-host/config.yaml` via CMD directive
docker run newrelic/nrdot-collector-host
# docker with config override
docker run newrelic/nrdot-collector-host --config /etc/nrdot-collector-host/config.yaml --config 'yaml:service::telemetry::logs::level: WARN'
```

#### Kubernetes
```yaml
# k8s daemonset example
image: newrelic/nrdot-collector-k8s
# args can be fully omitted if no override as `CMD` directive supplies the default config
args: [ "--config", "/etc/nrdot-collector-k8s/config-daemonset.yaml", "--config", "yaml:service::telemetry::logs::level: WARN" ]
```

#### Linux packages
Our linux packages (if available for the distro) are started as a `systemd` service, so you'd have to edit the `OTELCOL_OPTIONS` environment variable in the `.conf` file and restart the service via `systemctl` The exact location of this file and the default configuration is distribution-specific, but can be looked up in the `nfpms` section of the `goreleaser.yaml` in the respective distribution directory.
```
# edit and save /etc/nrdot-collector-host/nrdot-collector-host.conf
OTELCOL_OPTIONS="--config=/etc/nrdot-collector-host/config.yaml --config 'yaml:service::telemetry::logs::level: WARN'"
# restart NRDOT
systemctl reload-or-restart nrdot-collector-host.service
```


### Collector logs
Each component in the collector emits logs, enabled by the [internal telemetry configuration](https://opentelemetry.io/docs/collector/internal-telemetry/#configure-internal-logs) of the collector which are often the quickest way to determine the root cause of an issue.
By default, the log `level` is set to `INFO` which is usually sufficient for debugging, but you can use the above-mentioned instructions to override it to a value that suits your needs.
```
# Example warning log by hostmetricsreceiver
2025-01-01T23:19:05.097Z    warn    filesystemscraper/factory.go:48    No `root_path` config set when running in docker environment, will report container filesystem stats. See https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/hostmetricsreceiver#collecting-host-metrics-from-inside-a-container-linux-only    {"otelcol.component.id": "hostmetrics", "otelcol.component.kind": "Receiver", "otelcol.signal": "metrics"}
# Example of logs of successful startup and processing
2025-01-01T23:19:14.828Z    info    service@v0.121.0/service.go:281    Everything is ready. Begin running and processing data.
2025-01-01T23:19:16.031Z    info    Metrics    {"otelcol.component.id": "debug", "otelcol.component.kind": "Exporter", "otelcol.signal": "metrics", "resource metrics": 4, "metrics": 9, "data points": 32}
```
The logs are written to `stderr` unless you installed the collector as a linux package, then you'll have to access the logs via `journalctl | grep nrdot-collector`.

### Debugexporter
The [debugexporter](https://github.com/open-telemetry/opentelemetry-collector/blob/main/exporter/debugexporter/README.md) is a special exporter that logs details about all the telemetry it receives and then drops it. This can be very helpful to quickly validate changes to the configuration without
having to wait for data being ingested by New Relic. 
```
# Example log by debugexporter verifying telemetry is being processed
2025-02-28T20:16:08.620Z info Metrics {"otelcol.component.id": "debug", "otelcol.component.kind": "Exporter", "otelcol.signal": "metrics", "resource metrics": 4, "metrics": 9, "data points": 32}`
```
All NRDOT collector distributions include the `debugexporter` but disable it by default due to its verbosity and performance overhead. In order to enable it, you'll have to add it as a component and use it in the pipeline you're trying to debug.
```
# Configure debugexporter (empty config is valid) and use it in pipeline 'metrics'
--config /etc/nrdot-collector-k8s/config-daemonset.yaml --config 'yaml:exporters::debug: ' --config 'yaml:service::pipelines::metrics::exporters: [otlphttp/newrelic, debug]'
```
Additional configuration options to increase verbosity or enable sampling are available in the [exporter's docs](https://github.com/open-telemetry/opentelemetry-collector/blob/main/exporter/debugexporter/README.md#getting-started).

## Common Issues

### Collector not starting up
The collector validates the configuration on startup and fails to start if a provided value cannot be parsed as expected. Example when providing `100m` instead of `100` as `NEW_RELIC_MEMORY_LIMIT_MIB`:
```
Error: failed to get config: cannot unmarshal the configuration: decoding failed due to the following error(s):
error decoding 'processors': error reading configuration for "memory_limiter": decoding failed due to the following error(s):
'limit_mib' expected type 'uint32', got unconvertible type 'string', value: '100m'
```

<a id="stablelink-telemetry-not-reaching-new-relic"></a>
### Telemetry not reaching New Relic
If the UI does not light up as you expect it to and there are no [NrIntegrationError](https://docs.newrelic.com/docs/data-apis/ingest-apis/metric-api/troubleshoot-nrintegrationerror-events/), you can run some basic NRQL queries to check whether telemetry is reaching New Relic at all. If you do see some but not all the data you expect, there is either an issue with a specific pipeline or you might be running into [cardinality limits](https://docs.newrelic.com/docs/data-apis/ingest-apis/metric-api/NRQL-high-cardinality-metrics/).
```
# Metrics
FROM Metric SELECT * WHERE newrelic.source='api.metrics.otlp' WHERE otel.library.name like 'github.com/open-telemetry/opentelemetry-collector-contrib/receiver%' SINCE 1 hour ago

# Logs (if expected)
FROM Log SELECT * where newrelic.source='api.logs.otlp' SINCE 1 hour ago

# Traces (in the form of Spans - if expected)
FROM Span SELECT * where newrelic.source='api.traces.otlp' SINCE 1 hour ago
```

#### 1. Connection issues
Check the logs. As an example, a misconfigured `OTEL_EXPORTER_OTLP_ENDPOINT` would result in the following log. 
```
2025-01-01T23:35:50.906Z    info    internal/retry_sender.go:126    Exporting failed. Will retry the request after interval.
{"otelcol.component.id": "otlphttp", "otelcol.component.kind": "Exporter", "otelcol.signal": "metrics", "error": "failed to make an HTTP request:
Post \"url-missing-scheme/v1/metrics\": unsupported protocol scheme \"\"", "interval": "9.519061682s"}
```
For addressing general OTLP connection issues, please refer to our [OTLP troubleshooting guide](https://docs.newrelic.com/docs/opentelemetry/best-practices/opentelemetry-otlp-troubleshooting/).

#### 2. Check your config (if customized)
- If you see some but not all telemetry types, double-check your [pipelines](https://opentelemetry.io/docs/collector/configuration/#pipelines). Each telemetry type (logs, metrics, traces) has a separate pipeline, so make sure the missing telemetry has a pipeline with the appropriate receiver.
- Rule out processors dropping data. Add a debugging pipeline with just the receiver in question and the `debugexporter`. Add processors one at a time and ensure the `debugexporter` still observers logs any time you add one.
- Consult the receivers docs, e.g. [hostmetricsreceiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/hostmetricsreceiver/README.md). Some receivers need elevated access to scrape data (filesystem, credentials etc.). If they lack permissions, a `warn` log is usually written.