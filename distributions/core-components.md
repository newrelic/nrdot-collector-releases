# Core Components
This document describes the core components of the NRDOT distribution which should be included in all distributions.

| Component                                | Reason                                                                                       |
|------------------------------------------|----------------------------------------------------------------------------------------------|
| `otlpreceiver`                           | Basic OTLP-based gateway capabilities                                                        |
| `batchprocessor`                         | Performance optimization                                                                     |
| `memorylimiterprocessor`                 | Reliability - Control over resource usage                                                    |
| `routingconnector`                       | Reduce config redundancy for complex pipelines, e.g. multiple NR accounts based on attributes |
| `otlpexporter`                           | Required to write to NR OTLP endpoint via HTTP                                               |
| `otlphttpexporter`                       | Required to write to NR OTLP endpoint via gRPC                                               |
| `debugexporter`                          | Debugging, testing, config validation                                                        |
| `healthcheckextension`                   | Reliability - Basic health check capabilities                                                |
| `[env\|file\|http\|https\|yaml]provider` | Configuration from various sources |
