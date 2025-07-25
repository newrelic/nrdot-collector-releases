package spec

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	envutil "test/e2e/util/env"
	testutil "test/e2e/util/test"
)

type CollectorChart struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}
type TestSpec struct {
	WhereClause map[string]RenderableTemplate `yaml:"whereClause"`
	Nightly     struct {
		EC2 struct {
			Enabled bool `yaml:"enabled"`
		} `yaml:"ec2"`
		TestCaseSpecs []string `yaml:"testCaseSpecs"`
	} `yaml:"nightly"`
}

func LoadTestSpec() *TestSpec {
	distro := envutil.GetDistro()
	testSpecFile := testutil.NewPathRelativeToRootDir("distributions/" + distro + "/test/spec-nightly.yaml")
	testSpecData, err := os.ReadFile(testSpecFile)
	if err != nil {
		panic(fmt.Errorf("failed to read test spec file: %w", err))
	}

	var testSpec TestSpec
	err = yaml.Unmarshal(testSpecData, &testSpec)
	if err != nil {
		panic(fmt.Errorf("failed to unmarshal test spec: %w", err))
	}

	return &testSpec
}
