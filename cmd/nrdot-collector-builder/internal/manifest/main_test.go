// Copyright 2025 New Relic Corporation. All rights reserved.
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
