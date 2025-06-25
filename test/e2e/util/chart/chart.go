package chart

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"test/e2e/util/spec"
	testutil "test/e2e/util/test"
)

type Chart interface {
	Meta() Meta
	Version() string
	RequiredChartValues(testId string) map[string]string
	WaitUntilPodReadySelector() metav1.ListOptions
	CollectorContainerName() string
}

type Meta struct {
	name string
}

func (m Meta) FullyQualifiedChartName() string {
	if strings.HasPrefix(m.name, NrRepoName) {
		return m.name
	}
	return testutil.NewPathRelativeToRootDir(m.name)
}

var shortNameToFactory = map[string]func(version string) Chart{
	mockedBackendChartShortName:      newMockedBackendChart,
	nrBackendChartShortName:          newNrBackendChart,
	nrK8sOtelCollectorChartShortName: newNrK8sOtelCollector,
}

func GetFastTestChart(spec *spec.TestSpec) Chart {
	factory := shortNameToFactory[spec.Fast.CollectorChart.Name]
	return factory(spec.Fast.CollectorChart.Version)
}

func GetSlowTestChart(spec *spec.TestSpec, testId string) Chart {
	factory := shortNameToFactory[spec.Slow.CollectorChart.Name]
	return factory(spec.Fast.CollectorChart.Version)
}
