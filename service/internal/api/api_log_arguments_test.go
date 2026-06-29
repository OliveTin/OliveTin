package api

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiv1 "github.com/OliveTin/OliveTin/gen/olivetin/api/v1"
	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/executor"
)

func argumentAction(title, shell string, args []config.ActionArgument) *config.Action {
	return &config.Action{
		Title:         title,
		Shell:         shell,
		MaxConcurrent: 1,
		Arguments:     args,
	}
}

func waitForLogArguments(t *testing.T, ex *executor.Executor, trackingID string) map[string]string {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		entry, ok := ex.GetLog(trackingID)
		if ok && entry.Arguments != nil {
			return entry.Arguments
		}

		time.Sleep(5 * time.Millisecond)
	}

	t.Fatalf("timed out waiting for arguments on log %s", trackingID)
	return nil
}

func waitForLogJustification(t *testing.T, ex *executor.Executor, trackingID, expected string) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		entry, ok := ex.GetLog(trackingID)
		if ok && entry.Justification == expected {
			return
		}

		time.Sleep(5 * time.Millisecond)
	}

	t.Fatalf("timed out waiting for justification %q on log %s", expected, trackingID)
}

func TestExecutionStatusIncludesStoredArguments(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Actions = []*config.Action{
		argumentAction("Ping host", "echo {{ host }}", []config.ActionArgument{
			{Name: "host", Type: "ascii_identifier"},
		}),
	}

	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()
	binding := ex.FindBindingWithNoEntity(cfg.Actions[0])
	require.NotNil(t, binding)

	ts, client := getNewTestServerAndClientWithExecutor(cfg, ex)
	defer ts.Close()

	startResp, err := client.StartAction(context.Background(), connect.NewRequest(&apiv1.StartActionRequest{
		BindingId: binding.ID,
		Arguments: []*apiv1.StartActionArgument{
			{Name: "host", Value: "example.com"},
		},
	}))
	require.NoError(t, err)

	waitForLogArguments(t, ex, startResp.Msg.ExecutionTrackingId)

	statusResp, err := client.ExecutionStatus(context.Background(), connect.NewRequest(&apiv1.ExecutionStatusRequest{
		ExecutionTrackingId: startResp.Msg.ExecutionTrackingId,
	}))
	require.NoError(t, err)
	require.NotNil(t, statusResp.Msg.LogEntry)
	require.Len(t, statusResp.Msg.LogEntry.Arguments, 1)
	assert.Equal(t, "host", statusResp.Msg.LogEntry.Arguments[0].Name)
	assert.Equal(t, "example.com", statusResp.Msg.LogEntry.Arguments[0].Value)
}

func TestExecutionStatusOmitsPasswordArguments(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Actions = []*config.Action{
		{
			Title:         "Connect",
			Exec:          []string{"echo", "{{ user }}"},
			MaxConcurrent: 1,
			Arguments: []config.ActionArgument{
				{Name: "user", Type: "ascii_identifier"},
				{Name: "pass", Type: "password"},
			},
		},
	}

	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()
	binding := ex.FindBindingWithNoEntity(cfg.Actions[0])
	require.NotNil(t, binding)

	ts, client := getNewTestServerAndClientWithExecutor(cfg, ex)
	defer ts.Close()

	startResp, err := client.StartAction(context.Background(), connect.NewRequest(&apiv1.StartActionRequest{
		BindingId: binding.ID,
		Arguments: []*apiv1.StartActionArgument{
			{Name: "user", Value: "alice"},
			{Name: "pass", Value: "secret"},
		},
	}))
	require.NoError(t, err)

	waitForLogArguments(t, ex, startResp.Msg.ExecutionTrackingId)

	statusResp, err := client.ExecutionStatus(context.Background(), connect.NewRequest(&apiv1.ExecutionStatusRequest{
		ExecutionTrackingId: startResp.Msg.ExecutionTrackingId,
	}))
	require.NoError(t, err)
	require.NotNil(t, statusResp.Msg.LogEntry)

	for _, arg := range statusResp.Msg.LogEntry.Arguments {
		assert.NotEqual(t, "pass", arg.Name)
	}

	require.Len(t, statusResp.Msg.LogEntry.Arguments, 1)
	assert.Equal(t, "user", statusResp.Msg.LogEntry.Arguments[0].Name)
	assert.Equal(t, "alice", statusResp.Msg.LogEntry.Arguments[0].Value)
}

