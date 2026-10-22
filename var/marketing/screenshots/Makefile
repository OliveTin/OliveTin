# Regenerates marketing screenshots against a throwaway OliveTin that runs
# the fictional config in seed/. Run `make` here (or `make screenshots` from
# the repo root); OliveTin is always stopped afterwards.
#
# OliveTin keeps all of its state in config files and memory, so there are
# no database or broker containers to start.

ROOT := $(abspath ../../..)
HERE := $(abspath .)
TMP := $(HERE)/tmp

BASE_URL := http://127.0.0.1:18337
API := $(BASE_URL)/api/olivetin.api.v1.OliveTinApiService

.PHONY: all build up seed shoot manifest down clean

all:
	@$(MAKE) build up seed shoot manifest; rc=$$?; $(MAKE) down; exit $$rc

build:
	$(MAKE) -C $(ROOT) service webui-dist

up: down
	@if curl -fsS $(BASE_URL)/ >/dev/null 2>&1; then \
		echo "Port 18337 is already in use"; exit 1; fi
	rm -rf $(TMP)
	mkdir -p $(TMP)
	cp -r seed $(TMP)/config
	# Started from service/ so that OliveTin finds ../webui.
	cd $(ROOT)/service && { ./OliveTin -configdir $(TMP)/config > $(TMP)/service.log 2>&1 & echo $$! > $(TMP)/service.pid; }
	@for i in $$(seq 60); do curl -fsS -X POST -H 'Content-Type: application/json' -d '{}' \
		$(API)/Init >/dev/null 2>&1 && exit 0; sleep 0.5; done; \
		echo "OliveTin did not start; see $(TMP)/service.log"; exit 1

seed:
	./seed.sh $(BASE_URL)

shoot:
	repo-helper screenshot --config screenshots.ini

manifest:
	./generate-manifest.py

down:
	-@[ -f $(TMP)/service.pid ] && kill $$(cat $(TMP)/service.pid) 2>/dev/null; rm -f $(TMP)/service.pid

clean: down
	rm -rf $(TMP)
