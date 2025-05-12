// Copyright 2025 New Relic Corporation. All rights reserved.
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
