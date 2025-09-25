KIND_CLUSTER_NAME ?=e2etest
K8S_CONTEXT_NAME ?=kind-${KIND_CLUSTER_NAME}
DISTRO ?=nrdot-collector-host
FIPS ?=false
THIS_MAKEFILE_DIR := $(realpath $(dir $(realpath $(lastword $(MAKEFILE_LIST)))))
ROOT_DIR := $(realpath $(THIS_MAKEFILE_DIR)/../..)

################
###### CI ######
################
.PHONY: ci_test-nightly
ci_test-nightly: TEST_MODE=nightlyOnly
ci_test-nightly: ci_test

.PHONY: ci_test
ci_test:
	cd ${THIS_MAKEFILE_DIR} && \
	E2E_TEST__K8S_CONTEXT_NAME=${K8S_CONTEXT_NAME} \
	E2E_TEST__DISTRO_TO_TEST=${DISTRO} \
	E2E_TEST__NR_API_KEY=${NR_API_KEY} \
	E2E_TEST__NR_ACCOUNT_ID=${NR_ACCOUNT_ID} \
	E2E_TEST__NR_API_BASE_URL=${NR_API_BASE_URL} \
	E2E_TEST__TEST_MODE=${TEST_MODE} \
	E2E_TEST__FIPS_MODE=${FIPS}
	go test ./... -count=1 -timeout=10m -parallel=50