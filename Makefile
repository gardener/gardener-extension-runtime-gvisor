# Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

EXTENSION_PREFIX            := gardener-extension
NAME                        := runtime-gvisor
NAME_INSTALLATION           := runtime-gvisor-installation
REGISTRY                    := eu.gcr.io/gardener-project/gardener
IMAGE_PREFIX                := $(REGISTRY)/extensions
REPO_ROOT                   := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
HACK_DIR                    := $(REPO_ROOT)/hack
VERSION                     := $(shell cat "$(REPO_ROOT)/VERSION")
LD_FLAGS                    := "-w -X github.com/gardener/$(EXTENSION_PREFIX)-$(NAME)/pkg/version.Version=$(IMAGE_TAG)"
VERIFY                      := true
LEADER_ELECTION             := false
IGNORE_OPERATION_ANNOTATION := true

### GVisor version: https://github.com/google/gvisor/releases
RUNSC_VERSION				 	:= 20200219.0

### GVisor containerd shim version: https://github.com/google/gvisor-containerd-shim/releases
CONTAINERD_RUNSC_SHIM_VERSION 	:= v0.0.4

### Build commands

.PHONY: format
format:
	@$(REPO_ROOT)/vendor/github.com/gardener/gardener-extensions/hack/format.sh ./cmd ./pkg

.PHONY: clean
clean:
	@$(REPO_ROOT)/vendor/github.com/gardener/gardener-extensions/hack/clean.sh ./pkg/...

.PHONY: generate
generate:
	@$(REPO_ROOT)/vendor/github.com/gardener/gardener-extensions/hack/generate.sh ./...

.PHONY: check
check:
	@$(REPO_ROOT)/vendor/github.com/gardener/gardener-extensions/hack/check.sh ./...

.PHONY: test
test:
	@$(REPO_ROOT)/vendor/github.com/gardener/gardener-extensions/hack/test.sh -r ./...

.PHONY: verify
verify: check generate test format

.PHONY: install
install:
	@LD_FLAGS="-w -X github.com/gardener/$(EXTENSION_PREFIX)-$(NAME)/pkg/version.Version=$(VERSION) \
			   -w -X github.com/gardener/$(EXTENSION_PREFIX)-$(NAME)/pkg/version.gitVersion=$(VERSION) \
			   -w -X github.com/gardener/$(EXTENSION_PREFIX)-$(NAME)/pkg/version.gitTreeState=$(shell sh -c '[ -z git status --porcelain 2>/dev/null ] && echo clean || echo dirty') \
			   -w -X github.com/gardener/$(EXTENSION_PREFIX)-$(NAME)/pkg/version.gitCommit=$(shell sh -c 'git rev-parse --verify HEAD') \
			   -w -X github.com/gardener/$(EXTENSION_PREFIX)-$(NAME)/pkg/version.buildDate=$(shell sh -c 'date --iso-8601=seconds')" \
	$(REPO_ROOT)/vendor/github.com/gardener/gardener-extensions/hack/install.sh ./...

.PHONY: install-requirements
install-requirements:
	@go install -mod=vendor $(REPO_ROOT)/vendor/github.com/ahmetb/gen-crd-api-reference-docs
	@go install -mod=vendor $(REPO_ROOT)/vendor/github.com/gobuffalo/packr/v2/packr2
	@go install -mod=vendor $(REPO_ROOT)/vendor/github.com/onsi/ginkgo/ginkgo
	@$(REPO_ROOT)/vendor/github.com/gardener/gardener-extensions/hack/install-requirements.sh

.PHONY: install-binaries
install-binaries:
	@bash $(HACK_DIR)/install-binaries.sh $(RUNSC_VERSION) $(CONTAINERD_RUNSC_SHIM_VERSION)

.PHONY: all
ifeq ($(VERIFY),true)
all: verify generate install
else
all: generate install
endif

### Docker commands

.PHONY: docker-login
docker-login:
	@gcloud auth activate-service-account --key-file .kube-secrets/gcr/gcr-readwrite.json

# eu.gcr.io/gardener-project/gardener/extensions/runtime-gvisor:0.0.0-dev
.PHONY: docker-images
docker-images:
	@docker build -t $(IMAGE_PREFIX)/$(NAME):$(VERSION) -t $(IMAGE_PREFIX)/$(NAME):latest -f Dockerfile -m 6g --target $(EXTENSION_PREFIX)-$(NAME) .
	@docker build -t $(IMAGE_PREFIX)/$(NAME_INSTALLATION):$(VERSION) -t $(IMAGE_PREFIX)/$(NAME_INSTALLATION):latest -f Dockerfile -m 200m --target $(EXTENSION_PREFIX)-$(NAME_INSTALLATION) .

### Debug / Development commands

.PHONY: revendor
revendor:
	@GO111MODULE=on go mod vendor
	@GO111MODULE=on go mod tidy
	@chmod +x $(REPO_ROOT)/vendor/github.com/gardener/gardener-extensions/hack/*
	@chmod +x $(REPO_ROOT)/vendor/github.com/gardener/gardener-extensions/hack/.ci/*

.PHONY: start
start:
	@LEADER_ELECTION_NAMESPACE=garden GO111MODULE=on go run \
		-mod=vendor \
		-ldflags $(LD_FLAGS) \
		./cmd/$(EXTENSION_PREFIX)-$(NAME) \
		--leader-election=$(LEADER_ELECTION)
