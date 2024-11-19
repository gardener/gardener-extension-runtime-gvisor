# SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

ENSURE_GARDENER_MOD         := $(shell go get github.com/gardener/gardener@$$(go list -m -f "{{.Version}}" github.com/gardener/gardener))
GARDENER_HACK_DIR           := $(shell go list -m -f "{{.Dir}}" github.com/gardener/gardener)/hack
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
LD_FLAGS_GENERATOR          := $(GARDENER_HACK_DIR)/get-build-ld-flags.sh
IGNORE_OPERATION_ANNOTATION := true
PLATFORM                    := linux/amd64

ifneq ($(strip $(shell git status --porcelain 2>/dev/null)),)
	EFFECTIVE_VERSION := $(EFFECTIVE_VERSION)-dirty
endif

LD_FLAGS := $(shell chmod +x $(LD_FLAGS_GENERATOR) && EFFECTIVE_VERSION=$(EFFECTIVE_VERSION) $(LD_FLAGS_GENERATOR) k8s.io/component-base $(REPO_ROOT)/VERSION $(EXTENSION_PREFIX))

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

TOOLS_DIR := $(HACK_DIR)/tools
include $(GARDENER_HACK_DIR)/tools.mk

#################################################################
# Rules related to binary build, Docker image build and release #
#################################################################

.PHONY: install
install:
	@LD_FLAGS="$(LD_FLAGS)" \
	bash $(GARDENER_HACK_DIR)/install.sh ./...

.PHONY: install-binaries
install-binaries:
	$(HACK_DIR)/install-binaries.sh $(GVISOR_VERSION)

.PHONY: docker-login
docker-login:
	@gcloud auth activate-service-account --key-file .kube-secrets/gcr/gcr-readwrite.json

.PHONY: docker-image-runtime
docker-image-runtime:
	@docker buildx build --platform=$(PLATFORM) \
		--build-arg EFFECTIVE_VERSION=$(EFFECTIVE_VERSION) \
		-t $(IMAGE_PREFIX)/$(NAME):$(EFFECTIVE_VERSION) \
		-t $(IMAGE_PREFIX)/$(NAME):latest \
		-f Dockerfile \
		-m 6g \
		--target $(EXTENSION_PREFIX)-$(NAME) \
		.

.PHONY: docker-image-installation
docker-image-installation:
	@docker buildx build --platform=$(PLATFORM) \
		-t $(IMAGE_PREFIX)/$(NAME_INSTALLATION):$(EFFECTIVE_VERSION) \
		-t $(IMAGE_PREFIX)/$(NAME_INSTALLATION):latest \
		-f Dockerfile \
		-m 6g \
		--target $(EXTENSION_PREFIX)-$(NAME_INSTALLATION) \
		.

.PHONY: docker-images
docker-images: docker-image-installation docker-image-runtime

#####################################################################
# Rules for verification, formatting, linting, testing and cleaning #
#####################################################################

.PHONY: tidy
tidy:
	@GO111MODULE=on go mod tidy
	@mkdir -p $(REPO_ROOT)/.ci/hack && cp $(GARDENER_HACK_DIR)/.ci/* $(REPO_ROOT)/.ci/hack/ && chmod +xw $(REPO_ROOT)/.ci/hack/*
	@cp $(GARDENER_HACK_DIR)/cherry-pick-pull.sh $(HACK_DIR)/cherry-pick-pull.sh && chmod +xw $(HACK_DIR)/cherry-pick-pull.sh

.PHONY: clean
clean:
	@$(shell find ./example -type f -name "controller-registration.yaml" -exec rm '{}' \;)
	@bash $(GARDENER_HACK_DIR)/clean.sh ./cmd/... ./pkg/...

.PHONY: check-generate
check-generate:
	@bash $(GARDENER_HACK_DIR)/check-generate.sh $(REPO_ROOT)

.PHONY: check
check: $(GOIMPORTS) $(GOLANGCI_LINT) $(HELM)
	@REPO_ROOT=$(REPO_ROOT) bash $(GARDENER_HACK_DIR)/check.sh --golangci-lint-config=./.golangci.yaml ./cmd/... ./pkg/...
	@REPO_ROOT=$(REPO_ROOT) bash $(GARDENER_HACK_DIR)/check-charts.sh ./charts

.PHONY: generate
generate: $(VGOPATH) $(CONTROLLER_GEN) $(GEN_CRD_API_REFERENCE_DOCS) $(HELM) $(MOCKGEN) $(YQ)
	@REPO_ROOT=$(REPO_ROOT) VGOPATH=$(VGOPATH) GARDENER_HACK_DIR=$(GARDENER_HACK_DIR) bash $(GARDENER_HACK_DIR)/generate-sequential.sh ./charts/... ./cmd/... ./pkg/...
	$(MAKE) format

.PHONY: format
format: $(GOIMPORTS) $(GOIMPORTSREVISER)
	@bash $(GARDENER_HACK_DIR)/format.sh ./cmd ./pkg

.PHONY: sast
sast: $(GOSEC)
	@./hack/sast.sh

.PHONY: sast-report
sast-report: $(GOSEC)
	@./hack/sast.sh --gosec-report true

.PHONY: test
test:
	@bash $(GARDENER_HACK_DIR)/test.sh ./cmd/... ./pkg/...

.PHONY: test-cov
test-cov:
	@bash $(GARDENER_HACK_DIR)/test-cover.sh ./cmd/... ./pkg/...

.PHONY: test-cov-clean
test-cov-clean:
	@bash $(GARDENER_HACK_DIR)/test-cover-clean.sh

.PHONY: verify
verify: check format sast test

.PHONY: verify-extended
verify-extended: check-generate check format sast-report test-cov test-cov-clean

.PHONY: start
start:
	@./hack/start-gvisor-extension.sh -l "$(LD_FLAGS)" -d "$(CMD_DIRECTORY)" -i "$(IGNORE_OPERATION_ANNOTATION)" -r "$(REPO_ROOT)"
