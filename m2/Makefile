TIME := /usr/bin/time -f ":: Time elapsed: %es"

GO_BIN := $(if $(GOBIN),$(GOBIN),$(shell go env GOPATH)/bin)

shell: image-m2test
	docker run --rm -it m2test
.PHONY: shell

test: _test-clean
.PHONY: test

_test-reset:
	@go clean -testcache
.PHONY: reset-test-cache

_test-run: _test-reset gotest image-m2test
	@-$(TIME) $(GO_BIN)/gotest $(GOFLAGS) ./...
.PHONY: test

_test-clean: _test-run
	@docker container prune -f
.PHONY: _test-clean

gotest:
	@go install github.com/rakyll/gotest@latest
.PHONY: gotest

image-m2test:
	@docker build docker \
		--quiet \
		-f docker/m2test.dockerfile \
		-t m2test \
		>/dev/null
.PHONY: image-m2test

print-go-paths:
	@echo "GOBIN: $(GOBIN)"
	@echo "GOPATH: $(shell go env GOPATH)"
	@echo "GO_BIN: $(GO_BIN)"
.PHONY: print-go-paths