func TestRestartActionReusesStoredArguments(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Actions = []*config.Action{
		argumentAction("Ping host", "echo {{ host }}", []config.ActionArgument{
			{Name: "host", Type: "ascii_identifier"},
		}),
	}

	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()
	binding := ex.FindBindingWithNoEntity(cfg.Actions[0])
	require.NotNil(t, binding)

	ts, client := getNewTestServerAndClientWithExecutor(cfg, ex)
	defer ts.Close()

	startResp, err := client.StartAction(context.Background(), connect.NewRequest(&apiv1.StartActionRequest{
		BindingId: binding.ID,
		Arguments: []*apiv1.StartActionArgument{
			{Name: "host", Value: "server-a"},
		},
	}))
	require.NoError(t, err)

	originalArgs := waitForLogArguments(t, ex, startResp.Msg.ExecutionTrackingId)
	assert.Equal(t, "server-a", originalArgs["host"])

	restartResp, err := client.RestartAction(context.Background(), connect.NewRequest(&apiv1.RestartActionRequest{
		ExecutionTrackingId: startResp.Msg.ExecutionTrackingId,
	}))
	require.NoError(t, err)
	require.NotEmpty(t, restartResp.Msg.ExecutionTrackingId)
	assert.NotEqual(t, startResp.Msg.ExecutionTrackingId, restartResp.Msg.ExecutionTrackingId)

	restartedArgs := waitForLogArguments(t, ex, restartResp.Msg.ExecutionTrackingId)
	assert.Equal(t, "server-a", restartedArgs["host"])
}

func TestRestartActionRejectsIncompleteStoredArguments(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Actions = []*config.Action{
		{
			Title:         "Connect",
			Exec:          []string{"echo", "{{ user }}"},
			MaxConcurrent: 1,
			Arguments: []config.ActionArgument{
				{Name: "user", Type: "ascii_identifier"},
				{Name: "pass", Type: "password"},
			},
		},
	}

	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()
	binding := ex.FindBindingWithNoEntity(cfg.Actions[0])
	require.NotNil(t, binding)

	ts, client := getNewTestServerAndClientWithExecutor(cfg, ex)
	defer ts.Close()

	startResp, err := client.StartAction(context.Background(), connect.NewRequest(&apiv1.StartActionRequest{
		BindingId: binding.ID,
		Arguments: []*apiv1.StartActionArgument{
			{Name: "user", Value: "alice"},
			{Name: "pass", Value: "secret"},
		},
	}))
	require.NoError(t, err)

	_, err = client.RestartAction(context.Background(), connect.NewRequest(&apiv1.RestartActionRequest{
		ExecutionTrackingId: startResp.Msg.ExecutionTrackingId,
	}))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "stored arguments are incomplete for restart")
}

func TestRestartActionRejectsMissingRequiredStoredArguments(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Actions = []*config.Action{
		argumentAction("Ping host", "echo {{ host }}", []config.ActionArgument{
			{Name: "host", Type: "ascii_identifier"},
		}),
	}

	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()
	binding := ex.FindBindingWithNoEntity(cfg.Actions[0])
	require.NotNil(t, binding)

	trackingID := "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
	ex.SetLog(trackingID, &executor.InternalLogEntry{
		Binding:             binding,
		ExecutionFinished:   true,
		ExecutionTrackingID: trackingID,
		Arguments:           map[string]string{},
	})

	ts, client := getNewTestServerAndClientWithExecutor(cfg, ex)
	defer ts.Close()

	_, err := client.RestartAction(context.Background(), connect.NewRequest(&apiv1.RestartActionRequest{
		ExecutionTrackingId: trackingID,
	}))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "stored arguments are incomplete for restart")
}

func TestLogEntryArgumentsToProto(t *testing.T) {
	assert.Nil(t, logEntryArgumentsToProto(nil))
	assert.Nil(t, logEntryArgumentsToProto(map[string]string{}))

	out := logEntryArgumentsToProto(map[string]string{
		"host": "example.com",
		"port": "443",
	})
	require.Len(t, out, 2)

	values := map[string]string{}
	for _, arg := range out {
		values[arg.Name] = arg.Value
	}

	assert.Equal(t, "example.com", values["host"])
	assert.Equal(t, "443", values["port"])
}

