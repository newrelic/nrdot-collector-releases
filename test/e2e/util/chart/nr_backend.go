package chart

import (
	"fmt"
	"log"
	osuser "os/user"
	envutil "test/e2e/util/env"
	testutil "test/e2e/util/test"
)

type NrBackendChart struct {
	collectorHostNamePrefix string
	NrQueryHostNamePattern  string
	namespace               string
}

func NewNrBackendChart(testId string) NrBackendChart {
	var environmentName string
	if envutil.IsContinuousIntegration() {
		environmentName = "ci"
	} else {
		user, err := osuser.Current()
		if err != nil {
			log.Panicf("Couldn't determine current user: %v", err)
		}
		environmentName = fmt.Sprintf("local_%s", user.Username)
	}
	hostNamePrefix := testutil.NewHostNamePrefix(environmentName, testId, "k8s_node")
	hostNamePattern := testutil.NewNrQueryHostNamePattern(environmentName, testId, "k8s_node")

	return NrBackendChart{
		collectorHostNamePrefix: hostNamePrefix,
		NrQueryHostNamePattern:  hostNamePattern,
	}
}

func (m *NrBackendChart) AsChart() Chart {
	var chart Chart = m
	return chart
}

func (m *NrBackendChart) Meta() Meta {
	return Meta{
		name: "nr_backend",
	}
}

func (m *NrBackendChart) RequiredChartValues() map[string]string {
	return map[string]string{
		"image.repository":     envutil.GetImageRepo(),
		"image.tag":            envutil.GetImageTag(),
		"secrets.nrBackendUrl": envutil.GetNrBackendUrl(),
		"secrets.nrIngestKey":  envutil.GetNrIngestKey(),
		"collector.hostname":   m.collectorHostNamePrefix,
	}
}
