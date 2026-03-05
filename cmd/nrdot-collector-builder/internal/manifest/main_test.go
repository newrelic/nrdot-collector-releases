// Copyright New Relic, Inc. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package manifest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestFetchAllModuleVersions_Success(t *testing.T) {
	cfg := &Config{
		Verbose: true,
		Logger:  zap.NewNop(),
		Dir:     "./",
		Distribution: Distribution{
			Go: "go",
		},
	}

	modules := []string{"github.com/stretchr/testify"}
	versions, err := fetchAllModuleVersions(cfg, modules)
	assert.NoError(t, err)
	assert.Contains(t, versions, "github.com/stretchr/testify")
}

func TestFetchAllModuleVersions_Failure(t *testing.T) {
	cfg := &Config{
		Verbose: true,
		Logger:  zap.NewNop(),
		Dir:     "./",
		Distribution: Distribution{
			Go: "nonexistent-go",
		},
	}

	modules := []string{"github.com/stretchr/testify"}
	_, err := fetchAllModuleVersions(cfg, modules)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch module versions")
}

// newTestCfg returns a Config with the minimum set of modules required by
// SetVersions: at least one beta core module (v0.x), one stable core module
// (v1.x), and one contrib module. These are placed in Connectors so that
// individual tests can freely override Receivers, Processors, etc.
func newTestCfg() *Config {
	return &Config{
		Logger: zap.NewNop(),
		Connectors: []Module{
			{GoMod: "go.opentelemetry.io/collector/connector/forwardconnector v0.142.0"},
			{GoMod: "go.opentelemetry.io/collector/connector v1.48.0"},
			{GoMod: "github.com/open-telemetry/opentelemetry-collector-contrib/connector/routingconnector v0.142.0"},
		},
	}
}

func TestCopyAndUpdateConfigModules_BetaCore(t *testing.T) {
	cfg := newTestCfg()
	cfg.Receivers = []Module{
		{GoMod: "go.opentelemetry.io/collector/receiver/otlpreceiver v0.142.0"},
	}
	updates := map[string]VersionUpdate{
		"go.opentelemetry.io/collector": {BetaVersion: "v0.147.0", StableVersion: "v1.53.0"},
	}
	result, err := CopyAndUpdateConfigModules(cfg, updates)
	assert.NoError(t, err)
	assert.Equal(t, "go.opentelemetry.io/collector/receiver/otlpreceiver v0.147.0", result.Receivers[0].GoMod)
}

func TestCopyAndUpdateConfigModules_StableCore(t *testing.T) {
	cfg := newTestCfg()
	cfg.ConfmapProviders = []Module{
		{GoMod: "go.opentelemetry.io/collector/confmap/provider/envprovider v1.48.0"},
	}
	updates := map[string]VersionUpdate{
		"go.opentelemetry.io/collector": {BetaVersion: "v0.147.0", StableVersion: "v1.53.0"},
	}
	result, err := CopyAndUpdateConfigModules(cfg, updates)
	assert.NoError(t, err)
	assert.Equal(t, "go.opentelemetry.io/collector/confmap/provider/envprovider v1.53.0", result.ConfmapProviders[0].GoMod)
}

func TestCopyAndUpdateConfigModules_StableNotAppliedToBeta(t *testing.T) {
	cfg := newTestCfg()
	cfg.Receivers = []Module{
		{GoMod: "go.opentelemetry.io/collector/receiver/otlpreceiver v0.142.0"},
	}
	// Only stable version provided — beta module should not be changed
	updates := map[string]VersionUpdate{
		"go.opentelemetry.io/collector": {StableVersion: "v1.53.0"},
	}
	result, err := CopyAndUpdateConfigModules(cfg, updates)
	assert.NoError(t, err)
	assert.Equal(t, "go.opentelemetry.io/collector/receiver/otlpreceiver v0.142.0", result.Receivers[0].GoMod)
}

func TestCopyAndUpdateConfigModules_Contrib(t *testing.T) {
	cfg := newTestCfg()
	cfg.Receivers = []Module{
		{GoMod: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/filelogreceiver v0.142.0"},
	}
	updates := map[string]VersionUpdate{
		"github.com/open-telemetry/opentelemetry-collector-contrib": {BetaVersion: "v0.147.2"},
	}
	result, err := CopyAndUpdateConfigModules(cfg, updates)
	assert.NoError(t, err)
	assert.Equal(t, "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/filelogreceiver v0.147.2", result.Receivers[0].GoMod)
}

func TestCopyAndUpdateConfigModules_Nrdot(t *testing.T) {
	cfg := newTestCfg()
	cfg.Processors = []Module{
		{GoMod: "github.com/newrelic/nrdot-collector-components/processor/adaptivetelemetryprocessor v0.142.2"},
	}
	updates := map[string]VersionUpdate{
		"github.com/newrelic/nrdot-collector-components": {BetaVersion: "v0.147.0"},
	}
	result, err := CopyAndUpdateConfigModules(cfg, updates)
	assert.NoError(t, err)
	assert.Equal(t, "github.com/newrelic/nrdot-collector-components/processor/adaptivetelemetryprocessor v0.147.0", result.Processors[0].GoMod)
}

func TestCopyAndUpdateConfigModules_NoMatchUnchanged(t *testing.T) {
	cfg := newTestCfg()
	cfg.Receivers = []Module{
		{GoMod: "github.com/some/other/module v1.2.3"},
	}
	updates := map[string]VersionUpdate{
		"go.opentelemetry.io/collector": {BetaVersion: "v0.147.0", StableVersion: "v1.53.0"},
	}
	result, err := CopyAndUpdateConfigModules(cfg, updates)
	assert.NoError(t, err)
	assert.Equal(t, "github.com/some/other/module v1.2.3", result.Receivers[0].GoMod)
}
