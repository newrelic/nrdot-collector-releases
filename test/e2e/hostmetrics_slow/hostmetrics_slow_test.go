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

func TestStartupBehavior(t *testing.T) {
	testutil.TagAsSlowTest(t)
	kubectlOptions = k8sutil.NewKubectlOptions(TestNamespace)
	testId := testutil.NewTestId()
	testChart = chart.NewNrBackendChart(testId, kubectlOptions.Namespace)

	t.Logf("hostname used for test: %s", testChart.NrQueryHostNamePattern)
	helmutil.ApplyChart(t, kubectlOptions, testChart.AsChart(), "hostmetrics-startup", testId)
	k8sutil.WaitForCollectorReady(t, kubectlOptions)
	// wait for at least one default metric harvest cycle (60s) and some buffer to allow NR ingest to process data
	time.Sleep(70 * time.Second)
	// space out requests to not run into 25 concurrent request limit
	requestsPerSecond := 4.0
	requestSpacing := time.Duration((1/requestsPerSecond)*1000) * time.Millisecond
	client := nr.NewClient()

	for i, testCase := range spec.GetOnHostTestCases() {
		t.Run(fmt.Sprintf(testCase.Name), func(t *testing.T) {
			t.Parallel()
			assertionFactory := assert.NewNrMetricAssertionFactory(
				fmt.Sprintf("WHERE host.name like '%s'", testChart.NrQueryHostNamePattern),
				"5 minutes ago",
			)
			assertion := assertionFactory.NewNrMetricAssertion(testCase.Metric, testCase.Assertions)
			// space out requests to avoid rate limiting
			time.Sleep(time.Duration(i) * requestSpacing)
			assertion.ExecuteWithRetries(t, client, 15, 5*time.Second)
		})
	}
}
