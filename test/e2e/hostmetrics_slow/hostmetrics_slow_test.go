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
	TestNamespace = "nr-hostmetrics"
)

var (
	kubectlOptions *k8s.KubectlOptions
	testChart      chart.NrBackendChart
)

func TestSlow(t *testing.T) {
	testutil.TagAsSlowTest(t)
	testSpec := spec.LoadTestSpec()

	kubectlOptions = k8sutil.NewKubectlOptions(TestNamespace)
	testId := testutil.NewTestId()
	testChart = chart.NewNrBackendChart(testId)

	t.Logf("hostname used for test: %s", testChart.NrQueryHostNamePattern)
	helmutil.ApplyChart(t, kubectlOptions, testChart.AsChart(), "hostmetrics-startup", testId)
	k8sutil.WaitForCollectorReady(t, kubectlOptions)
	// wait for at least one default metric harvest cycle (60s) and some buffer to allow NR ingest to process data
	time.Sleep(70 * time.Second)
	// space out requests to not run into 25 concurrent request limit
	requestsPerSecond := 4.0
	requestSpacing := time.Duration((1/requestsPerSecond)*1000) * time.Millisecond
	client := nr.NewClient()

	testEnvironment := map[string]string{
		"hostName":    testChart.NrQueryHostNamePattern,
		"clusterName": kubectlOptions.ContextName,
	}
	for _, testCaseSpecName := range testSpec.Slow.TestCaseSpecs {
		testCaseSpec := spec.LoadTestCaseSpec(testCaseSpecName)
		whereClause := testCaseSpec.RenderWhereClause(testEnvironment)
		t.Logf("test case spec where clause: %s", whereClause)

		counter := 0
		for caseName, testCase := range testCaseSpec.TestCases {
			t.Run(fmt.Sprintf("%s/%s", testCaseSpecName, caseName), func(t *testing.T) {
				t.Parallel()
				assertionFactory := assert.NewNrMetricAssertionFactory(
					whereClause,
					"5 minutes ago",
				)
				assertion := assertionFactory.NewNrMetricAssertion(testCase.Metric, testCase.Assertions)
				// space out requests to avoid rate limiting
				time.Sleep(time.Duration(counter) * requestSpacing)
				assertion.ExecuteWithRetries(t, client, 24, 5*time.Second)
			})
			counter += 1
		}
	}
}
