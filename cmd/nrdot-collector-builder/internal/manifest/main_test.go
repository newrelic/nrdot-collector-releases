// Copyright 2025 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0
package manifest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestIsOtelCoreComponent(t *testing.T) {
	assert.True(t, isOtelCoreComponent("go.opentelemetry.io/collector/component v1.0.0"))
	assert.True(t, isOtelCoreComponent("go.opentelemetry.io/collector/component"))
	assert.False(t, isOtelCoreComponent("github.com/some/other/module"))
}

func TestIsOtelContribComponent(t *testing.T) {
	assert.True(t, isOtelContribComponent("github.com/open-telemetry/opentelemetry-collector-contrib/component v1.0.0"))
	assert.True(t, isOtelContribComponent("github.com/open-telemetry/opentelemetry-collector-contrib/component"))
	assert.False(t, isOtelContribComponent("github.com/some/other/module"))
}

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
