package config

import (
	"bytes"
	"runtime"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	assert.Equal(t, "hugeicons:CommandLineIcon", a2.Icon, "Default icon is the neutral CLI glyph")
	assert.Equal(t, "Carrots", a2.Arguments[0].Title, "Arg title is set to name")
	assert.Equal(t, "Waffle", a2.Arguments[0].Choices[0].Title, "Choice title is set to name")
}

func TestSanitizePopupOnStartHistory(t *testing.T) {
	c := DefaultConfig()
	c.DefaultPopupOnStart = "nothing"

	c.Actions = append(c.Actions, &Action{
		Title:        "With history",
		PopupOnStart: "history",
		Shell:        "true",
	})
	c.Sanitize()

	a := c.findAction("With history")
	if assert.NotNil(t, a) {
		assert.Equal(t, "history", a.PopupOnStart, "history must be preserved, not replaced by defaultPopupOnStart")
	}
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

func TestValidateReservedActionArgumentNames(t *testing.T) {
	c := DefaultConfig()
	c.Actions = append(c.Actions, &Action{
		Title: "Reserved arg",
		Arguments: []ActionArgument{
			{Name: "ot_custom", Type: "ascii"},
		},
	})

	err := c.validateReservedActionArgumentNames()

	require.Error(t, err)
	assert.Contains(t, err.Error(), `action "Reserved arg" argument "ot_custom" uses reserved prefix "ot_"`)
}

func TestSanitizeActionGroupsDedupesGroupNames(t *testing.T) {
	c := DefaultConfig()
	c.ActionGroups = map[string]*ActionGroup{
		"unity": {MaxConcurrent: 1},
	}
	c.Actions = append(c.Actions, &Action{
		Title:  "Build",
		Shell:  "true",
		Groups: []string{"unity", "unity", ""},
	})

	c.Sanitize()

	action := c.findAction("Build")
	require.NotNil(t, action)
	assert.Equal(t, []string{"unity"}, action.Groups)
}

func TestValidateReservedActionArgumentNamesAllowsNonReserved(t *testing.T) {
	c := DefaultConfig()
	c.Actions = append(c.Actions, &Action{
		Title: "Allowed arg",
		Arguments: []ActionArgument{
			{Name: "target", Type: "ascii"},
		},
	})

	require.NoError(t, c.validateReservedActionArgumentNames())
}

func TestValidateReservedActionArgumentNamesChecksInlineActions(t *testing.T) {
	c := DefaultConfig()
	c.Dashboards = []*DashboardComponent{
		{
			Title: "Dashboard",
			Contents: []*DashboardComponent{
				{
					Title: "Inline reserved arg",
					InlineAction: &Action{
						Shell: "echo test",
						Arguments: []ActionArgument{
							{Name: "ot_custom", Type: "ascii"},
						},
					},
				},
			},
		},
	}

	c.sanitizeDashboardsForInlineActions()
	err := c.validateReservedActionArgumentNames()

	require.Error(t, err)
	assert.Contains(t, err.Error(), `action "Inline reserved arg" argument "ot_custom" uses reserved prefix "ot_"`)
}

func TestValidateUniqueLocalUserAPIKeys(t *testing.T) {
	t.Parallel()

	err := validateUniqueLocalUserAPIKeys([]*LocalUser{
		{Username: "a", ApiKey: "same"},
		{Username: "b", ApiKey: "same"},
	})
	require.Error(t, err)

	err = validateUniqueLocalUserAPIKeys([]*LocalUser{
		{Username: "a", ApiKey: "one"},
		{Username: "b", ApiKey: "two"},
	})
	require.NoError(t, err)
}

func TestSanitizeServiceLogsUnsupportedPlatform(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("serviceLogs.directory platform check only applies on non-Windows")
	}

	var logBuffer bytes.Buffer
	previousOutput := logrus.StandardLogger().Out
	logrus.SetOutput(&logBuffer)
	t.Cleanup(func() {
		logrus.SetOutput(previousOutput)
	})

	cfg := DefaultConfig()
	cfg.ServiceLogs.Directory = "/var/log/OliveTin"
	cfg.Sanitize()

	assert.Contains(t, logBuffer.String(), "serviceLogs.directory is configured but this option is only supported on Windows")
}
