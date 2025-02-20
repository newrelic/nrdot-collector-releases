package assert

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/newrelic/newrelic-client-go/v2/newrelic"
	"github.com/newrelic/newrelic-client-go/v2/pkg/nrdb"
	"log"
	"reflect"
	envutil "test/e2e/util/env"
	"test/e2e/util/spec"
	"testing"
	"text/template"
	"time"
)

type NrMetricAssertionFactory struct {
	entityWhereClause string
	since             string
	until             string
}

func NewNrMetricAssertionFactory(entityWhereClause string, since string) NrMetricAssertionFactory {
	return NrMetricAssertionFactory{
		entityWhereClause: entityWhereClause,
		since:             since,
		until:             "now",
	}
}

func (f *NrMetricAssertionFactory) NewNrMetricAssertion(metric spec.NrMetric, assertions []spec.NrAssertion) NrMetricAssertion {
	return NrMetricAssertion{
		Query: NrBaseQuery{
			Metric:            metric,
			EntityWhereClause: f.entityWhereClause,
			Since:             f.since,
			Until:             f.until,
		},
		Assertions: assertions,
	}
}

type NrMetricAssertion struct {
	Query      NrBaseQuery
	Assertions []spec.NrAssertion
}

type NrBaseQuery struct {
	Metric            spec.NrMetric
	EntityWhereClause string
	Since             string
	Until             string
}

func (m *NrMetricAssertion) ExecuteWithRetries(t testing.TB, client *newrelic.NewRelic, retries int, retryInterval time.Duration) {
	for attempt := 0; attempt < retries; attempt++ {
		if attempt > 0 {
			time.Sleep(retryInterval)
		}
		err := m.execute(client)
		if err == nil {
			return
		} else {
			t.Logf("Assertion attempt %d failed with: %+v", attempt, err)
		}
	}
	t.Fatalf("Assertions failed after %d attempts", retries)
}

func (m *NrMetricAssertion) execute(client *newrelic.NewRelic) error {
	query := nrdb.NRQL(m.AsQuery())
	successfulAssertions := 0
	response, err := client.Nrdb.Query(envutil.GetNrAccountId(), query)
	if err != nil {
		return err
	}
	if len(response.Results) != len(m.Assertions) {
		return errors.New(fmt.Sprintf("query results (%+v) and number of assertions (%+v) do not match", response.Results, m.Assertions))
	}
	for i, assertion := range m.Assertions {
		actualContainer := response.Results[i]
		actualValue := actualContainer[assertion.AggregationFunction+"."+m.Query.Metric.Name]
		valueType := reflect.TypeOf(actualValue)
		if valueType == nil {
			return errors.New("received nil, metric might not be ingested yet")
		}
		if valueType.Kind() != reflect.Float64 {
			return errors.New(fmt.Sprintf("Expected float64 for assertion %+v, but received %+v in response %+v. Retrying...", actualContainer, valueType, response.Results))
		}
		if !satisfiesCondition(assertion, actualValue.(float64)) {
			return errors.New(fmt.Sprintf("Expected %s(%s) %s %f, but received %f", assertion.AggregationFunction, m.Query.Metric.Name, assertion.ComparisonOperator, assertion.Threshold, actualValue))
		}
		successfulAssertions += 1
	}
	if successfulAssertions != len(m.Assertions) {
		return errors.New(fmt.Sprintf("Expected %d successful assertions, but received %d. Check logs for more details", len(m.Assertions), successfulAssertions))
	}
	return nil
}

func satisfiesCondition(assertion spec.NrAssertion, actualValue float64) bool {
	switch assertion.ComparisonOperator {
	case ">":
		return actualValue > assertion.Threshold
	case ">=":
		return actualValue >= assertion.Threshold
	default:
		log.Panicf("Unknown comparison operator: %s", assertion.ComparisonOperator)
		return false
	}
}

func (m *NrMetricAssertion) AsQuery() string {
	tmpl, err := template.New("query").Parse(`
SELECT {{ range $idx, $assert := .Assertions -}}
	{{- if $idx }},{{ end }}{{ $assert.AggregationFunction }}(` + "`" + `{{ $.Query.Metric.Name }}` + "`" + `)
{{- end }}
FROM Metric
{{ .Query.Metric.WhereClause }}
{{ .Query.EntityWhereClause }}
SINCE {{ .Query.Since }} UNTIL {{ .Query.Until }}
`)
	if err != nil {
		log.Panicf("Couldn't parse template: %s", err)
	}
	query := new(bytes.Buffer)
	err = tmpl.Execute(query, m)
	if err != nil {
		log.Panicf("Couldn't execute template using %+v: %s", m, err)
	}
	return query.String()
}
