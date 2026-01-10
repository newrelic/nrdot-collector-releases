# Component Inventory

This document maps each component in the `nrdot-collector` distribution to the use cases that require it. As use cases might evolve and begin to use components that are already present in the distro, this list might become stale and need updating, so it's not meant to be an authoritative source but best-effort context for maintainers.

**Legend:**
- **Core**: Core (not use-case specific)
- **Host**: Host Monitoring
- **Gateway**: Gateway Mode
- **OHI**: On-Host Integrations

## Receivers

| Component | Use Cases |
|-----------|-----------|
| `dockerstatsreceiver` | OHI |
| `elasticsearchreceiver` | OHI |
| `filelogreceiver` | Host |
| `hostmetricsreceiver` | Host, OHI |
| `jmxreceiver` | OHI |
| `k8seventsreceiver` | OHI |
| `kafkametricsreceiver` | OHI |
| `kubeletstatsreceiver` | OHI |
| `nginxreceiver` | OHI |
| `otlpreceiver` | Core |
| `prometheusreceiver` | Gateway |
| `rabbitmqreceiver` | OHI |
| `receivercreator` | OHI |

## Processors

| Component | Use Cases |
|-----------|-----------|
| `attributesprocessor` | Core |
| `batchprocessor` | Core |
| `cumulativetodeltaprocessor` | Core |
| `filterprocessor` | Core |
| `groupbyattrsprocessor` | Gateway |
| `memorylimiterprocessor` | Core |
| `metricstransformprocessor` | Core |
| `resourcedetectionprocessor` | Core |
| `resourceprocessor` | Gateway |
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
