CONFIGDIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
ENTITY_SEED := $(CONFIGDIR)../../../examples/solutions/single-button-toggle/water_valve.yaml

# The screenshot config reads and writes this fixture, so captures never touch
# a real /etc/OliveTin install.
export WATER_VALVE_ENTITY_FILE := $(CONFIGDIR)water_valve.yaml

include ../../screenshots.mk

start: seed-entity seed-icons

.DEFAULT_GOAL := screenshots
.PHONY: screenshots seed-entity seed-icons

seed-entity:
	cp "$(ENTITY_SEED)" "$(WATER_VALVE_ENTITY_FILE)"

screenshots: update-screenshots

# OliveTin serves the action icons from custom-webui/ next to config.yaml.
seed-icons:
	mkdir -p "$(CONFIGDIR)custom-webui/icons"
	cp "$(CONFIGDIR)icons/"*.png "$(CONFIGDIR)custom-webui/icons/"
