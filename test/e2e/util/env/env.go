package env

import (
	"fmt"
	"log"
	"os"
	osuser "os/user"
	"strconv"
)

const (
	TestMode       = "E2E_TEST__TEST_MODE"
	K8sContextName = "E2E_TEST__K8S_CONTEXT_NAME"
	Distro         = "E2E_TEST__DISTRO_TO_TEST"
	ImageTag       = "E2E_TEST__IMAGE_TAG"
	ImageRepo      = "E2E_TEST__IMAGE_REPO"
	NrBackendUrl   = "E2E_TEST__NR_BACKEND_URL"
	NrIngestKey    = "E2E_TEST__NR_INGEST_KEY"
	NrApiKey       = "E2E_TEST__NR_API_KEY"
	NrAccountId    = "E2E_TEST__NR_ACCOUNT_ID"
	NrApiBaseUrl   = "E2E_TEST__NR_API_BASE_URL"
	CI             = "CI" // auto-populated by github action
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

func GetImageTag() string {
	return getEnvVar(ImageTag)
}

func GetImageRepo() string {
	return getEnvVar(ImageRepo)
}

func GetNrBackendUrl() string {
	return getEnvVar(NrBackendUrl)
}

func GetNrIngestKey() string {
	return getEnvVar(NrIngestKey)
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

func IsContinuousIntegration() bool {
	return os.Getenv(CI) == "true"
}

func GetEnvironmentName() string {
	var environmentName string
	if IsContinuousIntegration() {
		environmentName = "ci"
	} else {
		user, err := osuser.Current()
		if err != nil {
			log.Panicf("Couldn't determine current user: %v", err)
		}
		environmentName = fmt.Sprintf("local_%s", user.Username)
	}
	return environmentName
}
