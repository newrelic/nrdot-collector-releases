# Core Components
This document describes the core components of the NRDOT distribution which should be included in all distributions.

## Receivers
| Component                                | Reason                                                                                       |
|------------------------------------------|----------------------------------------------------------------------------------------------|
| `otlpreceiver`                           | Basic OTLP-based gateway capabilities                                                        |
| `filelogreceiver`                        | Local file log collection for host monitoring                                               |
| `hostmetricsreceiver`                    | System metrics collection (CPU, memory, disk, network)                                      |
| `prometheusreceiver`                     | Prometheus metrics scraping for application monitoring                                      |

## Processors
| Component                                | Reason                                                                                       |
|------------------------------------------|----------------------------------------------------------------------------------------------|
| `batchprocessor`                         | Performance optimization                                                                     |
| `memorylimiterprocessor`                 | Reliability - Control over resource usage                                                    |
| `attributesprocessor`                    | Attribute manipulation and enrichment                                                       |
| `cumulativetodeltaprocessor`             | Convert cumulative metrics to delta for proper aggregation                                 |
| `filterprocessor`                        | Filter out unwanted telemetry data                                                         |
| `groupbyattrsprocessor`                  | Group and aggregate telemetry data by attributes                                            |
| `metricstransformprocessor`              | Transform metric names and attributes for compatibility                                     |
| `resourcedetectionprocessor`             | Automatic detection of resource attributes (cloud, host, etc.)                             |
| `resourceprocessor`                      | Resource attribute manipulation and standardization                                         |
| `spanprocessor`                          | Span attribute manipulation and sampling decisions                                          |
| `tailsamplingprocessor`                  | Intelligent sampling based on trace content and patterns                                   |
| `transformprocessor`                     | Advanced telemetry data transformation using OTTL                                          |

## Exporters
| Component                                | Reason                                                                                       |
|------------------------------------------|----------------------------------------------------------------------------------------------|
| `debugexporter`                          | Debugging, testing, config validation                                                        |
| `otlpexporter`                           | Required to write to NR OTLP endpoint via gRPC                                              |
| `otlphttpexporter`                       | Required to write to NR OTLP endpoint via HTTP                                              |
| `loadbalancingexporter`                  | Load balancing and failover across multiple backends                                       |

## Connectors
| Component                                | Reason                                                                                       |
|------------------------------------------|----------------------------------------------------------------------------------------------|
| `routingconnector`                       | Reduce config redundancy for complex pipelines, e.g. multiple NR accounts based on attributes |

## Extensions
| Component                                | Reason                                                                                       |
|------------------------------------------|----------------------------------------------------------------------------------------------|
| `healthcheckextension`                   | Reliability - Basic health check capabilities                                                |

## Providers
| Component                                | Reason                                                                                       |
|------------------------------------------|----------------------------------------------------------------------------------------------|
| `envprovider`                            | Configuration from environment variables                                                     |
| `fileprovider`                           | Configuration from local files                                                              |
| `httpprovider`                           | Configuration from HTTP endpoints                                                           |
| `httpsprovider`                          | Configuration from HTTPS endpoints                                                          |
| `yamlprovider`                           | Configuration from YAML sources                                                             |
