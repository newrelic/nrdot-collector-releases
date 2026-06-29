# Component Inventory

This document maps each component in the `nrdot-collector` distribution to the use cases that require it. As use cases might evolve and begin to use components that are already present in the distro, this list might become stale and need updating, so it's not meant to be an authoritative source but best-effort context for maintainers.

**Legend:**
- **Core**: Core (not use-case specific)
- **Host**: Host Monitoring
- **k8s**: Kubernetes Monitoring
- **ATP**: Host Minitoring with process metrics
- **Gateway**: Gateway Mode
- **OHI**: On-Host Integrations (shared infrastructure used across multiple integrations)
- **OHI-{name}**: On-Host Integration specific to a product. The suffix indicates the target integration (e.g., `OHI-kafka` for Kafka monitoring).

## Receivers

| Component | Use Cases |
|-----------|-----------|
| `awsecscontainermetricsreceiver` | OHI-ecs |
| `dockerstatsreceiver` | OHI-docker |
| `elasticsearchreceiver` | OHI-elasticsearch |
| `filelogreceiver` | Host, k8s |
| `haproxyreceiver` | OHI-haproxy |
| `hostmetricsreceiver` | Host, k8s, OHI |
| `jmxreceiver` | OHI-kafka |
| `k8seventsreceiver` | OHI, k8s |
| `kafkametricsreceiver` | OHI-kafka |
| `kafkareceiver` | OHI-kafka |
| `kubeletstatsreceiver` | OHI, k8s |
| `nginxreceiver` | OHI-nginx |
| `otlpreceiver` | Core |
| `prometheusreceiver` | Gateway, k8s |
| `rabbitmqreceiver` | OHI-rabbitmq |
| `redisreceiver` | OHI-redis |
| `nroracledbreceiver` | OHI-oracle |
| `nrsqlserverreceiver` | OHI-sqlserver |
| `receivercreator` | OHI |

## Processors

| Component | Use Cases |
|-----------|-----------|
| `adaptivetelemetryprocessor` | ATP |
| `attributesprocessor` | Core |
| `batchprocessor` | Core |
| `cumulativetodeltaprocessor` | Core |
| `filterprocessor` | Core |
| `groupbyattrsprocessor` | k8s, Gateway |
| `k8sattributesprocessor` | k8s, ATP, OHI |
| `memorylimiterprocessor` | Core |
| `metricsgenerationprocessor` | k8s, ATP, OHI |
| `metricstransformprocessor` | Core |
| `resourcedetectionprocessor` | Core |
| `resourceprocessor` | k8s, Gateway |
| `spanprocessor` | Gateway |
| `tailsamplingprocessor` | Gateway |
| `transformprocessor` | Core |

## Exporters

| Component | Use Cases |
|-----------|-----------|
| `debugexporter` | Core |
| `loadbalancingexporter` | Gateway |
| `otlpexporter` | Core |
| `otlphttpexporter` | Core |

## Connectors

| Component | Use Cases |
|-----------|-----------|
| `routingconnector` | Core |

## Extensions

| Component | Use Cases |
|-----------|-----------|
| `healthcheckextension` | Core |
| `observer/dockerobserver` | OHI |
| `observer/hostobserver` | OHI |
| `observer/k8sobserver` | OHI |

## Providers

| Component | Use Cases |
|-----------|-----------|
| `envprovider` | Core |
| `fileprovider` | Core |
| `httpprovider` | Core |
| `httpsprovider` | Core |
| `yamlprovider` | Core |
| `aesprovider` | OHI-oracle, OHI-sqlserver |
| `secretsmanagerprovider` | OHI-oracle, OHI-sqlserver |
