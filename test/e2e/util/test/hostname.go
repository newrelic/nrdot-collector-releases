package test

import (
	"strings"
	envutil "test/e2e/util/env"
)

const (
	Wildcard                 = "%"
	hostNameSegmentSeparator = "-"
)

func NewNrQueryHostNamePattern(envName string, deployId string, hostType string) string {
	distro := envutil.GetDistro()
	fips := envutil.GetFipsMode()
	hostId := Wildcard
	return strings.Join([]string{envName, fips, deployId, distro, hostType, hostId}, hostNameSegmentSeparator)
}
