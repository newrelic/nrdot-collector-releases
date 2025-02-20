package assert

import (
	"fmt"
	"strings"
	"test/e2e/util/spec"
	"testing"
)

func TestAsQueryWithSingleAssertion(t *testing.T) {
	assertionFactory := NewNrMetricAssertionFactory(
		fmt.Sprintf("WHERE host.name = 'nrdot-collector-host-foobar'"),
		"5 minutes ago",
	)
	singleAssertion := assertionFactory.NewNrMetricAssertion(
		spec.NrMetric{Name: "system.cpu.utilization", WhereClause: "WHERE state='user'"}, []spec.NrAssertion{
			{AggregationFunction: "max", ComparisonOperator: ">", Threshold: 0},
		})
	actual := singleAssertion.AsQuery()
	assertEqual(actual, `
SELECT max(^system.cpu.utilization^)
FROM Metric
WHERE state='user'
WHERE host.name = 'nrdot-collector-host-foobar'
SINCE 5 minutes ago UNTIL now
`, t)
}

func TestAsQueryWithMultipleAssertions(t *testing.T) {
	assertionFactory := NewNrMetricAssertionFactory(
		fmt.Sprintf("WHERE host.name = 'nrdot-collector-host-foobar'"),
		"5 minutes ago",
	)
	singleAssertion := assertionFactory.NewNrMetricAssertion(spec.NrMetric{Name: "system.cpu.utilization", WhereClause: "WHERE state='user'"}, []spec.NrAssertion{
		{AggregationFunction: "max", ComparisonOperator: "<", Threshold: 0},
		{AggregationFunction: "min", ComparisonOperator: ">", Threshold: 0},
		{AggregationFunction: "average", ComparisonOperator: ">", Threshold: 0},
	})
	actual := singleAssertion.AsQuery()
	assertEqual(actual, `
SELECT max(^system.cpu.utilization^),min(^system.cpu.utilization^),average(^system.cpu.utilization^)
FROM Metric
WHERE state='user'
WHERE host.name = 'nrdot-collector-host-foobar'
SINCE 5 minutes ago UNTIL now
`, t)
}

func assertEqual(actual string, expected string, t *testing.T) {
	actualTrimmed := strings.TrimSpace(actual)
	// no way to escape backticks, so we use '^' as a placeholder
	expectedTrimmed := strings.Replace(strings.TrimSpace(expected), "^", "`", -1)
	if actualTrimmed != expectedTrimmed {
		t.Fatalf("\nExpected:\n[%s]\nbut received:\n[%s]\n", expectedTrimmed, actualTrimmed)
	}
}
