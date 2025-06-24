package chart

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	envutil "test/e2e/util/env"
	"test/e2e/util/nr"
)

const (
	nrBackendChartShortName = "nr_backend"
)

type NrBackendChart struct {
	version string
}

func newNrBackendChart(version string) Chart {
	return NrBackendChart{
		version: version,
	}
}

func (m NrBackendChart) Meta() Meta {
	return Meta{
		name: "test/charts/nr_backend",
	}
}

func (m NrBackendChart) RequiredChartValues(testId string) map[string]string {
	return map[string]string{
		"image.repository":     envutil.GetImageRepo(),
		"image.tag":            envutil.GetImageTag(),
		"secrets.nrBackendUrl": envutil.GetNrBackendUrl(),
		"secrets.nrIngestKey":  envutil.GetNrIngestKey(),
		"collector.hostname":   nr.GetHostNamePrefix(testId),
		"clusterName":          envutil.GetK8sContextName(),
	}
}

func (m NrBackendChart) Version() string {
	return m.version
}

func (m NrBackendChart) WaitUntilPodReadySelector() metav1.ListOptions {
	return metav1.ListOptions{
		LabelSelector: "app=nrdot-collector-daemonset",
	}
}

func (m NrBackendChart) CollectorContainerName() string {
	return "nrdot-collector-daemonset"
}
