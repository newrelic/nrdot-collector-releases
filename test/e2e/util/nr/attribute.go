package nr

import (
	envutil "test/e2e/util/env"
	testutil "test/e2e/util/test"
)

func GetHostNamePrefix(testId string) string {
	var environmentName = envutil.GetEnvironmentName()
	hostNamePrefix := testutil.NewHostNamePrefix(environmentName, testId, "k8s_node")
	return hostNamePrefix
}

func GetHostNamePattern(testId string) string {
	var environmentName = envutil.GetEnvironmentName()
	hostNamePattern := testutil.NewNrQueryHostNamePattern(environmentName, testId, "k8s_node")
	return hostNamePattern
}
