# nrdot-collector-k8s

| Status    |                                                                                     |
|-----------|-------------------------------------------------------------------------------------|
| Distro    | `nrdot-collector-k8s`                                                               |
| Stability | `preview`                                                                           |
| Artifacts | [Docker images on DockerHub](https://hub.docker.com/r/newrelic/nrdot-collector-k8s) |

Note: See [general README](../README.md) for information that applies to all distributions.

A distribution of the NRDOT collector focused on gathering metrics in a kubernetes environment with two different configs:
- [config-daemonset.yaml](./config-daemonset.yaml) (default): Typically deployed as a `DaemonSet`. Collects node-level metrics via `hostmetricsreceiver`, `filelogreceiver`, `kubeletstatsreceiver` and `prometheusreceiver` (`cAdvisor`, `kubelet`).
- [config-deployment.yaml](./config-deployment.yaml): Typically deployed as a `Deployment`. Collects cluster-level metrics via `k8seventsreceiver`,  `prometheusreceiver` (`kube-state-metrics`, `apiserver`, `controller-manager`, `scheduler`). Can be enabled by overriding the default docker `CMD`, i.e. `--config /etc/nrdot-collector-k8s/config-deployment.yaml`.

## Installation
The distribution's main purpose is to be a building block for the [nr-k8s-otel-collector](https://github.com/newrelic/helm-charts/tree/master/charts/nr-k8s-otel-collector) helm chart which we recommend using. The helm chart takes care of a lot of configuration required to ensure a smooth operation of the collector and drive the NR Kubernetes experience.
If you choose not to use the helm chart, you'll have to follow the [general installation](../README.md#installation) and provide the necessary permissions for the collector to access the k8s APIs yourself, see also our [troubleshooting guide](./TROUBLESHOOTING.md).

### Dependencies
- Most k8s APIs scraped by the various receivers require additional permissions setup which are provided by the [nr-k8s-otel-collector](https://github.com/newrelic/helm-charts/tree/master/charts/nr-k8s-otel-collector) chart out of the box in the form of a service account. If you wish to add those permissions by hand, please refer to the chart itself.
- [kube-state-metrics](https://github.com/kubernetes/kube-state-metrics) is required to be running in your cluster. The metrics emitted by this add-on are used to create NR entities for various k8s resources.

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