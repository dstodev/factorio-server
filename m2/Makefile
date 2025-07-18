TIME := /usr/bin/time -f ":: Time elapsed: %es"

test-image:
	@docker build docker \
		--quiet \
		-f docker/m2test.dockerfile \
		-t m2test \
		>/dev/null

.PHONY: test-image

test-shell: test-image
	docker run --rm -it m2test
.PHONY: test-shell

test: test-image
	@go clean -testcache
	@${TIME} go test -v ./...
	@docker container prune -f
.PHONY: test
