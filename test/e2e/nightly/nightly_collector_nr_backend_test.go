package hostmetrics

import (
	"fmt"
	"test/e2e/util/assert"
	envutil "test/e2e/util/env"
	"test/e2e/util/nr"
	"test/e2e/util/spec"
	testutil "test/e2e/util/test"
	"testing"
	"time"
)

var ec2Ubuntu22 = spec.NightlySystemUnderTest{
	TestKeyPattern: testutil.NewNrQueryHostNamePattern("nightly", testutil.Wildcard, "ec2_ubuntu22_04"),
	SkipIf: func(testSpec *spec.TestSpec) bool {
		return !testSpec.Nightly.EC2.Enabled
	},
}
var ec2Ubuntu24 = spec.NightlySystemUnderTest{
	TestKeyPattern: testutil.NewNrQueryHostNamePattern("nightly", testutil.Wildcard, "ec2_ubuntu24_04"),
	SkipIf: func(testSpec *spec.TestSpec) bool {
		return !testSpec.Nightly.EC2.Enabled
	},
}
var k8sNode = spec.NightlySystemUnderTest{
	TestKeyPattern: testutil.NewNrQueryHostNamePattern("nightly", testutil.Wildcard, "k8s_node"),
}

func TestNightlyCollectorWithNrBackend(t *testing.T) {
	testutil.TagAsNightlyTest(t)
	testSpec := spec.LoadTestSpec()

	client := nr.NewClient()

	for _, sut := range []spec.NightlySystemUnderTest{ec2Ubuntu22, ec2Ubuntu24, k8sNode} {
		if sut.SkipIf != nil && sut.SkipIf(testSpec) {
			t.Logf("Skipping nightly system-under-test: %s", sut.TestKeyPattern)
			continue
		}
		testEnvironment := map[string]string{
			"clusterName": envutil.GetK8sContextName(),
			"testKey":     sut.TestKeyPattern,
		}
		for _, testCaseSpecName := range testSpec.Nightly.TestCaseSpecs {
			testCaseSpec := spec.LoadTestCaseSpec(testCaseSpecName)

			// Allow overriding where clause in distro test specs
			if clause, exists := testSpec.WhereClause[testCaseSpecName]; exists {
				testCaseSpec.WhereClause = clause
			}

			if envutil.IsFipsMode() {
				testEnvironment["clusterNamespace"] = fmt.Sprintf("%s%s", envutil.GetK8sContextName(), "-fips")
				testCaseSpec.WhereClause = fmt.Sprintf("%s AND k8s.namespace.name='{{ .clusterNamespace }}'", testCaseSpec.WhereClause)
			}

			whereClause := testCaseSpec.RenderWhereClause(testEnvironment)
			for caseName, testCase := range testCaseSpec.GetTestCasesWithout(sut.ExcludedMetrics) {
				t.Run(fmt.Sprintf("%s/%s/%s", sut.TestKeyPattern, testCaseSpecName, caseName), func(t *testing.T) {
					t.Parallel()
					assertionFactory := assert.NewNrMetricAssertionFactory(
						whereClause,
						"2 hour ago",
					)
					assertion := assertionFactory.NewNrMetricAssertion(testCase.Metric, testCase.Assertions)
					assertion.ExecuteWithRetries(t, client, 50, 10*time.Second)
				})
			}
		}
	}
}
