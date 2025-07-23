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
	hostId := Wildcard
	return strings.Join([]string{envName, deployId, distro, hostType, hostId}, hostNameSegmentSeparator)
}
