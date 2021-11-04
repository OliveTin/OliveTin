package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSanitizeConfig(t *testing.T) {
	c := DefaultConfig()

	a := Action{
		Title: "Mr Waffles",
		Arguments: []ActionArgument{
			ActionArgument{
				Name: "Carrots",
				Choices: []ActionArgumentChoice{
					{
						Value: "Waffle",
					},
				},
			},
			{
				Name: "foobar",
			},
		},
	}

	c.Actions = append(c.Actions, a)

	Sanitize(c)

	a2 := c.FindAction("Mr Waffles")

	assert.NotNil(t, a2, "Found action after adding it")
	assert.Equal(t, 3, a2.Timeout, "Default timeout is set")
	assert.Equal(t, "&#x1F600;", a2.Icon, "Default icon is a smiley")
	assert.Equal(t, "Carrots", a2.Arguments[0].Title, "Arg title is set to name")
	assert.Equal(t, "Waffle", a2.Arguments[0].Choices[0].Title, "Choice title is set to name")
}
