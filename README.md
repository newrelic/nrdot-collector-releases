<a href="https://opensource.newrelic.com/oss-category/#community-project"><picture><source media="(prefers-color-scheme: dark)" srcset="https://github.com/newrelic/opensource-website/raw/main/src/images/categories/dark/Community_Project.png"><source media="(prefers-color-scheme: light)" srcset="https://github.com/newrelic/opensource-website/raw/main/src/images/categories/Community_Project.png"><img alt="New Relic Open Source community project banner." src="https://github.com/newrelic/opensource-website/raw/main/src/images/categories/Community_Project.png"></picture></a>

# New Relic Distribution of OpenTelemetry (NRDOT) Releases 

This repository assembles various [custom distributions](https://opentelemetry.io/docs/collector/distributions/#custom-distributions) of the [OpenTelemetry Collector](https://opentelemetry.io/docs/collector/) focused on specific use cases and pre-configured to work with NewRelic out-of-the-box.

Generated assets are available in the corresponding Github release page and as docker images published within the [newrelic organization on Docker Hub](https://hub.docker.com/u/newrelic).

Current list of distributions:

- [nrdot-collector](./distributions/nrdot-collector/): comprehensive core distribution with full OTLP gateway capabilities, host monitoring, and Prometheus scraping.
- [nrdot-collector-host](./distributions/nrdot-collector-host/): distribution focused on monitoring host metrics and logs
- [nrdot-collector-k8s](./distributions/nrdot-collector-k8s/): distribution focused on monitoring a Kubernetes cluster

Please refer to [this README](./distributions/README.md) for documentation.

## Distribution Governance
Our intention is that each distribution we maintain serves a specific use case powering a particular slice of the New Relic experience. This ensures that each distribution is easy to understand, has a small attack surface and can be updated quickly to patch security issues or bugs.
However, this also means that we need to be deliberate about the components we include in a distribution. Our current
philosophy is that we will only add a new component to a distribution if
- it is required to support the use case of the distribution
- we consider it [essential ](./distributions/core-components.md) for all distributions
Our goal is to work with customers and internal teams to iteratively create new distributions when a strong enough use case has been developed and the components to support it have been identified.

Please note that while the set of distributions is still limited, we encourage you to also explore the distributions provided by the [OpenTelemetry community](https://github.com/open-telemetry/opentelemetry-collector-releases/tree/main/distributions). In particular the `contrib` distribution can be helpful as a stopgap solution as it includes all core and contrib components.

## Support

New Relic hosts and moderates an online forum where customers can interact with New Relic employees as well as other customers to get help and share best practices. You can find this project's topic/threads here: [New Relic Community](https://forum.newrelic.com).

## Contribute

We encourage your contributions to improve the New Relic OpenTelemetry collector! Keep in mind that when you submit your pull request, you'll need to sign the CLA via the click-through using CLA-Assistant. You only have to sign the CLA one time per project.

If you have any questions, or to execute our corporate CLA (which is required if your contribution is on behalf of a company), drop us an email at opensource@newrelic.com.

**A note about vulnerabilities**

As noted in our [security policy](../../security/policy), New Relic is committed to the privacy and security of our customers and their data. We believe that providing coordinated disclosure by security researchers and engaging with the security community are important means to achieve our security goals.

If you believe you have found a security vulnerability in this project or any of New Relic's products or websites, we welcome and greatly appreciate you reporting it to New Relic through [HackerOne](https://hackerone.com/newrelic).

If you would like to contribute to this project, review [these guidelines](./CONTRIBUTING.md).

To all contributors, we thank you!  Without your contribution, this project would not be what it is today.

## License
[Collector releases] is licensed under the [Apache 2.0](http://apache.org/licenses/LICENSE-2.0.txt) License.
