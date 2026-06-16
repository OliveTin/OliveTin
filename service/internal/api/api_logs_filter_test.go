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

func TestGetLogsFilterExpression(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Actions = []*config.Action{
		{Title: "Update packages", Shell: "echo update"},
		{Title: "Ping host", Shell: "echo ping"},
	}
	cfg.Sanitize()

	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()

	updateBinding := ex.FindBindingWithNoEntity(cfg.Actions[0])
	pingBinding := ex.FindBindingWithNoEntity(cfg.Actions[1])
	require.NotNil(t, updateBinding)
	require.NotNil(t, pingBinding)

	ex.SetLog(uuid.NewString(), finishedLogEntry(updateBinding, "Update packages", "Completed"))
	ex.SetLog(uuid.NewString(), finishedLogEntry(pingBinding, "Ping host", "Blocked"))

	ts, client := getNewTestServerAndClientWithExecutor(cfg, ex)
	defer ts.Close()

	resp, err := client.GetLogs(context.Background(), connect.NewRequest(&apiv1.GetLogsRequest{
		Filter: "!Update",
	}))
	require.NoError(t, err)
	require.Len(t, resp.Msg.Logs, 1)
	assert.Equal(t, "Ping host", resp.Msg.Logs[0].ActionTitle)
}

func TestGetLogsInvalidFilterReturnsError(t *testing.T) {
	cfg := config.DefaultConfig()
	ts, client := getNewTestServerAndClient(cfg)
	defer ts.Close()

	_, err := client.GetLogs(context.Background(), connect.NewRequest(&apiv1.GetLogsRequest{
		Filter: `SecretField == "x"`,
	}))
	require.Error(t, err)
	assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
}

func finishedLogEntry(binding *executor.ActionBinding, title, status string) *executor.InternalLogEntry {
	entry := &executor.InternalLogEntry{
		Binding:             binding,
		DatetimeStarted:     time.Now(),
		DatetimeFinished:    time.Now(),
		ExecutionTrackingID: uuid.NewString(),
		ActionTitle:         title,
		ExecutionFinished:   true,
		Username:            "guest",
	}
	switch status {
	case "Blocked":
		entry.Blocked = true
	case "Completed":
		entry.ExitCode = 0
	}
	return entry
}
