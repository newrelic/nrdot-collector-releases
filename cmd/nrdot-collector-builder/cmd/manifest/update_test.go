// Copyright New Relic, Inc. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package manifest

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockConfig struct {
	mock.Mock
}

func (m *MockConfig) Validate() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockConfig) SetGoPath() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockConfig) ParseModules() error {
	args := m.Called()
	return args.Error(0)
}

func TestUpdateCmd_RunE(t *testing.T) {

	// Load the test-config.yaml file
	testConfigPath := "testdata/test-config.yaml"
	yamlData, err := os.ReadFile(testConfigPath)
	assert.NoError(t, err)

	// Create a temporary file to simulate writing to a file
	tempFile, err := os.CreateTemp("", "test-config-*.yaml")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name()) // Clean up the file after the test

	// Write the loaded YAML data to the temporary file
	_, err = tempFile.Write(yamlData)
	assert.NoError(t, err)
	tempFile.Close() // Close the file to ensure the changes are flushed

	cmd := &cobra.Command{}
	cmd.Flags().String("config", tempFile.Name(), "")

	err = UpdateCmd.RunE(cmd, []string{})
	assert.NoError(t, err)

	updatedYamlData, err := os.ReadFile(tempFile.Name())
	assert.NoError(t, err)

	assert.NotEqual(t, yamlData, updatedYamlData)
}

func TestUpdateCmd_RunE_InvalidConfig(t *testing.T) {
	// Load the test-config.yaml file
	testConfigPath := "testdata/test-config-invalid.yaml"
	yamlData, err := os.ReadFile(testConfigPath)
	assert.NoError(t, err)

	// Create a temporary file to simulate writing to a file
	tempFile, err := os.CreateTemp("", "test-config-*.yaml")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name()) // Clean up the file after the test

	// Write the loaded YAML data to the temporary file
	_, err = tempFile.Write(yamlData)
	assert.NoError(t, err)
	tempFile.Close() // Close the file to ensure the changes are flushed

	cmd := &cobra.Command{}
	cmd.Flags().String("config", tempFile.Name(), "")

	err = UpdateCmd.RunE(cmd, []string{})
	assert.Error(t, err)
}

func TestUpdateCmd_RunE_NrdotComponents(t *testing.T) {
	// Load the test-config-nrdot.yaml file
	testConfigPath := "testdata/test-config-nrdot.yaml"
	yamlData, err := os.ReadFile(testConfigPath)
	assert.NoError(t, err)

	// Create a temporary file to simulate writing to a file
	tempFile, err := os.CreateTemp("", "test-config-*.yaml")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name()) // Clean up the file after the test

	// Write the loaded YAML data to the temporary file
	_, err = tempFile.Write(yamlData)
	assert.NoError(t, err)
	tempFile.Close() // Close the file to ensure the changes are flushed

	targetStable := "v1.61.0"
	targetBeta := "v0.155.0"

	cmd := &cobra.Command{}
	cmd.Flags().String("config", tempFile.Name(), "")

	// Version overrides are read from the root command's persistent flags.
	cmd.PersistentFlags().String("nrdot-version", targetBeta, "")
	cmd.PersistentFlags().String("nr-fork-contrib-version", targetBeta, "")
	cmd.PersistentFlags().String("core-stable", targetStable, "")
	cmd.PersistentFlags().String("core-beta", targetBeta, "")
	cmd.PersistentFlags().String("contrib-beta", targetBeta, "")

	err = UpdateCmd.RunE(cmd, []string{})
	assert.NoError(t, err)

	updatedYamlData, err := os.ReadFile(tempFile.Name())
	assert.NoError(t, err)
	updated := string(updatedYamlData)

	betaModules := []string{
		"github.com/newrelic/nrdot-collector-components/processor/adaptivetelemetryprocessor",
		"github.com/newrelic-forks/opentelemetry-collector-contrib/receiver/nrsqlserverreceiver",
		"go.opentelemetry.io/collector/receiver/nopreceiver",
		"go.opentelemetry.io/collector/receiver/nopexporter",
	}
	for _, module := range betaModules {
		assert.Contains(t, updated, module+" "+targetBeta)
	}
}
