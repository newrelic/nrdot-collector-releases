package spec

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	testutil "test/e2e/util/test"
	"text/template"
)

func LoadTestCaseSpec(testCaseSpecFile string) TestCaseSpec {
	filePath := testutil.NewPathRelativeToRootDir(fmt.Sprintf("test/e2e/util/spec/%s.yaml", testCaseSpecFile))
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	var testCaseSpec TestCaseSpec
	err = yaml.Unmarshal(data, &testCaseSpec)
	return testCaseSpec
}

type TestCaseSpec struct {
	WhereClause RenderableTemplate  `yaml:"whereClause"`
	TestCases   map[string]TestCase `yaml:"testCases"`
}

type RenderableTemplate struct {
	Template string   `yaml:"template"`
	Vars     []string `yaml:"vars"`
}

func (t *TestCaseSpec) RenderWhereClause(vars map[string]string) string {
	tmpl, err := template.New("whereClause").Parse(t.WhereClause.Template)
	if err != nil {
		panic(err)
	}
	for _, requiredVar := range t.WhereClause.Vars {
		if _, ok := vars[requiredVar]; !ok {
			panic(fmt.Errorf("missing required variable '%s' for where clause", requiredVar))
		}
	}
	rendered := new(bytes.Buffer)
	err = tmpl.Execute(rendered, vars)
	if err != nil {
		panic(err)
	}
	return rendered.String()
}

func (t *TestCaseSpec) GetTestCasesWithout(excludedMetrics []string) map[string]TestCase {
	var filteredTestCases map[string]TestCase
	for testCaseName, testCase := range t.TestCases {
		included := true
		for _, skippedMetric := range excludedMetrics {
			if testCase.Metric.Name == skippedMetric {
				included = false
			}
		}
		if included {
			filteredTestCases[testCaseName] = testCase
		}
	}
	return filteredTestCases
}

type TestCase struct {
	Metric     NrMetric      `yaml:"metric"`
	Assertions []NrAssertion `yaml:"assertions"`
}

type NrAssertion struct {
	AggregationFunction string  `yaml:"aggregationFunction"`
	ComparisonOperator  string  `yaml:"comparisonOperator"`
	Threshold           float64 `yaml:"threshold"`
}

type NrMetric struct {
	Name        string `yaml:"name"`
	WhereClause string `yaml:"whereClause"`
}
