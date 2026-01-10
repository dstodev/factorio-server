MAKEFLAGS += --no-print-directory

TIME := /usr/bin/time -f ':: Time elapsed: %es'

THIS_MAKEFILE := $(realpath $(lastword $(MAKEFILE_LIST)))
THIS_DIR := $(patsubst %/,%,$(dir $(THIS_MAKEFILE)))

GO_PATH := $(shell go env GOPATH)
GO_BIN := $(if $(GOBIN),$(GOBIN),$(GO_PATH)/bin)

# Only include files after defining THIS_MAKEFILE above,
# because `include` appends to MAKEFILE_LIST

include script/util.mk

help:
	@echo "Available targets:"
# avoid listing targets of all makefiles by using THIS_MAKEFILE over MAKEFILE_LIST
	@awk -F: '/^[^_.][A-Za-z0-9_-]*:/ { print "  " $$1 }' $(THIS_MAKEFILE)
.PHONY: help

test:
	@$(MAKE) test-reset
	-@$(MAKE) test-run
	@$(MAKE) test-clean
.PHONY: test

test-reset:
	@go clean -testcache
.PHONY: test-reset

test-run: _gotest m2test-image
	@$(TIME) $(GO_BIN)/gotest $(GOFLAGS) ./...
.PHONY: test-run

test-clean:
	@docker container prune --force
.PHONY: test-clean

_gotest:
	@go install github.com/rakyll/gotest@latest
.PHONY: _gotest

m2test-image:
	@docker build docker \
		--quiet \
		-f $(THIS_DIR)/docker/m2test.dockerfile \
		-t m2test \
		>/dev/null
.PHONY: m2test-image

m2test-shell: m2test-image
	docker run --rm -it m2test
.PHONY: m2test-shell

# Build in the context of ../rcon, but use m2/docker/server.dockerfile instead
# of ../rcon/Dockerfile
server-image:
	@cd $(THIS_DIR)/..; docker build rcon \
		--quiet \
		-f $(THIS_DIR)/docker/server.dockerfile \
		-t m2-server \
		>/dev/null
.PHONY: server-image

server-shell: server-image
	@docker run --rm -it m2-server
.PHONY: server-shell

# List this makefile's variables
# see: https://stackoverflow.com/a/26339924
var-names:
# 	@$(MAKE) \
# 		--print-data-base \
# 		--question \
# 		--no-builtin-rules \
# 		--no-builtin-variables \
# 		--file $(THIS_MAKEFILE) \
# 		: 2>/dev/null |
	@grep --extended-regexp '^[^[:space:].#=]+[[:space:]]*[^=]?=' $(THIS_MAKEFILE) \
	| cut -d' ' -f1 \
	| LC_ALL=C sort
.PHONY: var-names
