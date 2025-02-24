KIND_CLUSTER_NAME ?=e2etest
K8S_CONTEXT_NAME=kind-${KIND_CLUSTER_NAME}
DISTRO ?=nrdot-collector-host
IMAGE_REPO ?=newrelic/${DISTRO}
THIS_MAKEFILE_DIR := $(realpath $(dir $(realpath $(lastword $(MAKEFILE_LIST)))))
ROOT_DIR := $(realpath $(THIS_MAKEFILE_DIR)/../..)


################
#### ASSERT ####
################
.PHONY: assert_require-binary
assert_require-binary:
	@{ \
		for bin in $(REQUIRED_BINS); do \
			if ! command -v "$$bin" >/dev/null 2>/dev/null; then \
				echo >&2 "command '$$bin' not found. Please install $$bin"; \
				exit 1; \
			else \
				echo "Found required dependency '$$bin'"; \
			fi; \
		done \
	}

.PHONY: assert_image-tag-present
assert_image-tag-present:
	@[ "${IMAGE_TAG}" ] || ( echo ">> env var IMAGE_TAG is not set"; exit 1 )

.PHONY: assert_cluster-exists
assert_cluster-exists: REQUIRED_BINS=kind go
assert_cluster-exists: assert_require-binary
	kind get clusters | grep -q ${KIND_CLUSTER_NAME} || ( echo ">> cluster ${KIND_CLUSTER_NAME} does not exist"; exit 1 )

################
###### CI ######
################
.PHONY: ci_load-image
ci_load-image: REQUIRED_BINS=kind
ci_load-image: assert_require-binary assert_cluster-exists assert_image-tag-present
	kind load docker-image ${IMAGE_REPO}:${IMAGE_TAG} --name ${KIND_CLUSTER_NAME}

.PHONY: ci_build-load-mocked-otlp-image
ci_build-load-mocked-otlp-image: REQUIRED_BINS=docker kind
ci_build-load-mocked-otlp-image: BUILD_PATH=${THIS_MAKEFILE_DIR}/mocked_otlp/http
ci_build-load-mocked-otlp-image: assert_require-binary assert_cluster-exists
	docker build -t mocked_otlp:latest -f ${BUILD_PATH}/Dockerfile ${BUILD_PATH}
	kind load docker-image mocked_otlp:latest --name ${KIND_CLUSTER_NAME}

.PHONY: ci_test-fast
ci_test-fast: TEST_MODE=fastOnly
ci_test-fast: ci_load-image ci_build-load-mocked-otlp-image ci_test

.PHONY: ci_test-slow
ci_test-slow: TEST_MODE=slowOnly
ci_test-slow: ci_load-image ci_test

.PHONY: ci_test-nightly
ci_test-nightly: TEST_MODE=nightlyOnly
ci_test-nightly: ci_test

.PHONY: ci_test
ci_test: REQUIRED_BINS=go
ci_test: assert_require-binary
	cd ${THIS_MAKEFILE_DIR} && \
	E2E_TEST__K8S_CONTEXT_NAME=${K8S_CONTEXT_NAME} \
	E2E_TEST__DISTRO_TO_TEST=${DISTRO} \
	E2E_TEST__IMAGE_TAG=${IMAGE_TAG} \
	E2E_TEST__IMAGE_REPO=${IMAGE_REPO} \
	E2E_TEST__NR_BACKEND_URL=${NR_BACKEND_URL} \
	E2E_TEST__NR_INGEST_KEY=${NR_INGEST_KEY} \
	E2E_TEST__NR_API_KEY=${NR_API_KEY} \
	E2E_TEST__NR_ACCOUNT_ID=${NR_ACCOUNT_ID} \
	E2E_TEST__NR_API_BASE_URL=${NR_API_BASE_URL} \
	E2E_TEST__TEST_MODE=${TEST_MODE} \
	go test ./... -count=1 -timeout=3m

################
#### LOCAL #####
################
MOST_RECENT_IMAGE_TAG := $(shell docker images -a --filter 'reference=newrelic/${DISTRO}' '--format={{.CreatedAt}} {{.Tag}}' | sort -r | grep 'arm64' | head -n1 | awk '{print $$NF}')

.PHONY: local_create-cluster-if-not-exists
local_create-cluster-if-not-exists:
	kind get clusters | grep -q ${KIND_CLUSTER_NAME} \
	|| kind create cluster --name ${KIND_CLUSTER_NAME}

.PHONY: local_build-image
local_build-image:
	cd $(ROOT_DIR) && goreleaser --snapshot --clean --skip=sign

.PHONY: local_test-fast
local_test-fast: TEST_MODE=fastOnly
local_test-fast: IMAGE_TAG=${MOST_RECENT_IMAGE_TAG}
local_test-fast: local_create-cluster-if-not-exists ci_test-fast

.PHONY: local_test-slow
local_test-slow: TEST_MODE=slowOnly
local_test-slow: IMAGE_TAG=${MOST_RECENT_IMAGE_TAG}
local_test-slow: local_create-cluster-if-not-exists ci_test-slow

.PHONY: local_helm-cleanup
local_helm-cleanup:
	kubectl config use-context ${K8S_CONTEXT_NAME} && \
	helm list -A --no-headers | awk 'NF {system("helm uninstall -n " $$2 " " $$1)}'
