TIME := /usr/bin/time -f ":: Time elapsed: %es"

test:
	@go clean -testcache
	@${TIME} go test -v ./...
.PHONY: test
