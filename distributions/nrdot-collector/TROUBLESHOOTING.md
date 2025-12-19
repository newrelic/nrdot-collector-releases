# Troubleshooting for nrdot-collector

For general NRDOT troubleshooting, see [this guide](../TROUBLESHOOTING.md). This document assumes you are familiar with the troubleshooting tools mentioned.

## Host Monitoring

### Missing host entity in New Relic UI due to missing `host.id`
If you are [seeing telemetry getting ingested into New Relic](../TROUBLESHOOTING.md#user-content-stablelink-telemetry-not-reaching-new-relic) but the Host UI does not show host entities after a few minutes, you may be running into limitations of the [resourcedetectionprocessor](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/processor/resourcedetectionprocessor/README.md) NRDOT uses to determine the `host.id` attribute required to [synthesize a host entity](https://github.com/newrelic/entity-definitions/blob/main/entity-types/infra-host/definition.yml#L62-L63).

Example log messages indicating this issue:
```
# example 1
2025-01-01T22:49:09.110Z        warn    system/system.go:143    failed to get host ID   {"otelcol.component.id": "resourcedetection", "otelcol.component.kind" : "Processor", "otelcol.pipeline.id": "logs/host", "otelcol.signal": "logs", "error": "failed to obtain \"host.id\": error detecting resource: host id not found in: /etc/machine-id or /var/lib/dbus/machine-id"}
# example 2
2025-01-01T23:07:27.866Z        warn    system/system.go:143    failed to get host ID   {"otelcol.component.id": "resourcedetection", "otelcol.component.kind": "Processor", "otelcol.pipeline.id": "metrics/host", "otelcol.signal": "metrics", "error": "empty \"host.id\""}
```

**Resolution:** Set the `host.id` attribute manually via the [environment variable](./README.md#use-case-specific-configuration) `OTEL_RESOURCE_ATTRIBUTES`:
```bash
export OTEL_RESOURCE_ATTRIBUTES='host.id=my-custom-host-id'
```

### No `root_path` in containerized environments
The `hostmetricsreceiver` auto-detects files to scrape system metrics from. When running in a container, this causes the receiver to scrape metrics of the container instead of the host system. The receiver provides the `root_path` option to specify the path where the host file system is mounted into the container.

Warning message indicating this issue:
```
2025-01-01T21:08:21.097Z	warn	filesystemscraper/factory.go:48	No `root_path` config set when running in docker environment, will report container filesystem stats. See https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/hostmetricsreceiver#collecting-host-metrics-from-inside-a-container-linux-only	{"otelcol.component.id": "hostmetrics", "otelcol.component.kind": "Receiver", "otelcol.signal": "metrics"}
```

**Resolution:** Follow the [receiver's docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/hostmetricsreceiver/README.md#collecting-host-metrics-from-inside-a-container-linux-only) to mount the host file system and configure `root_path`:
```bash
docker run -v /:/hostfs \
-e NEW_RELIC_LICENSE_KEY='license-key' newrelic/nrdot-collector \
--config /etc/nrdot-collector/config.yaml \
--config 'yaml:receivers::hostmetrics::root_path: /hostfs'
```

## Gateway Mode

### High memory usage in gateway deployments
When running in gateway mode with high throughput, the collector may experience elevated memory usage. This is typically due to buffering and batching of telemetry data from multiple sources. To mitigate this:

1. **Adjust batch processor settings**: Configure the batch processor with appropriate timeout and batch size limits:
```bash
newrelic/nrdot-collector \
--config /etc/nrdot-collector/config.yaml \
--config 'yaml:processors::batch::timeout: 5s' \
--config 'yaml:processors::batch::send_batch_size: 1000'
```

2. **Configure memory limiter**: Use the memory limiter processor to prevent out-of-memory conditions:
```bash
newrelic/nrdot-collector \
--config /etc/nrdot-collector/config.yaml \
--config 'yaml:processors::memory_limiter::limit_mib: 512' \
--config 'yaml:processors::memory_limiter::spike_limit_mib: 128'
```

### Connection issues from remote collectors
When using the collector in gateway mode, remote collectors may have trouble connecting. Common causes include:

1. **Firewall or network policies**: Ensure the OTLP receiver ports (default: 4317 for gRPC, 4318 for HTTP) are accessible from remote collectors.

2. **TLS configuration**: If using TLS, ensure certificates are properly configured and trusted by remote collectors.

3. **Incorrect endpoint configuration**: Verify remote collectors are configured with the correct endpoint URL and protocol (gRPC vs HTTP).

### Load balancing considerations
For high-availability gateway deployments with multiple collector instances:

1. **Use consistent hashing**: When load balancing, use consistent hashing based on resource attributes to ensure related telemetry data is routed to the same collector instance.

2. **Configure health checks**: Set up appropriate health check endpoints for load balancers to detect unhealthy collector instances.

3. **Monitor queue sizes**: Keep an eye on internal queue sizes to detect backpressure issues early.