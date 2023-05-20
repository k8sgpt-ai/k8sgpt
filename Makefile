# Copyright 2023 K8sgpt AI. All rights reserved.
# Use of this source code is governed by a MIT style
# license that can be found in the LICENSE file.

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

# include the common makefile
COMMON_SELF_DIR := $(dir $(lastword $(MAKEFILE_LIST)))
# ROOT_DIR: root directory of the code base
ifeq ($(origin ROOT_DIR),undefined)
ROOT_DIR := $(abspath $(shell cd $(COMMON_SELF_DIR)/. && pwd -P))
endif
# OUTPUT_DIR: The directory where the build output is stored.
ifeq ($(origin OUTPUT_DIR),undefined)
OUTPUT_DIR := $(ROOT_DIR)/bin
$(shell mkdir -p $(OUTPUT_DIR))
endif

ifeq ($(origin VERSION), undefined)
VERSION := $(shell git describe --abbrev=0 --dirty --always --tags | sed 's/-/./g')
endif

# Check if the tree is dirty. default to dirty(maybe u should commit?)
GIT_TREE_STATE:="dirty"
ifeq (, $(shell git status --porcelain 2>/dev/null))
	GIT_TREE_STATE="clean"
endif
GIT_COMMIT:=$(shell git rev-parse HEAD)

IMG ?= ghcr.io/k8sgpt-ai/k8sgpt:latest

BUILDFILE = "./main.go"
BUILDAPP = "$(OUTPUT_DIR)/k8sgpt"

.PHONY: all
all: tidy add-copyright lint cover build

# ==============================================================================
# Targets

## build: Build binaries by default
.PHONY: build
build: 
	@echo "$(shell go version)"
	@echo "===========> Building binary $(BUILDAPP) *[Git Info]: $(VERSION)-$(GIT_COMMIT)"
	@export CGO_ENABLED=0 && go build -o $(BUILDAPP) -ldflags "-s -w -X main.version=dev -X main.commit=$$(git rev-parse --short HEAD) -X main.date=$$(date +%FT%TZ)" $(BUILDFILE)

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
	@echo "===========> Building docker image"
	docker buildx build --build-arg=VERSION="$$(git describe --tags --abbrev=0)" --build-arg=COMMIT="$$(git rev-parse --short HEAD)" --build-arg DATE="$$(date +%FT%TZ)" --platform="linux/amd64,linux/arm64" -t ${IMG} -f container/Dockerfile . --push

## fmt: Run go fmt against code.
.PHONY: fmt
fmt:
	@$(GO) fmt ./...

## vet: Run go vet against code.
.PHONY: vet
vet:
	@$(GO) vet ./...

## lint: Run go lint against code.
.PHONY: lint
lint:
	@golangci-lint run -v ./...

## style: Code style -> fmt,vet,lint
.PHONY: style
style: fmt vet lint

## test: Run unit test
.PHONY: test
test: 
	@echo "===========> Run unit test"
	@$(GO) test ./... 

## cover: Run unit test with coverage
.PHONY: cover
cover: test
	@$(GO) test -cover

## go.clean: Clean all builds
.PHONY: clean
clean:
	@echo "===========> Cleaning all builds OUTPUT_DIR($(OUTPUT_DIR))"
	@-rm -vrf $(OUTPUT_DIR)
	@echo "===========> End clean..."

## help: Show this help info.
.PHONY: help
help: Makefile
	@printf "\n\033[1mUsage: make <TARGETS> ...\033[0m\n\n\\033[1mTargets:\\033[0m\n\n"
	@sed -n 's/^##//p' $< | awk -F':' '{printf "\033[36m%-28s\033[0m %s\n", $$1, $$2}' | sed -e 's/^/ /'

## copyright.verify: Validate boilerplate headers for assign files
.PHONY: copyright.verify
copyright.verify: tools.verify.addlicense
	@echo "===========> Validate boilerplate headers for assign files starting in the $(ROOT_DIR) directory"
#	@addlicense -v -check -ignore **/test/** -f $(LICENSE_TEMPLATE) $(CODE_DIRS)
	@echo "===========> End of boilerplate headers check..."

## copyright.add: Add the boilerplate headers for all files
.PHONY: copyright.add
copyright.add: tools.verify.addlicense
	@echo "===========> Adding $(LICENSE_TEMPLATE) the boilerplate headers for all files"
#	@addlicense -y $(shell date +"%Y") -v -c "K8sgpt AI." -f $(LICENSE_TEMPLATE) $(CODE_DIRS)
	@echo "===========> End the copyright is added..."

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