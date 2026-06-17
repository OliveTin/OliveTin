package api

import (
	"testing"

	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/executor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildActiveBindingStates(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Actions = []*config.Action{
		{Title: "backup", Shell: "sleep 1", MaxConcurrent: 1},
		{Title: "ping", Shell: "echo ping"},
	}
	cfg.Sanitize()

	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()

	backupBinding := ex.FindBindingWithNoEntity(cfg.Actions[0])
	pingBinding := ex.FindBindingWithNoEntity(cfg.Actions[1])
	require.NotNil(t, backupBinding)
	require.NotNil(t, pingBinding)

	backupRunning := newAPIQueueLogEntry(backupBinding, true, false)
	backupWaiting := newAPIQueueLogEntry(backupBinding, false, false)
	pingRunning := newAPIQueueLogEntry(pingBinding, true, false)

	states := buildActiveBindingStates([]*executor.InternalLogEntry{
		backupRunning,
		backupWaiting,
		pingRunning,
	})

	backupState, ok := states[backupBinding.ID]
	require.True(t, ok)
	assert.True(t, backupState.hasRunning)
	assert.True(t, backupState.hasQueued)

	pingState, ok := states[pingBinding.ID]
	require.True(t, ok)
	assert.True(t, pingState.hasRunning)
	assert.False(t, pingState.hasQueued)
}

func TestBuildActionIncludesActiveBindingState(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Actions = []*config.Action{
		{Title: "backup", Shell: "sleep 1"},
	}
	cfg.Sanitize()

	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()

	binding := ex.FindBindingWithNoEntity(cfg.Actions[0])
	require.NotNil(t, binding)

	running := newAPIQueueLogEntry(binding, true, false)
	queued := newAPIQueueLogEntry(binding, false, false)
	ex.SetLog(running.ExecutionTrackingID, running)
	ex.SetLog(queued.ExecutionTrackingID, queued)

	rr := &DashboardRenderRequest{
		cfg: cfg,
		ex:  ex,
	}
	populateActiveBindingStates(rr)

	action := buildAction(binding, rr)
	require.NotNil(t, action)
	assert.True(t, action.HasRunningInstance)
	assert.True(t, action.HasQueuedInstance)
}

func TestBuildActiveBindingStatesIgnoresFinished(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Actions = []*config.Action{{Title: "backup", Shell: "sleep 1"}}
	cfg.Sanitize()

	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()
	binding := ex.FindBindingWithNoEntity(cfg.Actions[0])
	require.NotNil(t, binding)

	finished := newAPIQueueLogEntry(binding, true, true)
	states := buildActiveBindingStates([]*executor.InternalLogEntry{finished})
	assert.Empty(t, states)
}
