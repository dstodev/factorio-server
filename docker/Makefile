gateway gw up:
	@shim/compose build gateway
	@shim/compose up --detach gateway
	@echo -n "Admin panel: "
	@util/gateway.sh
.PHONY: gateway gw up

down:
	@shim/compose down
.PHONY: down
