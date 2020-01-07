SHELL      := /bin/bash
CURDIR     := $(shell readlink -f ./)
GOSOURCES  := $(shell find * ! -path 'vendor*' -type f -name '*.go' )
GOPKG      ?= $(shell grep '^module' go.mod | cut -d' ' -f 2)  # assumes go.mod
GO_DIR     := $(GOPATH)/src/$(GOPKG)
BIN_NAME   := $(shell basename $(GOPKG))

# Modify the current path to use locally built tools
PATH := $(shell readlink -f ./bin/linux_amd64):$(shell readlink -f ./vendor/bin):$(PATH)

# Ensure go compiles statically-linked binaries with "ldflags"
GO_BUILD_FLAGS := -ldflags "-linkmode external -extldflags -static"

#####################################################################
# checks

ifndef GOPATH
$(error ERROR: GOPATH must be declared)
endif

#####################################################################
# targets

.DEFAULT_GOAL=help

.PHONY: _deps
_deps:
	@# Link this current golang project directly into the GOPATH/src
	@#   Golang needs to find the sources of this project in the GOPATH.
	@if [ ! -L $(GO_DIR) ]; then \
	  mkdir -p $(shell dirname $(GO_DIR)); \
	  ln -s $(CURDIR) $(GO_DIR); \
	fi

.PHONY: _check
_check: _deps
	@# This is a check to make sure you run this makefile from within the GOPATH
	@#  Go requires building to be be run within GOPATH
	@if ! pwd | grep "$$GOPATH" > /dev/null; then \
	  echo "Cannot build unless within GOPATH workspace at $$GOPATH"; \
	  echo "Please change to this current directory within GOPATH"; \
	  echo ""; \
	  echo "  cd $(GO_DIR)"; \
	  echo ""; \
	  exit 1; \
	fi

.PHONY: _test
_test:

.PHONY: printvars
printvars:
	@$(foreach V, $(sort $(.VARIABLES)), $(if $(filter-out environment% default automatic, $(origin $V)), $(warning $V=$($V) )))
	@exit 0

help: ## Print list of Makefile targets
	@# Taken from https://github.com/spf13/hugo/blob/master/Makefile
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
	  cut -d ":" -f2- | \
	  awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
