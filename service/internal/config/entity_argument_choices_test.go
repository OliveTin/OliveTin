package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateEntityArgumentChoicesForArgument(t *testing.T) {
	require.NoError(t, validateEntityArgumentChoicesForArgument("Reboot", ActionArgument{
		Name:   "target",
		Entity: "servers",
		Choices: []ActionArgumentChoice{
			{Value: "{{ servers.name }}"},
		},
	}))

	require.Error(t, validateEntityArgumentChoicesForArgument("Reboot", ActionArgument{
		Name:   "target",
		Entity: "servers",
		Choices: []ActionArgumentChoice{
			{Value: "{{ servers.name }}"},
			{Value: "web01"},
		},
	}))

	require.Error(t, validateEntityArgumentChoicesForArgument("Reboot", ActionArgument{
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
