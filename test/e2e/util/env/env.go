package env

import (
	"fmt"
	"os"
	"strconv"
)

const (
	TestMode       = "E2E_TEST__TEST_MODE"
	K8sContextName = "E2E_TEST__K8S_CONTEXT_NAME"
	Distro         = "E2E_TEST__DISTRO_TO_TEST"
	NrApiKey       = "E2E_TEST__NR_API_KEY"
	NrAccountId    = "E2E_TEST__NR_ACCOUNT_ID"
	NrApiBaseUrl   = "E2E_TEST__NR_API_BASE_URL"
)

func getEnvVar(envVar string) string {
	value := os.Getenv(envVar)
	if value == "" {
		panic(fmt.Sprintf("%s not set", envVar))
	}
	return value
}

func GetTestMode() string {
	return os.Getenv(TestMode)
}

func GetK8sContextName() string {
	return getEnvVar(K8sContextName)
}

func GetDistro() string {
	return getEnvVar(Distro)
}

func GetNrApiKey() string {
	return getEnvVar(NrApiKey)
}

func GetNrAccountId() int {
	accountIdStr := getEnvVar(NrAccountId)
	accountId, err := strconv.Atoi(accountIdStr)
	if err != nil {
		panic(fmt.Sprintf("Invalid accountId: %s", accountIdStr))
	}
	return accountId
}

func GetNrApiBaseUrl() string {
	return getEnvVar(NrApiBaseUrl)
}
