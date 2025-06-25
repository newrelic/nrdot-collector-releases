package spec

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"os"
	"testing"
)

func TestLoadTestSpecWithWhereClause(t *testing.T) {
	expectedSpec := &TestSpec{
		WhereClause: map[string]RenderableTemplate{
			"example": {
				Template: "WHERE var='{{ .var1 }}'",
				Vars:     []string{"var1"},
			},
		},
		Slow: struct {
			CollectorChart CollectorChart `yaml:"collectorChart"`
			TestCaseSpecs  []string       `yaml:"testCaseSpecs"`
		}{
			TestCaseSpecs: []string{"test1", "test2"},
		},
		Nightly: struct {
			EC2 struct {
				Enabled bool `yaml:"enabled"`
			} `yaml:"ec2"`
			TestCaseSpecs []string `yaml:"testCaseSpecs"`
		}{
			EC2: struct {
				Enabled bool `yaml:"enabled"`
			}{
				Enabled: true,
			},
			TestCaseSpecs: []string{"nightlyTest1", "nightlyTest2"},
		},
	}

	testSpecFile := "testdata/test-spec-with-where.yaml"
	testSpecData, err := os.ReadFile(testSpecFile)
	if err != nil {
		t.Fatalf("failed to read test spec file: %v", err)
	}

	var testSpec TestSpec
	err = yaml.Unmarshal(testSpecData, &testSpec)
	if err != nil {
		t.Fatalf("failed to unmarshal test spec: %v", err)
	}

	assert.Equal(t, expectedSpec, &testSpec)
}

func TestLoadTestSpecWithoutWhereClause(t *testing.T) {
	expectedSpec := &TestSpec{
		Slow: struct {
			CollectorChart CollectorChart `yaml:"collectorChart"`
			TestCaseSpecs  []string       `yaml:"testCaseSpecs"`
		}{
			TestCaseSpecs: []string{"test1", "test2"},
		},
		Nightly: struct {
			EC2 struct {
				Enabled bool `yaml:"enabled"`
			} `yaml:"ec2"`
			TestCaseSpecs []string `yaml:"testCaseSpecs"`
		}{
			EC2: struct {
				Enabled bool `yaml:"enabled"`
			}{
				Enabled: true,
			},
			TestCaseSpecs: []string{"nightlyTest1", "nightlyTest2"},
		},
	}

	testSpecFile := "testdata/test-spec-without-where.yaml"
	testSpecData, err := os.ReadFile(testSpecFile)
	if err != nil {
		t.Fatalf("failed to read test spec file: %v", err)
	}

	var testSpec TestSpec
	err = yaml.Unmarshal(testSpecData, &testSpec)
	if err != nil {
		t.Fatalf("failed to unmarshal test spec: %v", err)
	}

	assert.Equal(t, expectedSpec, &testSpec)
}
