CONFIGDIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
ENTITY_SEED := $(CONFIGDIR)../../../examples/solutions/latching-toggle/water_valve.yaml

# The screenshot config reads and writes this fixture, so captures never touch
# a real /etc/OliveTin install.
export WATER_VALVE_ENTITY_FILE := $(CONFIGDIR)water_valve.yaml

include ../../screenshots.mk

start: seed-entity

.DEFAULT_GOAL := screenshots
.PHONY: screenshots seed-entity

seed-entity:
	cp "$(ENTITY_SEED)" "$(WATER_VALVE_ENTITY_FILE)"

screenshots: update-screenshots
