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

GO := go
# !I have only tested gvm since 1.20. You are advised to use the gvm switchover version of the tools toolkit
GO_SUPPORTED_VERSIONS ?= |1.20|1.21|

ifeq ($(ROOT_PACKAGE),)
	$(error the variable ROOT_PACKAGE must be set prior to including golang.mk, -> /Makefile)
endif

GOPATH ?= $(shell go env GOPATH)
ifeq ($(origin GOBIN), undefined)
	GOBIN := $(GOPATH)/bin
endif

define funcsecret
ifndef SECRET
	$(error SECRET environment variable is not set)
endif
	kubectl create ns k8sgpt || true
	kubectl create secret generic ai-backend-secret --from-literal=secret-key=$(SECRET) --namespace=k8sgpt || true
	kubectl apply -f container/manifests
endef

## go.build: Build binaries by default
.PHONY: go.build
go.build: 
	@echo "$(shell go version)"
	@echo "===========> Building binary $(BUILDAPP) *[Git Info]: $(VERSION)-$(GIT_COMMIT)"
	@export CGO_ENABLED=0 && go build -o $(BUILDAPP) -ldflags '-s -w' $(BUILDFILE)

## go.deploy: Deploy k8sgpt
.PHONY: go.deploy
go.deploy:
	@echo "===========> Deploying k8sgpt"
	@$(call funcsecret)

## go.undeploy: Undeploy k8sgpt
.PHONY: go.undeploy
go.undeploy:
	@echo "===========> Undeploying k8sgpt"
	kubectl delete secret ai-backend-secret --namespace=k8sgpt
	kubectl delete -f container/manifests
	kubectl delete ns k8sgpt

## go.docker-build: Build docker image
.PHONY: go.docker-build
go.docker-build:
	@echo "===========> Building docker image"
	docker buildx build --build-arg=VERSION="$$(git describe --tags --abbrev=0)" --build-arg=COMMIT="$$(git rev-parse --short HEAD)" --build-arg DATE="$$(date +%FT%TZ)" --platform="linux/amd64,linux/arm64" -t ${IMG} -f container/Dockerfile . --push

## go.imports: task to automatically handle import packages in Go files using goimports tool
.PHONY: go.imports
go.imports: tools.verify.goimports
	@$(TOOLS_DIR)/goimports -l -w $(SRC)

## go.lint: Run the golangci-lint
.PHONY: go.lint
go.lint: tools.verify.golangci-lint
	@echo "===========> Run golangci to lint source codes"
	@$(TOOLS_DIR)/golangci-lint run -c $(ROOT_DIR)/.golangci.yml $(ROOT_DIR)/...

## go.test: Run unit test
.PHONY: go.test
go.test: #tools.verify.go-junit-report
	@echo "===========> Run unit test"
	@$(GO) test ./... 

## go.cover: Run unit test with coverage
.PHONY: go.cover
go.cover: go.test
	@touch $(OUTPUT_DIR)/coverage.out
	@$(GO) tool cover -func=$(OUTPUT_DIR)/coverage.out | \
		awk -v target=$(COVERAGE) -f $(ROOT_DIR)/scripts/coverage.awk

## copyright.help: Show copyright help
.PHONY: go.help
go.help: scripts/make-rules/golang.mk
	$(call smallhelp)

## go.clean: Clean all builds
.PHONY: go.clean
go.clean:
	@echo "===========> Cleaning all builds OUTPUT_DIR($(OUTPUT_DIR))"
	@-rm -vrf $(OUTPUT_DIR)
	@echo "===========> End clean..."