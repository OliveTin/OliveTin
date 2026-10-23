CONFIGDIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
ENTITY_FILE := /etc/OliveTin/entities/water_valve.yaml
ENTITY_SEED := $(CONFIGDIR)../../../examples/solutions/latching-toggle/water_valve.yaml

include ../../screenshots.mk

.DEFAULT_GOAL := screenshots
.PHONY: screenshots seed-entity

seed-entity:
	cp "$(ENTITY_SEED)" "$(ENTITY_FILE)"

screenshots: seed-entity
	$(MAKE) update-screenshots
	$(MAKE) seed-entity
