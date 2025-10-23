### internal-telemetry-config.yaml

Version based on resource attribute `newrelic.collector_telemetry.version`

#### 0.4.0
- Disable tracing by default

#### 0.3.0
- Add resource attribute `newrelic.service.type: otel_collector`

#### 0.2.0
- Expose sampling / detail levels as environment variables
- Fix invalid header configuration