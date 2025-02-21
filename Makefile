GO ?= go
GORELEASER ?= goreleaser

# SRC_ROOT is the top of the source tree.
SRC_ROOT := $(shell git rev-parse --show-toplevel)
OTELCOL_BUILDER_VERSION ?= 0.119.0
OTELCOL_BUILDER_DIR ?= ${HOME}/bin
OTELCOL_BUILDER ?= ${OTELCOL_BUILDER_DIR}/ocb

GOCMD?= go
TOOLS_MOD_DIR   := $(SRC_ROOT)/internal/tools
TOOLS_BIN_DIR   := $(SRC_ROOT)/.tools
TOOLS_MOD_REGEX := "\s+_\s+\".*\""
TOOLS_PKG_NAMES := $(shell grep -E $(TOOLS_MOD_REGEX) < $(TOOLS_MOD_DIR)/tools.go | tr -d " _\"" | grep -vE '/v[0-9]+$$')
TOOLS_BIN_NAMES := $(addprefix $(TOOLS_BIN_DIR)/, $(notdir $(shell echo $(TOOLS_PKG_NAMES))))
GO_LICENCE_DETECTOR        := $(TOOLS_BIN_DIR)/go-licence-detector
GO_LICENCE_DETECTOR_CONFIG   := $(SRC_ROOT)/internal/assets/license/rules.json

DISTRIBUTIONS ?= "nrdot-collector-host,nrdot-collector-k8s"

ci: check build version-check licenses-check
check: ensure-goreleaser-up-to-date

build: go ocb
	@./scripts/build.sh -d "${DISTRIBUTIONS}" -b ${OTELCOL_BUILDER}

generate: generate-sources generate-goreleaser

generate-goreleaser: go
	@./scripts/generate-goreleaser.sh -d "${DISTRIBUTIONS}" -g ${GO}

generate-sources: go ocb
	@./scripts/build.sh -d "${DISTRIBUTIONS}" -s true -b ${OTELCOL_BUILDER}

goreleaser-verify: goreleaser
	@${GORELEASER} release --snapshot --clean

ensure-goreleaser-up-to-date: generate-goreleaser
	@git diff -s --exit-code distributions/*/.goreleaser.yaml || (echo "Check failed: The goreleaser templates have changed but the .goreleaser.yamls haven't. Run 'make generate-goreleaser' and update your PR." && exit 1)

validate-components:
	@./scripts/validate-components.sh

.PHONY: ocb
ocb:
ifeq (, $(shell command -v ocb 2>/dev/null))
	@{ \
	[ ! -x '$(OTELCOL_BUILDER)' ] || exit 0; \
	set -e ;\
	os=$$(uname | tr A-Z a-z) ;\
	machine=$$(uname -m) ;\
	[ "$${machine}" != x86 ] || machine=386 ;\
	[ "$${machine}" != x86_64 ] || machine=amd64 ;\
	echo "Installing ocb ($${os}/$${machine}) at $(OTELCOL_BUILDER_DIR)";\
	mkdir -p $(OTELCOL_BUILDER_DIR) ;\
	CGO_ENABLED=0 go install -trimpath -ldflags="-s -w" go.opentelemetry.io/collector/cmd/builder@v$(OTELCOL_BUILDER_VERSION) ;\
	mv $$(go env GOPATH)/bin/builder $(OTELCOL_BUILDER) ;\
	}
else
OTELCOL_BUILDER=$(shell command -v ocb)
endif

.PHONY: go
go:
	@{ \
		if ! command -v '$(GO)' >/dev/null 2>/dev/null; then \
			echo >&2 '$(GO) command not found. Please install golang. https://go.dev/doc/install'; \
			exit 1; \
		fi \
	}

.PHONY: goreleaser
goreleaser:
	@{ \
		if ! command -v '$(GORELEASER)' >/dev/null 2>/dev/null; then \
			echo >&2 '$(GORELEASER) command not found. Please install goreleaser. https://goreleaser.com/install/'; \
			exit 1; \
		fi \
	}

VERSION := $(shell ./scripts/get-version.sh)

.PHONY: version-check
version-check:
	@echo $(VERSION)

REMOTE?=git@github.com:mailo-nr/opentelemetry-collector-releases.git
.PHONY: push-release-tag
push-release-tag:
	@[ "${VERSION}" ] || ( echo ">> VERSION is not set"; exit 1 )
	@echo "Adding tag ${VERSION}"
	@git tag -a ${VERSION} -s -m "Version ${VERSION}"
	@read -p "Are you sure you want to push the tag ${VERSION} to ${REMOTE}? (y/N) " confirm && [ $${confirm} = y ]
	@echo "Pushing tag ${VERSION}"
	@git push ${REMOTE} ${VERSION}


.PHONY: delete-tag
delete-tag:
	@[ "${VERSION}" ] || ( echo ">> VERSION is not set"; exit 1 )
	@echo "Deleting tag ${VERSION}"
	@git tag -d ${VERSION}
	@echo "Deleting remote tag ${VERSION}"
	@read -p "Are you sure you want to push the tag ${VERSION} to ${REMOTE}? (y/N) " confirm && [ $${confirm} = y ]
	@git push ${REMOTE} :refs/tags/${VERSION}

.PHONY: install-tools
install-tools: $(TOOLS_BIN_NAMES)

$(TOOLS_BIN_DIR):
	mkdir -p $@

$(TOOLS_BIN_NAMES): $(TOOLS_BIN_DIR) $(TOOLS_MOD_DIR)/go.mod
	cd $(TOOLS_MOD_DIR) && $(GOCMD) build -o $@ -trimpath $(filter %/$(notdir $@),$(TOOLS_PKG_NAMES))

FILENAME?=$(shell git branch --show-current)
NOTICE_OUTPUT?=THIRD_PARTY_NOTICES.md

.PHONY: licenses
licenses: go generate-sources $(GO_LICENCE_DETECTOR)
	@./scripts/licenses.sh -d "${DISTRIBUTIONS}" -b ${GO_LICENCE_DETECTOR} -n ${NOTICE_OUTPUT} -g ${GO}

.PHONY: licenses-check
licenses-check: licenses
	@git diff --name-only | grep -q $(NOTICE_OUTPUT) \
		&& { \
			echo "Third party notices out of date, please run \"make licenses\" and commit the changes in this PR.";\
			echo "Diff of $(NOTICE_OUTPUT):";\
			git --no-pager diff HEAD -- */$(NOTICE_OUTPUT);\
			exit 1;\
		} \
		|| exit 0
