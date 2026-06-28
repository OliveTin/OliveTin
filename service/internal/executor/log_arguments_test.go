package executor

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	auth "github.com/OliveTin/OliveTin/internal/auth"
	config "github.com/OliveTin/OliveTin/internal/config"
)

func TestArgumentTypeStorableInLog(t *testing.T) {
	assert.True(t, argumentTypeStorableInLog("ascii"))
	assert.True(t, argumentTypeStorableInLog("shell_safe_identifier"))
	assert.False(t, argumentTypeStorableInLog("password"))
	assert.False(t, argumentTypeStorableInLog("very_dangerous_raw_string"))
}

func TestStorableArgumentsFromRequestExcludesSensitiveAndSystemArgs(t *testing.T) {
	req := newExecRequest()
	req.Binding.Action.Arguments = []config.ActionArgument{
		{Name: "host", Type: "ascii_identifier"},
		{Name: "secret", Type: "password"},
		{Name: "payload", Type: "very_dangerous_raw_string"},
	}
	req.Arguments = map[string]string{
		"host":                   "example.com",
		"secret":                 "hunter2",
		"payload":                "rm -rf /",
		"ot_executionTrackingId": "track-123",
		"ot_username":            "alice",
		"extra_undefined":        "drop-me",
	}

	args := storableArgumentsFromRequest(req)

	require.Len(t, args, 1)
	assert.Equal(t, "example.com", args["host"])
	assert.NotContains(t, args, "secret")
	assert.NotContains(t, args, "payload")
	assert.NotContains(t, args, "ot_executionTrackingId")
	assert.NotContains(t, args, "ot_username")
	assert.NotContains(t, args, "extra_undefined")
}

func TestStorableArgumentsFromRequestReturnsNilWhenEmpty(t *testing.T) {
	req := newExecRequest()
	req.Binding.Action.Arguments = []config.ActionArgument{
		{Name: "secret", Type: "password"},
	}
	req.Arguments = map[string]string{
		"secret": "hunter2",
	}

	assert.Nil(t, storableArgumentsFromRequest(req))
}

func TestStorableArgumentsFromRequestStoresMangledCheckboxValue(t *testing.T) {
	req := newExecRequest()
	req.Binding.Action.Arguments = []config.ActionArgument{
		{
			Name: "mode",
			Type: "checkbox",
			Choices: []config.ActionArgumentChoice{
				{Title: "Enabled", Value: "1"},
				{Title: "Disabled", Value: "0"},
			},
		},
	}
	req.Arguments = map[string]string{
		"mode": "Enabled",
	}

	mangleInvalidArgumentValues(req)
	args := storableArgumentsFromRequest(req)

	require.Len(t, args, 1)
	assert.Equal(t, "1", args["mode"])
}

func TestCopyStorableArgumentsToLogEntry(t *testing.T) {
	req := newExecRequest()
	req.logEntry = &InternalLogEntry{}
	req.Binding.Action.Arguments = []config.ActionArgument{
		{Name: "target", Type: "ascii_identifier"},
	}
	req.Arguments = map[string]string{
		"target": "server-a",
	}

	copyStorableArgumentsToLogEntry(req)

	require.NotNil(t, req.logEntry.Arguments)
	assert.Equal(t, "server-a", req.logEntry.Arguments["target"])
}

func TestExecRequestStoresArgumentsOnLogEntry(t *testing.T) {
	e, cfg := testingExecutor()

	e.RebuildActionMap()
	binding := e.FindBindingWithNoEntity(cfg.Actions[0])
	require.NotNil(t, binding)

	req := ExecutionRequest{
		Binding:           binding,
		Cfg:               cfg,
		AuthenticatedUser: auth.UserGuest(cfg),
		Arguments: map[string]string{
			"person": "yourself",
		},
	}

	wg, trackingID := e.ExecRequest(&req)
	wg.Wait()

	logEntry, ok := e.GetLog(trackingID)
	require.True(t, ok)
	require.NotNil(t, logEntry.Arguments)
	assert.Equal(t, "yourself", logEntry.Arguments["person"])
}

func TestRestartArgumentsIncompleteDetectsNonStorableArguments(t *testing.T) {
	action := &config.Action{
		Arguments: []config.ActionArgument{
			{Name: "host", Type: "ascii_identifier"},
			{Name: "pass", Type: "password"},
		},
	}

	assert.True(t, RestartArgumentsIncomplete(action, nil, map[string]string{
		"host": "db-1",
	}))
}

func TestRestartArgumentsIncompleteDetectsMissingRequiredStoredArguments(t *testing.T) {
	action := &config.Action{
		Arguments: []config.ActionArgument{
			{Name: "host", Type: "ascii_identifier"},
		},
	}

	assert.True(t, RestartArgumentsIncomplete(action, nil, map[string]string{}))
	assert.False(t, RestartArgumentsIncomplete(action, nil, map[string]string{
		"host": "db-1",
	}))
}

func TestRestartArgumentsIncompleteAllowsOptionalArgumentsWithDefaults(t *testing.T) {
	action := &config.Action{
		Arguments: []config.ActionArgument{
			{Name: "host", Type: "ascii_identifier", Default: "example.com"},
		},
	}

	assert.False(t, RestartArgumentsIncomplete(action, nil, map[string]string{}))
}
