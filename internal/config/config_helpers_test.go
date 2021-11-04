package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFindAction(t *testing.T) {
	c := DefaultConfig()

	a1 := Action{}
	a1.Title = "a1"
	c.Actions = append(c.Actions, a1)

	a2 := Action{
		Title: "a2",
		Arguments: []ActionArgument{
			{
				Name: "Blat",
			},
		},
	}

	c.Actions = append(c.Actions, a2)

	assert.NotNil(t, c.FindAction("a1"), "Find action a1")

	assert.NotNil(t, c.FindAction("a2"), "Find action a2")
	assert.NotNil(t, c.FindAction("a2").FindArg("Blat"), "Find action argument")
	assert.Nil(t, c.FindAction("a2").FindArg("Blatey Cake"), "Find non-existent action argument")

	assert.Nil(t, c.FindAction("waffles"), "Find non-existent action")
}
