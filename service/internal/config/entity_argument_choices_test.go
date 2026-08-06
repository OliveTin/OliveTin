package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateEntityArgumentChoicesForArgument(t *testing.T) {
	assert.NoError(t, validateEntityArgumentChoicesForArgument("Reboot", ActionArgument{
		Name:   "target",
		Entity: "servers",
		Choices: []ActionArgumentChoice{
			{Value: "{{ servers.name }}"},
		},
	}))

	assert.Error(t, validateEntityArgumentChoicesForArgument("Reboot", ActionArgument{
		Name:   "target",
		Entity: "servers",
		Choices: []ActionArgumentChoice{
			{Value: "{{ servers.name }}"},
			{Value: "web01"},
		},
	}))

	assert.Error(t, validateEntityArgumentChoicesForArgument("Reboot", ActionArgument{
		Name:    "target",
		Entity:  "servers",
		Choices: nil,
	}))

	assert.NoError(t, validateEntityArgumentChoicesForArgument("Reboot", ActionArgument{
		Name: "plain",
		Choices: []ActionArgumentChoice{
			{Value: "a"},
			{Value: "b"},
		},
	}))
}
