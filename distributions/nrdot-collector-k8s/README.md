# nrdot-collector-k8s

Note: See [general README](../README.md) for information that applies to all distributions.

A distribution of the NRDOT collector focused on gathering metrics in a kubernetes environment with two different configs:
- [config-daemonset.yaml](./config-daemonset.yaml) (default): Typically deployed as a `DaemonSet`. Collects node-level metrics via `hostmetricsreceiver`, `filelogreceiver`, `kubeletstatsreceiver` and `prometheusreceiver` (`cAdvisor`, `kubelet`).
- [config-deployment.yaml](./config-deployment.yaml): Typically deployed as a `Deployment`. Collects cluster-level metrics via `k8seventsreceiver`,  `prometheusreceiver` (`kube-state-metrics`, `apiserver`, `controller-manager`, `scheduler`). Can be enabled by overriding the default docker `CMD`, i.e. `--config /etc/nrdot-collector-k8s/config-deployment.yaml`.

Distribution is available as docker image and runs in `daemonset` mode by default.

## Additional Configuration

See [general README](../README.md) for information that applies to all distributions.

### nrdot-collector-k8s
| Environment Variable | Description | Default |
|---|---|---|
| `K8S_CLUSTER_NAME` | Kubernetes Cluster Name used to populate attributes like `k8s.cluster.name` | `cluster-name-placeholder` |
| `MY_POD_IP` | Pod IP to configure `otlpreceiver` | `cluster-name-placeholder` |