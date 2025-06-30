package test

import (
	"github.com/gruntwork-io/terratest/modules/random"
	"strings"
	envutil "test/e2e/util/env"
	"testing"
)

const (
	slowOnly    = "slowOnly"
	nightlyOnly = "nightlyOnly"
)

var onlyModes = []string{slowOnly, nightlyOnly}

func TagAsSlowTest(t *testing.T) {
	if isAnyModeOf(allOnlyModesExcept(slowOnly)) {
		t.Skip("Slow test skipped due to test mode: ", t.Name())
	}
}

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

func NewTestId() string {
	return strings.ToLower(random.UniqueId())
}
