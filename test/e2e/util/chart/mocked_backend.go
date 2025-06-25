package chart

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	envutil "test/e2e/util/env"
)

const (
	mockedBackendChartShortName = "mocked_backend"
)

type MockedBackendChart struct {
	version string
}

func newMockedBackendChart(version string) Chart {
	return MockedBackendChart{
		version: version,
	}
}

func (m MockedBackendChart) Meta() Meta {
	return Meta{
		name: "test/charts/mocked_backend",
	}
}

func (m MockedBackendChart) RequiredChartValues(_testId string) map[string]string {
	return map[string]string{
		"image.repository": fmt.Sprintf("newrelic/%s", envutil.GetDistro()),
		"image.tag":        envutil.GetImageTag(),
		"clusterName":      envutil.GetK8sContextName(),
	}
}

func (m MockedBackendChart) Version() string {
	return m.version
}

func (m MockedBackendChart) WaitUntilPodReadySelector() metav1.ListOptions {
	return metav1.ListOptions{
		LabelSelector: "app=nrdot-collector-daemonset",
	}
}

func (m MockedBackendChart) CollectorContainerName() string {
	return "nrdot-collector-daemonset"
}
