package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSanitizeConfig(t *testing.T) {
	c := DefaultConfig()

	a := &Action{
		Title: "Mr Waffles",
		Arguments: []ActionArgument{
			{
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
	c.Sanitize()

	a2 := c.findAction("Mr Waffles")

	assert.NotNil(t, a2, "Found action after adding it")
	assert.Equal(t, 3, a2.Timeout, "Default timeout is set")
	assert.Equal(t, "&#x1F600;", a2.Icon, "Default icon is a smiley")
	assert.Equal(t, "Carrots", a2.Arguments[0].Title, "Arg title is set to name")
	assert.Equal(t, "Waffle", a2.Arguments[0].Choices[0].Title, "Choice title is set to name")
}

func TestSanitizeConfigInlineDashboardActions(t *testing.T) {
	c := DefaultConfig()

	inline := &Action{
		Shell: "date",
	}

	dashboardActionTitle := "Inline Dashboard Action"

	c.Dashboards = []*DashboardComponent{
		{
			Title: "My Dashboard",
			Contents: []*DashboardComponent{
				{
					Title:        dashboardActionTitle,
					InlineAction: inline,
				},
			},
		},
	}

	c.Sanitize()

	// Inline action should have been appended to the global Actions slice.
	assert.GreaterOrEqual(t, len(c.Actions), 1, "At least one action should exist after sanitization")

	// It should be discoverable by the dashboard component title when no explicit title was set.
	found := c.findAction(dashboardActionTitle)
	if assert.NotNil(t, found, "Inline dashboard action should be discoverable by title") {
		assert.Equal(t, dashboardActionTitle, found.Title, "Inline action title should default from dashboard component title")
		assert.Equal(t, 3, found.Timeout, "Inline action should have default timeout applied")
		assert.NotEmpty(t, found.Icon, "Inline action should have default icon applied")
		assert.NotEmpty(t, found.ID, "Inline action should have a generated ID")
	}
}
