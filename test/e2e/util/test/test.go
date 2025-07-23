package test

import (
	"strings"
	envutil "test/e2e/util/env"
	"testing"
)

const (
	nightlyOnly = "nightlyOnly"
)

var onlyModes = []string{nightlyOnly}

func TagAsNightlyTest(t *testing.T) {
	if isAnyModeOf(allOnlyModesExcept(nightlyOnly)) {
		t.Skip("Nightly test skipped due to test mode: ", t.Name())
	}
}

func isAnyModeOf(modes []string) bool {
	result := false
	for _, mode := range modes {
		result = result || strings.Contains(envutil.GetTestMode(), mode)
	}
	return result
}
func allOnlyModesExcept(filterOut string) []string {
	var result []string
	for _, onlyMode := range onlyModes {
		if onlyMode != filterOut {
			result = append(result, onlyMode)
		}
	}
	return result
}
