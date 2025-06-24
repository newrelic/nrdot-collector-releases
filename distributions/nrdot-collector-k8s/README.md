# nrdot-collector-k8s

| Status    |                                                                                     |
|-----------|-------------------------------------------------------------------------------------|
| Distro    | `nrdot-collector-k8s`                                                               |
| Stability | `preview`                                                                           |
| Artifacts | [Docker images on DockerHub](https://hub.docker.com/r/newrelic/nrdot-collector-k8s) |

A distribution of the NRDOT collector focused on gathering metrics in a kubernetes environment.

Note: See [general README](../README.md) for information that applies to all distributions.

## Installation
The distribution's primary purpose is to be a building block for the [nr-k8s-otel-collector](https://github.com/newrelic/helm-charts/tree/master/charts/nr-k8s-otel-collector) helm chart which we recommend using. The helm chart takes care of all the configuration required to ensure a smooth operation of the collector and drive the NR Kubernetes experience, including but not limited to: deploying collector as daemonset and deployment with different configurations for node vs cluster-level metrics, wiring up necessary permissions via service accounts, adding [kube-state-metrics](https://github.com/kubernetes/kube-state-metrics) for additional scrapeable metrics etc. 
While you can use the `nrdot-collector-k8s` image directly, we do not provide support for this use case.

## Configuration

If you use the [nr-k8s-otel-collector](https://github.com/newrelic/helm-charts/tree/master/charts/nr-k8s-otel-collector) helm chart, please refer to its documentation for configuration options.

See [general README](../README.md) for information that applies to all distributions.

### Distribution-specific configuration

| Environment Variable | Description | Default |
|---|---|---|
| `K8S_CLUSTER_NAME` | Kubernetes Cluster Name used to populate attributes like `k8s.cluster.name` | `cluster-name-placeholder` |
| `MY_POD_IP` | Pod IP to configure `otlpreceiver` | `cluster-name-placeholder` |

## Limitations of the standalone image (vs helm chart)
Key differences are formally documented in [the script that ensures the configurations stay in-sync](./sync-configs.sh) and can be summarized as:
- `lowDataMode: false` is hardcoded and cannot be toggled easily (it is always possible to overwrite the config with native collector options)
- `hostmetricsreceiver` in the daemonset config does not assume that the [host file system is mounted](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/hostmetricsreceiver/README.md#collecting-host-metrics-from-inside-a-container-linux-only), thus providing metrics about its container and not the host node.
- `healthcheckextension` is enabled by default