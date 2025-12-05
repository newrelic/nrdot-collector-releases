package test

import (
	"strings"
	envutil "test/e2e/util/env"
)

const (
	Wildcard                 = "%"
	testKeySeparator = "-"
)

func NewTestKeyPattern(envName string, deployId string, hostType string) string {
	distro := envutil.GetDistro()

	if envutil.IsFipsMode() {
		envName = envName + testKeySeparator + "fips"
	}

	hostId := Wildcard
	return strings.Join([]string{envName, deployId, distro, hostType, hostId}, testKeySeparator)
}