func TestCopyStringMap(t *testing.T) {
	source := map[string]string{"host": "example.com"}
	copied := copyStringMap(source)

	assert.Equal(t, source, copied)
	source["host"] = "changed"
	assert.Equal(t, "example.com", copied["host"])

	empty := copyStringMap(nil)
	assert.NotNil(t, empty)
	assert.Empty(t, empty)
}

func TestRestartActionRequiresJustificationWhenMissingFromStoredLog(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Actions = []*config.Action{
		{
			Title:         "Dangerous action",
			Shell:         "echo ok",
			MaxConcurrent: 1,
			Justification: true,
		},
	}

	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()
	binding := ex.FindBindingWithNoEntity(cfg.Actions[0])
	require.NotNil(t, binding)

	trackingID := "manual-log-without-justification"
	ex.SetLog(trackingID, &executor.InternalLogEntry{
		Binding:             binding,
		ExecutionFinished:   true,
		ExecutionTrackingID: trackingID,
	})

	ts, client := getNewTestServerAndClientWithExecutor(cfg, ex)
	defer ts.Close()

	_, err := client.RestartAction(context.Background(), connect.NewRequest(&apiv1.RestartActionRequest{
		ExecutionTrackingId: trackingID,
	}))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "justification")
}

func TestRestartActionReusesStoredJustificationViaStartActionPath(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Actions = []*config.Action{
		{
			Title:         "Dangerous action",
			Shell:         "echo ok",
			MaxConcurrent: 1,
			Justification: true,
		},
	}

	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()
	binding := ex.FindBindingWithNoEntity(cfg.Actions[0])
	require.NotNil(t, binding)

	ts, client := getNewTestServerAndClientWithExecutor(cfg, ex)
	defer ts.Close()

	startResp, err := client.StartAction(context.Background(), connect.NewRequest(&apiv1.StartActionRequest{
		BindingId:     binding.ID,
		Justification: "maintenance window",
	}))
	require.NoError(t, err)

	waitForLogJustification(t, ex, startResp.Msg.ExecutionTrackingId, "maintenance window")

	restartResp, err := client.RestartAction(context.Background(), connect.NewRequest(&apiv1.RestartActionRequest{
		ExecutionTrackingId: startResp.Msg.ExecutionTrackingId,
	}))
	require.NoError(t, err)

	waitForLogJustification(t, ex, restartResp.Msg.ExecutionTrackingId, "maintenance window")
}

func TestGetLogsIncludesStoredArguments(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Actions = []*config.Action{
		argumentAction("Ping host", "echo {{ host }}", []config.ActionArgument{
			{Name: "host", Type: "ascii_identifier"},
		}),
	}

	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()
	binding := ex.FindBindingWithNoEntity(cfg.Actions[0])
	require.NotNil(t, binding)

	ts, client := getNewTestServerAndClientWithExecutor(cfg, ex)
	defer ts.Close()

	startResp, err := client.StartAction(context.Background(), connect.NewRequest(&apiv1.StartActionRequest{
		BindingId: binding.ID,
		Arguments: []*apiv1.StartActionArgument{
			{Name: "host", Value: "db-1"},
		},
	}))
	require.NoError(t, err)
	require.NotEmpty(t, startResp.Msg.ExecutionTrackingId)

	waitForLogArguments(t, ex, startResp.Msg.ExecutionTrackingId)

	logsResp, err := client.GetLogs(context.Background(), connect.NewRequest(&apiv1.GetLogsRequest{}))
	require.NoError(t, err)
	require.NotEmpty(t, logsResp.Msg.Logs)

	var matched bool
	for _, entry := range logsResp.Msg.Logs {
		if entry.ExecutionTrackingId != startResp.Msg.ExecutionTrackingId {
			continue
		}

		matched = true
		require.Len(t, entry.Arguments, 1)
		assert.Equal(t, "host", entry.Arguments[0].Name)
		assert.Equal(t, "db-1", entry.Arguments[0].Value)
	}

	assert.True(t, matched, "expected log entry with stored arguments")
}
