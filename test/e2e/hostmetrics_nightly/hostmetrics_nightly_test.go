package hostmetrics

import (
	"fmt"
	"test/e2e/util/assert"
	"test/e2e/util/nr"
	"test/e2e/util/spec"
	testutil "test/e2e/util/test"
	"testing"
	"time"
)

var ec2Ubuntu22 = spec.NightlySystemUnderTest{
	HostNamePattern: testutil.NewNrQueryHostNamePattern("nightly", testutil.Wildcard, "ec2_ubuntu22_04"),
	// TODO: NR-362121
	ExcludedMetrics: []string{"system.paging.usage"},
	SkipIf: func(testSpec *spec.TestSpec) bool {
		return !testSpec.Nightly.EC2.Enabled
	},
}
var ec2Ubuntu24 = spec.NightlySystemUnderTest{
	HostNamePattern: testutil.NewNrQueryHostNamePattern("nightly", testutil.Wildcard, "ec2_ubuntu24_04"),
	// TODO: NR-362121
	ExcludedMetrics: []string{"system.paging.usage"},
	SkipIf: func(testSpec *spec.TestSpec) bool {
		return !testSpec.Nightly.EC2.Enabled
	},
}
var k8sNode = spec.NightlySystemUnderTest{
	HostNamePattern: testutil.NewNrQueryHostNamePattern("nightly", testutil.Wildcard, "k8s_node"),
	ExcludedMetrics: []string{"system.paging.usage"},
}

func TestNightly(t *testing.T) {
	testutil.TagAsNightlyTest(t)
	testSpec := spec.LoadTestSpec()

	// space out requests to not run into 25 concurrent request limit
	requestsPerSecond := 4.0
	requestSpacing := time.Duration((1/requestsPerSecond)*1000) * time.Millisecond
	client := nr.NewClient()

	for _, sut := range []spec.NightlySystemUnderTest{ec2Ubuntu22, ec2Ubuntu24, k8sNode} {
		if sut.SkipIf != nil && sut.SkipIf(testSpec) {
			t.Logf("Skipping nightly system-under-test: %s", sut.HostNamePattern)
			continue
		}
		testEnvironment := map[string]string{
			"hostName": sut.HostNamePattern,
		}
		for _, testCaseSpecName := range testSpec.Nightly.TestCaseSpecs {
			testCaseSpec := spec.LoadTestCaseSpec(testCaseSpecName)
			whereClause := testCaseSpec.RenderWhereClause(testEnvironment)
			counter := 0
			for caseName, testCase := range testCaseSpec.GetTestCasesWithout(sut.ExcludedMetrics) {
				t.Run(fmt.Sprintf("%s/%s/%s", sut.HostNamePattern, testCaseSpecName, caseName), func(t *testing.T) {
					t.Parallel()
					assertionFactory := assert.NewNrMetricAssertionFactory(
						whereClause,
						"2 hour ago",
					)
					assertion := assertionFactory.NewNrMetricAssertion(testCase.Metric, testCase.Assertions)
					// space out requests to avoid rate limiting
					time.Sleep(time.Duration(counter) * requestSpacing)
					assertion.ExecuteWithRetries(t, client, 15, 5*time.Second)
				})
				counter += 1
			}
		}
	}
}
