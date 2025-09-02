MAKEFLAGS += --no-print-directory

TIME := /usr/bin/time -f ':: Time elapsed: %es'

THIS_MAKEFILE := $(realpath $(lastword $(MAKEFILE_LIST)))
THIS_DIR := $(patsubst %/,%,$(dir $(THIS_MAKEFILE)))

GAME := $(shell basename $(THIS_DIR))

help:
	@echo "Available targets:"
	@awk -F: '/^[^_.][A-Za-z0-9_-]*:/ { print "  " $$1 }' $(MAKEFILE_LIST)
.PHONY: help

image: _rcon
	docker build . \
		--quiet \
		-f $(THIS_DIR)/server.dockerfile \
		-t $(GAME)-server \
		>/dev/null
.PHONY: image

_rcon:
	@$(MAKE) -C $(THIS_DIR)/../../rcon image
.PHONY: rcon

shell: image
	@docker run --rm -it $(GAME)-server
.PHONY: shell
