# Copyright 2023 The K8sGPT Authors.
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#     http://www.apache.org/licenses/LICENSE-2.0
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# ==============================================================================
# define the default goal
#
ROOT_PACKAGE=github.com/k8sgpt-ai/k8sgpt

SHELL := /bin/bash
DIRS=$(shell ls)
GO=go
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

.DEFAULT_GOAL := help

# ==============================================================================
# Includes
include scripts/make-rules/common.mk # make sure include common.mk at the first include line
include scripts/make-rules/golang.mk
include scripts/make-rules/tools.mk
include scripts/make-rules/copyright.mk
#include scripts/make-rules/gen.mk
#include scripts/make-rules/release.mk

# ==============================================================================
# Usage
define USAGE_OPTIONS

Options:

  DEBUG            Whether or not to generate debug symbols. Default is 0.

  BINS             Binaries to build. Default is all binaries under cmd.
                   This option is available when using: make {build}(.multiarch)
                   Example: make build BINS="k8sgpt"

  PLATFORMS        Platform to build for. Default is linux_arm64 and linux_amd64.
                   This option is available when using: make {build}.multiarch
                   Example: make build.multiarch PLATFORMS="linux_arm64 linux_amd64"

  V                Set to 1 enable verbose build. Default is 0.
endef
export USAGE_OPTIONS

# ==============================================================================
# Targets

## all: Run all the build steps
.PHONY: all
all: tidy add-copyright lint cover build

## build: Build binaries by default
.PHONY: build
build: 
	@$(MAKE) go.build

## tidy: tidy go.mod
.PHONY: tidy
tidy:
	@$(GO) mod tidy

## deploy: Deploy k8sgpt
.PHONY: deploy
deploy: helm
	@echo "===========> Deploying k8sgpt"
	$(HELM) install k8sgpt charts/k8sgpt -n k8sgpt --create-namespace

## update: Update k8sgpt
.PHONY: update
update: helm
	@echo "===========> Updating k8sgpt"
	$(HELM) upgrade k8sgpt charts/k8sgpt -n k8sgpt

## undeploy: Undeploy k8sgpt
.PHONY: undeploy
undeploy: helm
	@echo "===========> Undeploying k8sgpt"
	$(HELM) uninstall k8sgpt -n k8sgpt

## docker-build: Build docker image
.PHONY: docker-build
docker-build:
	@$(MAKE) go.docker-build
## fmt: Run go fmt against code.
.PHONY: fmt
fmt:
	@$(GO) fmt ./...

## vet: Run go vet against code.
.PHONY: vet
vet:
	@$(GO) vet ./...

## lint: Check syntax and styling of go sources.
.PHONY: lint
lint:
	@$(MAKE) go.lint

## tools: Install dependent tools.
.PHONY: tools
tools:
	@$(MAKE) tools.install

## style: Code style -> fmt,vet,lint
.PHONY: style
style: fmt vet lint

## test: Run unit test.
.PHONY: test
test:
	@$(MAKE) go.test

## cover: Run unit test with coverage
.PHONY: cover
cover: test
	@$(MAKE) go.cover

## verify-copyright: Verify the license headers for all files.
.PHONY: verify-license
verify-license:
	@$(MAKE) copyright.verify

## add-license: Add copyright ensure source code files have license headers.
.PHONY: add-license
add-license:
	@$(MAKE) copyright.add

## imports: task to automatically handle import packages in Go files using goimports tool
.PHONY: imports
imports:
	@$(MAKE) go.imports

## clean: Remove all files that are created by building. 
.PHONY: clean
clean:
	@$(MAKE) go.clean

## help: Show this help info.
.PHONY: help
help: Makefile
	$(call makehelp)

## help-all: Show all help details info.
.PHONY: help-all
help-all: go.help copyright.help tools.help help
	$(call makeallhelp)

# =====
# Tools

HELM_VERSION ?= v3.11.3

helm:
	if ! test -f  $(OUTPUT_DIR)/helm-$(GOOS)-$(GOARCH); then \
		curl -L https://get.helm.sh/helm-$(HELM_VERSION)-$(GOOS)-$(GOARCH).tar.gz | tar xz; \
		mv $(GOOS)-$(GOARCH)/helm $(OUTPUT_DIR)/helm-$(GOOS)-$(GOARCH); \
		chmod +x $(OUTPUT_DIR)/helm-$(GOOS)-$(GOARCH); \
		rm -rf ./$(GOOS)-$(GOARCH)/; \
	fi
HELM=$(OUTPUT_DIR)/helm-$(GOOS)-$(GOARCH)