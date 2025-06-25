# nrdot-collector-k8s

| Status    |                                                                                     |
|-----------|-------------------------------------------------------------------------------------|
| Distro    | `nrdot-collector-k8s`                                                               |
| Stability | `preview`                                                                           |
| Artifacts | [Docker images on DockerHub](https://hub.docker.com/r/newrelic/nrdot-collector-k8s) |

A distribution of the NRDOT collector focused on gathering metrics in a kubernetes environment.

Note: See [general README](../README.md) for information that applies to all distributions.

## Installation
The distribution's primary purpose is to be a building block for the [nr-k8s-otel-collector](https://github.com/newrelic/helm-charts/tree/master/charts/nr-k8s-otel-collector) helm chart which we recommend using.
The helm chart takes care of all the configuration required to ensure a smooth operation of the collector and drive the NR Kubernetes experience, including but not limited to: deploying collector as daemonset and deployment with different configurations for node vs cluster-level metrics, wiring up necessary permissions via service accounts, adding [kube-state-metrics](https://github.com/kubernetes/kube-state-metrics) for additional scrapeable metrics etc.
While you can use the `nrdot-collector-k8s` image directly, we do not provide support for this use case. If your main goal is to avoid helm as a dependency, please refer to the chart's docs on [helmless installation](https://github.com/newrelic/helm-charts/tree/master/charts/nr-k8s-otel-collector#helmless-installation).
