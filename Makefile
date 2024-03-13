# SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

EXTENSION_PREFIX            := gardener-extension
NAME                        := runtime-gvisor
NAME_INSTALLATION           := runtime-gvisor-installation
CMD_DIRECTORY		        := ./cmd/$(EXTENSION_PREFIX)-$(NAME)
REGISTRY                    := europe-docker.pkg.dev/gardener-project/public/gardener
IMAGE_PREFIX                := $(REGISTRY)/extensions
REPO_ROOT                   := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
HACK_DIR                    := $(REPO_ROOT)/hack
VERSION                     := $(shell cat "$(REPO_ROOT)/VERSION")
EFFECTIVE_VERSION           := $(VERSION)-$(shell git rev-parse HEAD)
LD_FLAGS_GENERATOR          := $(REPO_ROOT)/vendor/github.com/gardener/gardener/hack/get-build-ld-flags.sh
IGNORE_OPERATION_ANNOTATION := true

ifneq ($(strip $(shell git status --porcelain 2>/dev/null)),)
	EFFECTIVE_VERSION := $(EFFECTIVE_VERSION)-dirty
endif

LD_FLAGS := $(shell EFFECTIVE_VERSION=$(EFFECTIVE_VERSION) $(LD_FLAGS_GENERATOR) k8s.io/component-base $(REPO_ROOT)/VERSION $(EXTENSION_PREFIX))

### GVisor versions
# - https://github.com/google/gvisor/releases (not all Github tags are available in the registry)
# - https://gvisor.dev/docs/user_guide/install/
# To update the runsc + shim binary:
#  1) Download latest: https://storage.googleapis.com/gvisor/releases/release/latest/x86_64/runsc
#  2) Execute runsc --version to find version
#  3) Check that specific specific release can be downloaded: https://storage.googleapis.com/gvisor/releases/release/20230102.0/x86_64/runsc
#  4) Update version in GVISOR_VERSION file
GVISOR_VERSION := $(shell cat GVISOR_VERSION)

#########################################
# Tools                                 #
#########################################

TOOLS_DIR := hack/tools
include vendor/github.com/gardener/gardener/hack/tools.mk

#################################################################
# Rules related to binary build, Docker image build and release #
#################################################################

.PHONY: install
install:
	@LD_FLAGS="$(LD_FLAGS)" \
	$(REPO_ROOT)/vendor/github.com/gardener/gardener/hack/install.sh ./...

.PHONY: install-binaries
install-binaries:
	$(HACK_DIR)/install-binaries.sh $(GVISOR_VERSION)

.PHONY: docker-login
docker-login:
	@gcloud auth activate-service-account --key-file .kube-secrets/gcr/gcr-readwrite.json

.PHONY: docker-images
docker-images:
	@docker build \
		--build-arg EFFECTIVE_VERSION=$(EFFECTIVE_VERSION) \
		-t $(IMAGE_PREFIX)/$(NAME):$(EFFECTIVE_VERSION) \
		-t $(IMAGE_PREFIX)/$(NAME):latest \
		-f Dockerfile \
		-m 6g \
		--target $(EXTENSION_PREFIX)-$(NAME) \
		.
	@docker build \
		-t $(IMAGE_PREFIX)/$(NAME_INSTALLATION):$(EFFECTIVE_VERSION) \
		-t $(IMAGE_PREFIX)/$(NAME_INSTALLATION):latest \
		-f Dockerfile \
		-m 6g \
		--target $(EXTENSION_PREFIX)-$(NAME_INSTALLATION) \
		.

#####################################################################
# Rules for verification, formatting, linting, testing and cleaning #
#####################################################################

.PHONY: revendor
revendor:
	@GO111MODULE=on go mod tidy
	@GO111MODULE=on go mod vendor
	@chmod +x $(REPO_ROOT)/vendor/github.com/gardener/gardener/hack/*
	@chmod +x $(REPO_ROOT)/vendor/github.com/gardener/gardener/hack/.ci/*

.PHONY: clean
clean:
	@$(shell find ./example -type f -name "controller-registration.yaml" -exec rm '{}' \;)
	@$(REPO_ROOT)/vendor/github.com/gardener/gardener/hack/clean.sh ./cmd/... ./pkg/...

.PHONY: check-generate
check-generate:
	@$(REPO_ROOT)/vendor/github.com/gardener/gardener/hack/check-generate.sh $(REPO_ROOT)

.PHONY: check
check: $(GOIMPORTS) $(GOLANGCI_LINT) $(HELM)
	@$(REPO_ROOT)/vendor/github.com/gardener/gardener/hack/check.sh --golangci-lint-config=./.golangci.yaml ./cmd/... ./pkg/...
	@$(REPO_ROOT)/vendor/github.com/gardener/gardener/hack/check-charts.sh ./charts

.PHONY: generate
generate: $(CONTROLLER_GEN) $(GEN_CRD_API_REFERENCE_DOCS) $(HELM) $(MOCKGEN) $(YQ)
	@$(REPO_ROOT)/vendor/github.com/gardener/gardener/hack/generate-sequential.sh ./charts/... ./cmd/... ./pkg/...
	$(MAKE) format

.PHONY: format
format: $(GOIMPORTS) $(GOIMPORTSREVISER)
	@$(REPO_ROOT)/vendor/github.com/gardener/gardener/hack/format.sh ./cmd ./pkg

.PHONY: test
test:
	@$(REPO_ROOT)/vendor/github.com/gardener/gardener/hack/test.sh ./cmd/... ./pkg/...

.PHONY: test-cov
test-cov:
	@$(REPO_ROOT)/vendor/github.com/gardener/gardener/hack/test-cover.sh ./cmd/... ./pkg/...

.PHONY: test-cov-clean
test-cov-clean:
	@$(REPO_ROOT)/vendor/github.com/gardener/gardener/hack/test-cover-clean.sh

.PHONY: verify
verify: check format test

.PHONY: verify-extended
verify-extended: check-generate check format test-cov test-cov-clean

.PHONY: start
start:
	@./hack/start-gvisor-extension.sh -l "$(LD_FLAGS)" -d "$(CMD_DIRECTORY)" -i "$(IGNORE_OPERATION_ANNOTATION)" -r "$(REPO_ROOT)"
