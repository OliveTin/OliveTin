.PHONY: update-screenshots start stop

ROOT := $(abspath $(dir $(lastword $(MAKEFILE_LIST)))/../../../..)
OLIVETIN ?= $(ROOT)/service/OliveTin
PORT := 11337

ifndef CONFIGDIR
$(error CONFIGDIR must be set before including screenshots.mk)
endif

.DEFAULT_GOAL := update-screenshots

start:
	@set -e; \
	if curl -sf "http://localhost:$(PORT)/" >/dev/null 2>&1; then \
		echo "Port $(PORT) is already in use; run 'make stop' first"; \
		exit 1; \
	fi; \
	cd "$(ROOT)/service" && "$(OLIVETIN)" -configdir "$(CONFIGDIR)" & \
	pid=$$!; \
	for i in 1 2 3 4 5 6 7 8 9 10; do \
		if curl -sf "http://localhost:$(PORT)/" >/dev/null; then \
			exit 0; \
		fi; \
		if ! kill -0 $$pid 2>/dev/null; then \
			echo "OliveTin exited before listening on port $(PORT)"; \
			exit 1; \
		fi; \
		sleep 1; \
	done; \
	echo "Timed out waiting for OliveTin on port $(PORT)"; \
	exit 1

stop:
	@set +e; \
	if command -v fuser >/dev/null 2>&1; then \
		fuser -k $(PORT)/tcp 2>/dev/null; \
	else \
		for pid in $$(lsof -t -i :$(PORT) 2>/dev/null); do kill $$pid 2>/dev/null; done; \
	fi; \
	for i in 1 2 3 4 5; do \
		curl -sf "http://localhost:$(PORT)/" >/dev/null || exit 0; \
		sleep 1; \
	done; \
	exit 0

update-screenshots: stop start
	cd "$(CONFIGDIR)" && repo-helper screenshot --config screenshots.ini
	@$(MAKE) stop
