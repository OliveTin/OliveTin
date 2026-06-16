package api

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	apiv1 "github.com/OliveTin/OliveTin/gen/olivetin/api/v1"
	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/executor"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetExecutionQueueGroupsByBinding(t *testing.T) {
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

	ex.SetLog(backupRunning.ExecutionTrackingID, backupRunning)
	ex.SetLog(backupWaiting.ExecutionTrackingID, backupWaiting)
	ex.SetLog(pingRunning.ExecutionTrackingID, pingRunning)

	ts, client := getNewTestServerAndClientWithExecutor(cfg, ex)
	defer ts.Close()

	resp, err := client.GetExecutionQueue(context.Background(), connect.NewRequest(&apiv1.GetExecutionQueueRequest{}))
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, int32(3), resp.Msg.TotalActive)
	require.Len(t, resp.Msg.Groups, 2)

	var backupGroup *apiv1.ExecutionQueueGroup
	for _, group := range resp.Msg.Groups {
		if group.BindingId == backupBinding.ID {
			backupGroup = group
		}
	}

	require.NotNil(t, backupGroup)
	assert.Equal(t, "backup", backupGroup.ActionTitle)
	assert.Equal(t, int32(1), backupGroup.MaxConcurrent)
	assert.Equal(t, int32(2), backupGroup.ActiveCount)
	require.Len(t, backupGroup.Entries, 2)
	assert.False(t, backupGroup.Entries[1].ExecutionStarted)
	assert.True(t, backupGroup.Entries[0].ExecutionStarted)
}

func newAPIQueueLogEntry(binding *executor.ActionBinding, started bool, finished bool) *executor.InternalLogEntry {
	startedAt := time.Now().Add(-time.Minute)
	if started {
		startedAt = time.Now().Add(-2 * time.Minute)
	}

	entry := &executor.InternalLogEntry{
		Binding:             binding,
		DatetimeStarted:     startedAt,
		ExecutionTrackingID: uuid.NewString(),
		ActionTitle:         binding.Action.Title,
		ExecutionStarted:    started,
		ExecutionFinished:   finished,
	}

	if finished {
		entry.DatetimeFinished = time.Now()
	}

	return entry
}
