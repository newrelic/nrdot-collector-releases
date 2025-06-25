package chart

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	envutil "test/e2e/util/env"
)

const (
	NrRepoUrl                        = "https://helm-charts.newrelic.com"
	NrRepoName                       = "newrelic"
	nrK8sOtelCollectorChartFullName  = "newrelic/nr-k8s-otel-collector"
	nrK8sOtelCollectorChartShortName = nrK8sOtelCollectorChartFullName
)

type NrK8sOtelCollector struct {
	version string
	testId  string
}

func newNrK8sOtelCollector(version string) Chart {
	return NrK8sOtelCollector{
		version: version,
	}
}

func (m NrK8sOtelCollector) Meta() Meta {
	return Meta{
		name: nrK8sOtelCollectorChartFullName,
	}
}

func (m NrK8sOtelCollector) RequiredChartValues(_testId string) map[string]string {
	var nrStaging = "false"
	if strings.Contains(envutil.GetNrBackendUrl(), "staging") {
		nrStaging = "true"
	}
	return map[string]string{
		"image.repository": envutil.GetImageRepo(),
		"image.tag":        envutil.GetImageTag(),
		"licenseKey":       envutil.GetNrIngestKey(),
		"cluster":          envutil.GetK8sContextName(),
		"lowDataMode":      "false",
		"nrStaging":        nrStaging,
	}
}

func (m NrK8sOtelCollector) Version() string {
	return m.version
}

func (m NrK8sOtelCollector) WaitUntilPodReadySelector() metav1.ListOptions {
	return metav1.ListOptions{
		LabelSelector: "app.kubernetes.io/name=nr-k8s-otel-collector,component=daemonset",
	}
}

func (m NrK8sOtelCollector) CollectorContainerName() string {
	return "otel-collector-daemonset"
}
