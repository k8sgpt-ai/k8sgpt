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
# wget https://github.com/google/addlicense/releases/download/v1.0.0/addlicense_1.0.0_Linux_x86_64.tar.gz
# Makefile helper functions for copyright
#

LICENSE_TEMPLATE ?= $(ROOT_DIR)/scripts/LICENSE_TEMPLATE

# TODO: GOBIN -> TOOLS_DIR
## copyright.verify: Validate boilerplate headers for assign files
.PHONY: copyright.verify
copyright.verify: tools.verify.addlicense
	@echo "===========> Validate boilerplate headers for assign files starting in the $(ROOT_DIR) directory"
	@$(TOOLS_DIR)/addlicense -v -check -ignore **/test/** -f $(LICENSE_TEMPLATE) $(CODE_DIRS)
	@echo "===========> End of boilerplate headers check..."

## copyright.add: Add the boilerplate headers for all files
.PHONY: copyright.add
copyright.add: tools.verify.addlicense
	@echo "===========> Adding $(LICENSE_TEMPLATE) the boilerplate headers for all files"
	@$(TOOLS_DIR)/addlicense -y $(shell date +"%Y") -v -c "K8sGPT." -f $(LICENSE_TEMPLATE) $(CODE_DIRS)
	@echo "===========> End the copyright is added..."

## copyright.help: Show copyright help
.PHONY: copyright.help
copyright.help: scripts/make-rules/copyright.mk
	$(call smallhelp)
