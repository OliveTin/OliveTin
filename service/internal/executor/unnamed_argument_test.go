package executor

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	authpublic "github.com/OliveTin/OliveTin/internal/auth/authpublic"
	config "github.com/OliveTin/OliveTin/internal/config"
)

func unnamedConfirmationAction(shell string) *config.Action {
	return &config.Action{
		ID:    "unnamed-confirmation",
		Title: "Unnamed confirmation",
		Shell: shell,
		Arguments: []config.ActionArgument{
			{Type: "confirmation", Title: "Are you sure?"},
			{Name: "target", Type: "ascii_identifier"},
		},
	}
}

func TestFilterToDefinedArgumentsDropsEmptyName(t *testing.T) {
	req := newExecRequest()
	req.Binding.Action = unnamedConfirmationAction("echo test")
	req.Arguments = map[string]string{
		"":       "x; touch /tmp/pwned #",
		"target": "server1",
	}

	filterToDefinedArgumentsOnly(req)

	assert.NotContains(t, req.Arguments, "", "an argument with an empty name must never be kept")
	assert.Equal(t, "server1", req.Arguments["target"])
}

func runUnnamedConfirmationAction(t *testing.T, shell string, args map[string]string) *InternalLogEntry {
	t.Helper()

	cfg := config.DefaultConfig()
	cfg.Actions = []*config.Action{unnamedConfirmationAction(shell)}
	cfg.Sanitize()

	ex := DefaultExecutor(cfg)
	ex.RebuildActionMap()

	req := &ExecutionRequest{
		Binding:           ex.FindBindingByID("unnamed-confirmation"),
		Arguments:         args,
		AuthenticatedUser: &authpublic.AuthenticatedUser{Username: "alice"},
		Cfg:               cfg,
	}
	require.NotNil(t, req.Binding)

	wg, _ := ex.ExecRequest(req)
	wg.Wait()

	entry, ok := ex.GetLog(req.TrackingID)
	require.True(t, ok)

	return entry
}

func TestEmptyNameArgumentNeverReachesShellTemplate(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "pwned")

	entry := runUnnamedConfirmationAction(t, `echo "value=[{{ index .Arguments "" }}] target={{ .Arguments.target }}"`, map[string]string{
		"":       "x; touch " + marker + " #",
		"target": "server1",
	})

	assert.NoFileExists(t, marker, "the empty-name argument must not be rendered into the shell command")
	assert.NotContains(t, entry.Output, "touch")
	assert.Contains(t, entry.Output, "value=[]")
}

func TestUnnamedConfirmationActionStillRuns(t *testing.T) {
	entry := runUnnamedConfirmationAction(t, `echo "running against {{ .Arguments.target }}"`, map[string]string{
		"target": "server1",
	})

	assert.False(t, entry.Blocked)
	assert.Equal(t, int32(0), entry.ExitCode)
	assert.Contains(t, entry.Output, "running against server1")
}
