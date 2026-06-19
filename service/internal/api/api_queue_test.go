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

func TestGetExecutionQueueGroupsByActionGroup(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.ActionGroups = map[string]*config.ActionGroup{
		"deploy": {MaxConcurrent: 2, Icon: "backup"},
	}
	cfg.Actions = []*config.Action{
		{Title: "backup", Shell: "sleep 1", MaxConcurrent: 1, Groups: []string{"deploy"}},
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
	backupWaiting.Queued = true
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

	deployGroup := findExecutionQueueGroup(resp.Msg.Groups, "deploy")
	defaultGroup := findExecutionQueueGroup(resp.Msg.Groups, defaultActionGroupName)
	require.NotNil(t, deployGroup)
	require.NotNil(t, defaultGroup)

	assert.Equal(t, int32(2), deployGroup.MaxConcurrent)
	assert.Equal(t, int32(5), deployGroup.QueueSize)
	assert.Equal(t, "&#128190;", deployGroup.Icon)
	assert.Equal(t, int32(2), deployGroup.ActiveCount)
	assert.Equal(t, int32(1), deployGroup.QueuedCount)
	require.Len(t, deployGroup.Actions, 1)
	assert.Equal(t, "backup", deployGroup.Actions[0].ActionTitle)
	require.Len(t, deployGroup.Actions[0].Entries, 2)

	require.Len(t, defaultGroup.Actions, 1)
	assert.Equal(t, "ping", defaultGroup.Actions[0].ActionTitle)
	assert.Equal(t, int32(1), defaultGroup.Actions[0].ActiveCount)
}

func findExecutionQueueGroup(groups []*apiv1.ExecutionQueueGroup, name string) *apiv1.ExecutionQueueGroup {
	for _, group := range groups {
		if group.Name == name {
			return group
		}
	}

	return nil
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
