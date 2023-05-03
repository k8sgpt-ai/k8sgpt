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
# Makefile helper functions for tools(https://github.com/avelino/awesome-go) -> DIR: {TOOT_DIR}/tools | (go >= 1.20)
# Why download to the tools directory, thinking we might often switch Go versions using gvm.
#

# K8sGPT build use BUILD_TOOLS
BUILD_TOOLS ?= golangci-lint goimports addlicense helm

HELM_VERSION ?= v3.11.3

## tools.install: Install a must tools
.PHONY: tools.install
tools.install: $(addprefix tools.install., $(BUILD_TOOLS))

## tools.install.%: Install a single tool in $GOBIN/
.PHONY: tools.install.%
tools.install.%:
	@echo "===========> Installing $,The default installation path is $(GOBIN)/$*"
	@$(MAKE) install.$*

## tools.verify.%: Check if a tool is installed and install it
.PHONY: tools.verify.%
tools.verify.%:
	@echo "===========> Verifying $* is installed"
	@if [ ! -f $(TOOLS_DIR)/$* ]; then GOBIN=$(TOOLS_DIR) $(MAKE) tools.install.$*; fi
	@echo "===========> $* is install in $(TOOLS_DIR)/$*"

.PHONY: install.golangci-lint
## install.golangci-lint: Install golangci-lint
install.golangci-lint:
	@echo "===========> Installing golangci-lint,The default installation path is $(GOBIN)/golangci-lint"
	@$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

## install.goimports: Install goimports, used to format go source files
.PHONY: install.goimports
install.goimports:
	@echo "===========> Installing goimports,The default installation path is $(GOBIN)/goimports"
	@$(GO) install golang.org/x/tools/cmd/goimports@latest

## install.addlicense: Install addlicense, used to add license header to source files
.PHONY: install.addlicense
install.addlicense:
	@$(GO) install github.com/google/addlicense@latest

## install.helm: Install helm, used to deploy k8sgpt
HELM_VERSION ?= v3.11.3

## install.helm: Install helm, used to deploy k8sgpt
.PHONY: install.helm
install.helm:
	@echo "===========> Installing helm,The default installation path is $(TOOLS_DIR)/helm-$(GOOS)-$(GOARCH)"
	if ! test -f  $(TOOLS_DIR)/helm-$(GOOS)-$(GOARCH); then \
		curl -L https://get.helm.sh/helm-$(HELM_VERSION)-$(GOOS)-$(GOARCH).tar.gz | tar xz; \
		mv $(GOOS)-$(GOARCH)/helm $(TOOLS_DIR)/helm-$(GOOS)-$(GOARCH); \
		chmod +x $(TOOLS_DIR)/helm-$(GOOS)-$(GOARCH); \
		rm -rf ./$(GOOS)-$(GOARCH)/; \
	fi

# ==============================================================================
# Tools that might be used include go gvm
#

## install.ginkgo: Install ginkgo to run a single test or set of tests
.PHONY: install.ginkgo
install.ginkgo:
	@echo "===========> Installing ginkgo,The default installation path is $(GOBIN)/ginkgo"
	@$(GO) install github.com/onsi/ginkgo/ginkgo@latest

## install.kube-score: Install kube-score, used to check kubernetes yaml files
.PHONY: install.kube-score
install.kube-score:
	@$(GO) install github.com/zegl/kube-score/cmd/kube-score@latest

## Install go-gitlint: Install go-gitlint, used to check git commit message
.PHONY: install.go-gitlint
install.go-gitlint:
	@$(GO) install github.com/marmotedu/go-gitlint/cmd/go-gitlint@latest

## install.git-chglog: Install git-chglog, used to generate changelog
.PHONY: install.git-chglog
install.git-chglog:
	@$(GO) install github.com/git-chglog/git-chglog/cmd/git-chglog@latest

## install.github-release: Install github-release, used to create github release
.PHONY: install.github-release
install.github-release:
	@$(GO) install github.com/github-release/github-release@latest

## install.gvm: Install gvm, gvm is a Go version manager, built on top of the official go tool.
.PHONY: install.gvm
install.gvm:
	@echo "===========> Installing gvm,The default installation path is ~/.gvm/scripts/gvm"
	@bash < <(curl -s -S -L https://raw.gitee.com/moovweb/gvm/master/binscripts/gvm-installer)
	@$(shell source /root/.gvm/scripts/gvm)

## install.golines: Install golines, used to format long lines
.PHONY: install.golines
install.golines:
	@$(GO) install github.com/segmentio/golines@latest

## install.go-mod-outdated: Install go-mod-outdated, used to check outdated dependencies
.PHONY: install.go-mod-outdated
install.go-mod-outdated:
	@$(GO) install github.com/psampaz/go-mod-outdated@latest

## install.mockgen: Install mockgen, used to generate mock functions
.PHONY: install.mockgen
install.mockgen:
	@$(GO) install github.com/golang/mock/mockgen@latest

## install.gotests: Install gotests, used to generate test functions
.PHONY: install.gotests
install.gotests:
	@$(GO) install github.com/cweill/gotests/gotests@latest

## install.protoc-gen-go: Install protoc-gen-go, used to generate go source files from protobuf files
.PHONY: install.protoc-gen-go
install.protoc-gen-go:
	@$(GO) install github.com/golang/protobuf/protoc-gen-go@latest

## install.cfssl: Install cfssl, used to generate certificates
.PHONY: install.cfssl
install.cfssl:
	@$(ROOT_DIR)/scripts/install/install.sh iam::install::install_cfssl

## install.depth: Install depth, used to check dependency tree
.PHONY: install.depth
install.depth:
	@$(GO) install github.com/KyleBanks/depth/cmd/depth@latest

## install.go-callvis: Install go-callvis, used to visualize call graph
.PHONY: install.go-callvis
install.go-callvis:
	@$(GO) install github.com/ofabry/go-callvis@latest

## install.codegen: Install code generator, used to generate code
.PHONY: install.codegen
install.codegen:
	@$(GO) install ${ROOT_DIR}/tools/codegen/codegen.go

## tools.help: Display help information about the tools package
.PHONY: tools.help
tools.help: scripts/make-rules/tools.mk
	$(call smallhelp)
