TIME := /usr/bin/time -f ":: Time elapsed: %es"

GO_BIN := $(if $(GOBIN),$(GOBIN),$(shell go env GOPATH)/bin)

test: _test-clean
.PHONY: test

_test-reset:
	@go clean -testcache
.PHONY: _test-reset

_test-run: _test-reset gotest m2test-image
	@-$(TIME) $(GO_BIN)/gotest $(GOFLAGS) ./...
.PHONY: _test-run

_test-clean: _test-run
	@docker container prune -f
.PHONY: _test-clean

gotest:
	@go install github.com/rakyll/gotest@latest
.PHONY: gotest

m2test-image:
	@docker build docker \
		--quiet \
		-f docker/m2test.dockerfile \
		-t m2test \
		>/dev/null
.PHONY: m2test-image

m2test-shell: m2test-image
	@docker run --rm -it m2test
.PHONY: m2test-shell

print-go-paths:
	@echo "GOBIN: $(GOBIN)"
	@echo "GOPATH: $(shell go env GOPATH)"
	@echo "GO_BIN: $(GO_BIN)"
.PHONY: print-go-paths
