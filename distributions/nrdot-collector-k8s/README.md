# nrdot-collector-k8s

Note: See [general README](../README.md) for information that applies to all distributions.

A distribution of the NRDOT collector focused on gathering metrics in a kubernetes environment with two different configs:
- [config-daemonset.yaml](./config-daemonset.yaml) (default): Typically deployed as a `DaemonSet`. Collects node-level metrics via `hostmetricsreceiver`, `filelogreceiver`, `kubeletstatsreceiver` and `prometheusreceiver` (`cAdvisor`, `kubelet`).
- [config-deployment.yaml](./config-deployment.yaml): Typically deployed as a `Deployment`. Collects cluster-level metrics via `k8seventsreceiver`,  `prometheusreceiver` (`kube-state-metrics`, `apiserver`, `controller-manager`, `scheduler`). Can be enabled by overriding the default docker `CMD`, i.e. `--config /etc/nrdot-collector-k8s/config-deployment.yaml`.

Distribution is available as docker image and runs in `daemonset` mode by default.

## Additional Configuration

See [general README](../README.md) for information that applies to all distributions.

| Environment Variable | Description | Default |
|---|---|---|
| `K8S_CLUSTER_NAME` | Kubernetes Cluster Name used to populate attributes like `k8s.cluster.name` | `cluster-name-placeholder` |
| `MY_POD_IP` | Pod IP to configure `otlpreceiver` | `cluster-name-placeholder` |

## Distro vs Helm Chart
The initial choice of components and configuration of this distribution was driven by the [nr-k8s-otel-collector](https://github.com/newrelic/helm-charts/tree/master/charts/nr-k8s-otel-collector) helm chart. However, the helm templating syntax is more expressive than the collector configuration and should be used if possible. The configurations embedded in the distro are intended as a stable default with minimal setup for users who cannot use the helm chart but still want to monitor their k8s cluster with the NRDOT collector.

Key differences are formally documented in [the script that ensures the configurations stay in-sync](./sync-configs.sh) and can be summarized as:
- `lowDataMode: false` is hardcoded and cannot be toggled (easily - it is always possible to overwrite the config with native collector options)
- `hostmetricsreceiver` in the daemonset config does not assume that the [host file system is mounted](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/hostmetricsreceiver/README.md#collecting-host-metrics-from-inside-a-container-linux-only), thus providing metrics about its container and not the host node.