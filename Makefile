.PHONY: build-proxy run-proxy

build-proxy:
	@$(MAKE) -C proxy build

run-proxy: build-proxy
	@$(MAKE) -C proxy run

build-knocker:
	@$(MAKE) -C knocker build

run-knocker: build-proxy
	@$(MAKE) -C knocker run
