package chart

import (
	"fmt"
	envutil "test/e2e/util/env"
)

type MockedBackendChart struct {
}

func NewMockedBackendChart() MockedBackendChart {
	return MockedBackendChart{}
}

func (m *MockedBackendChart) AsChart() Chart {
	var chart Chart = m
	return chart
}

func (m *MockedBackendChart) Meta() Meta {
	return Meta{
		name: "mocked_backend",
	}
}

func (m *MockedBackendChart) RequiredChartValues() map[string]string {
	return map[string]string{
		"image.repository": fmt.Sprintf("newrelic/%s", envutil.GetDistro()),
		"image.tag":        envutil.GetImageTag(),
		"clusterName":      envutil.GetK8sContextName(),
	}
}
