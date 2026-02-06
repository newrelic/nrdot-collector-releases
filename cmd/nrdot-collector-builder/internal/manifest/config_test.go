// Copyright New Relic, Inc. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package manifest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestConfig_Validate(t *testing.T) {
	cfg := &Config{
		Extensions: []Module{
			{GoMod: "module1"},
		},
		Receivers: []Module{
			{GoMod: "module2"},
		},
	}

	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestConfig_Validate_MissingGoMod(t *testing.T) {
	cfg := &Config{
		Extensions: []Module{
			{},
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing gomod specification for module")
}

func TestConfig_SetGoPath(t *testing.T) {
	cfg := &Config{
		Logger: zap.NewNop(),
		Distribution: Distribution{
			Go: "go",
		},
	}

	err := cfg.SetGoPath()
	assert.NoError(t, err)
	assert.NotEmpty(t, cfg.Distribution.Go)
}

func TestConfig_ParseModules(t *testing.T) {
	cfg := &Config{
		Extensions: []Module{
			{GoMod: "module1"},
		},
		Receivers: []Module{
			{GoMod: "module2"},
		},
	}

	err := cfg.ParseModules()
	assert.NoError(t, err)
}

func TestConfig_ParseModules_InvalidPath(t *testing.T) {
	cfg := &Config{
		Extensions: []Module{
			{GoMod: "module1", Path: "invalid/path"},
		},
	}

	err := cfg.ParseModules()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "filepath does not exist")
}

func TestConfig_IsOtelCoreComponent(t *testing.T) {
	assert.True(t, isOtelCoreComponent("go.opentelemetry.io/collector/component v1.0.0"))
	assert.True(t, isOtelCoreComponent("go.opentelemetry.io/collector/component"))
	assert.False(t, isOtelCoreComponent("github.com/some/other/module"))
}

func TestConfig_IsOtelContribComponent(t *testing.T) {
	assert.True(t, isOtelContribComponent("github.com/open-telemetry/opentelemetry-collector-contrib/component v1.0.0"))
	assert.True(t, isOtelContribComponent("github.com/open-telemetry/opentelemetry-collector-contrib/component"))
	assert.False(t, isOtelContribComponent("github.com/some/other/module"))
}

func TestConfig_IsNrdotComponent(t *testing.T) {
	assert.True(t, isNrdotComponent(Module{GoMod: "github.com/newrelic/nrdot-collector-components/component v1.0.0"}))
	assert.True(t, isNrdotComponent(Module{GoMod: "github.com/newrelic/nrdot-collector-components/component"}))
	assert.False(t, isNrdotComponent(Module{GoMod: "github.com/some/other/module"}))
}

func TestConfig_SetVersions(t *testing.T) {
	cfg := &Config{
		Extensions: []Module{
			{GoMod: "github.com/open-telemetry/opentelemetry-collector-contrib/component v0.1.0"},
			{GoMod: "github.com/newrelic/nrdot-collector-components/component v0.1.0"},
		},
		Receivers: []Module{
			{GoMod: "go.opentelemetry.io/collector v1.0.0"},
			{GoMod: "go.opentelemetry.io/collector/component v0.1.0"},
		},
	}

	err := cfg.SetVersions()
	assert.NoError(t, err)

	assert.Equal(t, "v1.0.0", cfg.Versions.StableCoreVersion)
	assert.Equal(t, "v0.1.0", cfg.Versions.BetaCoreVersion)
	assert.Equal(t, "v0.1.0", cfg.Versions.BetaContribVersion)
	assert.Equal(t, "v0.1.0", cfg.Versions.NrdotVersion)
}

func TestConfig_SetVersions_MissingCore(t *testing.T) {
	cfg := &Config{
		Extensions: []Module{
			{GoMod: "github.com/open-telemetry/opentelemetry-collector-contrib/component v0.1.0"},
		},
		Receivers: []Module{
			{GoMod: "go.opentelemetry.io/collector v1.0.0"},
		},
	}

	err := cfg.SetVersions()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing beta core version")

}

func TestIsCompatibleWithNrdotComponent(t *testing.T) {
	tests := []struct {
		name          string
		nrdotVersion  string
		betaVersion   string
		expectedMatch bool
	}{
		{
			name:          "nrdot version equal to beta version",
			nrdotVersion:  "v0.142.0",
			betaVersion:   "v0.142.0",
			expectedMatch: true,
		},
		{
			name:          "different minor versions - nrdot higher",
			nrdotVersion:  "v0.143.0",
			betaVersion:   "v0.142.5",
			expectedMatch: true,
		},
		{
			name:          "different minor versions - beta higher",
			nrdotVersion:  "v0.142.0",
			betaVersion:   "v0.142.1",
			expectedMatch: false,
		},
		{
			name:          "different patch versions - nrdot higher",
			nrdotVersion:  "v0.142.5",
			betaVersion:   "v0.142.3",
			expectedMatch: true,
		},
		{
			name:          "different patch versions - beta higher",
			nrdotVersion:  "v0.142.3",
			betaVersion:   "v0.142.5",
			expectedMatch: false,
		},
		{
			name:          "edge case - large version numbers",
			nrdotVersion:  "v0.999.999",
			betaVersion:   "v0.999.998",
			expectedMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isCompatibleWithNrdotComponent(tt.nrdotVersion, tt.betaVersion)
			assert.Equal(t, tt.expectedMatch, result,
				"isCompatibleWithNrdotComponent(%s, %s) = %v, want %v",
				tt.nrdotVersion, tt.betaVersion, result, tt.expectedMatch)
		})
	}
}
