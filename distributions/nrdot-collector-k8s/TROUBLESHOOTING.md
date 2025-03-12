# Troubleshooting for nrdot-collector-k8s

For general NRDOT troubleshooting, see [this guide](../TROUBLESHOOTING.md). This document assumes you are familiar with
the troubleshooting tools mentioned.

## Known issues

### Missing permissions
There are many variations of this error due to all the different APIs the k8s components scrape. This is a log example indicating this issue:
```
  reflector.go:569] k8s.io/client-go@v0.32.2/tools/cache/reflector.go:251: failed to list *v1.Node: nodes is forbidden: User "system:serviceaccount:demo-3:default" cannot list resource "nodes" in API group "" at the cluster scope
```
The collector is missing permissions to access the cluster-internal k8s APIs. 
As mentioned in the [installation instructions](./README.md), we highly recommend NOT using this distro by itself but rather through our [helm-chart](https://github.com/newrelic/helm-charts/tree/master/charts/nr-k8s-otel-collector) which sets up the required permissions for you.

Otherwise, you will have to consult the documentation of the k8s receivers enabled in the configuration you are running, see the [manifest](./manifest.yaml) and [daemonset config](./config-daemonset.yaml) or [deployment config](./config-deployment.yaml). The error message unfortunately does not mention which component is causing the issue, so you will have to deduce from the 'resource' and 'scope' in the error message which receiver is causing the issue. As an example, the [k8seventsreceiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/k8seventsreceiver/README.md#service-account) has a section about setting up a `ServiceAccount` with the required permissions.