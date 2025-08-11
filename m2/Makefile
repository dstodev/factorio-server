MAKEFLAGS += --no-print-directory

TIME := /usr/bin/time -f ':: Time elapsed: %es'

THIS_MAKEFILE := $(realpath $(lastword $(MAKEFILE_LIST)))
THIS_DIR := $(patsubst %/,%,$(dir $(THIS_MAKEFILE)))

GO_BIN := $(if $(GOBIN),$(GOBIN),$(shell go env GOPATH)/bin)

help:
	@echo "Available targets:"
	@awk -F: '/^[^_.][A-Za-z0-9_-]*:/ { print "  " $$1 }' $(MAKEFILE_LIST)
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
	@docker run --rm -it m2test
.PHONY: m2test-shell

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

vars:
	@echo "THIS_MAKEFILE: $(THIS_MAKEFILE)"
	@echo "THIS_DIR: $(THIS_DIR)"
	@echo "GOBIN: $(GOBIN)"
	@echo "GOPATH: $(shell go env GOPATH)"
	@echo "GO_BIN: $(GO_BIN)"
.PHONY: vars
