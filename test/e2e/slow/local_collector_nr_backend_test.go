package hostmetrics

import (
	"fmt"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"test/e2e/util/assert"
	"test/e2e/util/chart"
	helmutil "test/e2e/util/helm"
	k8sutil "test/e2e/util/k8s"
	"test/e2e/util/nr"
	"test/e2e/util/spec"
	testutil "test/e2e/util/test"
	"testing"
	"time"
)

const (
	TestNamespace = "e2e-slow"
)

var (
	kubectlOptions *k8s.KubectlOptions
	testChart      chart.Chart
)

func TestLocalCollectorWithNrBackend(t *testing.T) {
	testutil.TagAsSlowTest(t)
	testSpec := spec.LoadTestSpec()

	kubectlOptions = k8sutil.NewKubectlOptions(TestNamespace)
	testId := testutil.NewTestId()
	testChart = chart.GetSlowTestChart(testSpec, testId)

	hostnamePattern := nr.GetHostNamePattern(testId)
	t.Logf("hostname used for test: %s", hostnamePattern)
	helmutil.ApplyChart(t, kubectlOptions, testChart, "slow", testId)
	k8sutil.WaitForCollectorReady(t, kubectlOptions, testChart.WaitUntilPodReadySelector(), testChart.CollectorContainerName())
	// wait a bit for collector startup sequence to complete and check for warn logs (in case log buffer is filled up by end of test)
	time.Sleep(10 * time.Second)
	k8sutil.FailOnUnexpectedWarningLogs(t, kubectlOptions, testChart.WaitUntilPodReadySelector(), testChart.CollectorContainerName(), testSpec)
	t.Cleanup(func() {
		k8sutil.FailOnUnexpectedWarningLogs(t, kubectlOptions, testChart.WaitUntilPodReadySelector(), testChart.CollectorContainerName(), testSpec)
	})
	time.Sleep(60 * time.Second)

	client := nr.NewClient()
	testEnvironment := map[string]string{
		"clusterName": kubectlOptions.ContextName,
		"hostName":    hostnamePattern,
	}

	for _, testCaseSpecName := range testSpec.Slow.TestCaseSpecs {
		testCaseSpec := spec.LoadTestCaseSpec(testCaseSpecName)

		// Allow overriding where clause in distro test specs
		if clause, exists := testSpec.WhereClause[testCaseSpecName]; exists {
			testCaseSpec.WhereClause = clause
		}

		whereClause := testCaseSpec.RenderWhereClause(testEnvironment)
		t.Logf("test case spec where clause: %s", whereClause)

		for caseName, testCase := range testCaseSpec.TestCases {
			t.Run(fmt.Sprintf("%s/%s", testCaseSpecName, caseName), func(t *testing.T) {
				t.Parallel()
				assertionFactory := assert.NewNrMetricAssertionFactory(
					whereClause,
					"5 minutes ago",
				)
				assertion := assertionFactory.NewNrMetricAssertion(testCase.Metric, testCase.Assertions)
				assertion.ExecuteWithRetries(t, client, 24, 5*time.Second)
			})
		}
	}
}
