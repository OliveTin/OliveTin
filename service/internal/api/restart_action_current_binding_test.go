package api

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiv1 "github.com/OliveTin/OliveTin/gen/olivetin/api/v1"
	apiv1connect "github.com/OliveTin/OliveTin/gen/olivetin/api/v1/apiv1connect"
	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/entities"
	"github.com/OliveTin/OliveTin/internal/executor"
)

const restartTestActionID = "restart-test-action"

func restartTestAction(command string, acls ...string) *config.Action {
	return &config.Action{
		ID:            restartTestActionID,
		Title:         "Restart test action",
		Exec:          []string{"echo", command},
		Acls:          acls,
		MaxConcurrent: 1,
		Timeout:       5,
	}
}

func startAndWaitForBinding(t *testing.T, client apiv1connect.OliveTinApiServiceClient, ex *executor.Executor, bindingID string) string {
	t.Helper()

	resp, err := client.StartAction(context.Background(), connect.NewRequest(&apiv1.StartActionRequest{BindingId: bindingID}))
	require.NoError(t, err)
	waitForLogFinished(t, ex, resp.Msg.ExecutionTrackingId)

	return resp.Msg.ExecutionTrackingId
}

func restartAndWait(t *testing.T, client apiv1connect.OliveTinApiServiceClient, ex *executor.Executor, trackingID string) *executor.InternalLogEntry {
	t.Helper()

	resp, err := client.RestartAction(context.Background(), connect.NewRequest(&apiv1.RestartActionRequest{ExecutionTrackingId: trackingID}))
	require.NoError(t, err)
	waitForLogFinished(t, ex, resp.Msg.ExecutionTrackingId)

	entry, ok := ex.GetLog(resp.Msg.ExecutionTrackingId)
	require.True(t, ok)

	return entry
}

func requireRestartNotFound(t *testing.T, client apiv1connect.OliveTinApiServiceClient, trackingID string) {
	t.Helper()

	_, err := client.RestartAction(context.Background(), connect.NewRequest(&apiv1.RestartActionRequest{ExecutionTrackingId: trackingID}))
	require.Error(t, err)
	assert.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
}

func logCount(ex *executor.Executor) int {
	logs, _ := ex.GetLogTrackingIds(0, 1000)

	return len(logs)
}

func entityBindingID(ex *executor.Executor) string {
	ex.MapActionBindingsLock.RLock()
	defer ex.MapActionBindingsLock.RUnlock()

	for id, binding := range ex.MapActionBindings {
		if binding.Entity != nil {
			return id
		}
	}

	return ""
}

func newRestartTestServer(t *testing.T, cfg *config.Config) (*executor.Executor, apiv1connect.OliveTinApiServiceClient) {
	t.Helper()

	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()

	ts, client := getNewTestServerAndClientWithExecutor(cfg, ex)
	t.Cleanup(ts.Close)

	return ex, client
}

func TestRestartActionRunsUnchangedActionAgain(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Actions = []*config.Action{restartTestAction("first-run")}
	ex, client := newRestartTestServer(t, cfg)

	trackingID := startAndWaitForBinding(t, client, ex, restartTestActionID)
	ex.RebuildActionMap()

	entry := restartAndWait(t, client, ex, trackingID)
	assert.False(t, entry.Blocked)
	assert.Contains(t, entry.Output, "first-run")
}

func TestRestartActionRejectsActionRemovedFromConfig(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Actions = []*config.Action{restartTestAction("removed-command")}
	ex, client := newRestartTestServer(t, cfg)

	trackingID := startAndWaitForBinding(t, client, ex, restartTestActionID)
	logsBefore := logCount(ex)

	cfg.Actions = nil
	ex.RebuildActionMap()

	requireRestartNotFound(t, client, trackingID)
	assert.Equal(t, logsBefore, logCount(ex), "a rejected restart must not start an execution")
}

func TestRestartActionUsesCurrentAclsAfterReload(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.DefaultPermissions.Exec = false
	cfg.AccessControlLists = []*config.AccessControlList{
		{Name: "guest-exec", MatchUsernames: []string{"guest"}, Permissions: config.PermissionsList{Exec: true}},
		{Name: "admin-exec", MatchUsernames: []string{"admin"}, Permissions: config.PermissionsList{Exec: true}},
	}
	cfg.Actions = []*config.Action{restartTestAction("old-guest-command", "guest-exec")}
	ex, client := newRestartTestServer(t, cfg)

	trackingID := startAndWaitForBinding(t, client, ex, restartTestActionID)

	cfg.Actions = []*config.Action{restartTestAction("new-admin-command", "admin-exec")}
	ex.RebuildActionMap()

	entry := restartAndWait(t, client, ex, trackingID)
	assert.True(t, entry.Blocked, "the guest must be blocked by the action's current ACLs")
	assert.NotContains(t, entry.Output, "old-guest-command")
	assert.NotContains(t, entry.Output, "new-admin-command")
}

func TestRestartActionRunsCurrentCommandAfterReload(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Actions = []*config.Action{restartTestAction("old-command")}
	ex, client := newRestartTestServer(t, cfg)

	trackingID := startAndWaitForBinding(t, client, ex, restartTestActionID)

	cfg.Actions = []*config.Action{restartTestAction("new-command")}
	ex.RebuildActionMap()

	entry := restartAndWait(t, client, ex, trackingID)
	assert.False(t, entry.Blocked)
	assert.Contains(t, entry.Output, "new-command")
	assert.NotContains(t, entry.Output, "old-command")
}

func TestRestartActionRequiresJustificationAddedAfterReload(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Actions = []*config.Action{restartTestAction("no-justification")}
	ex, client := newRestartTestServer(t, cfg)

	trackingID := startAndWaitForBinding(t, client, ex, restartTestActionID)

	action := restartTestAction("no-justification")
	action.Justification = config.JustificationRequiredNoTemplate
	cfg.Actions = []*config.Action{action}
	ex.RebuildActionMap()

	_, err := client.RestartAction(context.Background(), connect.NewRequest(&apiv1.RestartActionRequest{ExecutionTrackingId: trackingID}))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "justification")
}

func TestRestartActionRequiresArgumentAddedAfterReload(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Actions = []*config.Action{restartTestAction("no-arguments")}
	ex, client := newRestartTestServer(t, cfg)

	trackingID := startAndWaitForBinding(t, client, ex, restartTestActionID)

	action := restartTestAction("{{ host }}")
	action.Arguments = []config.ActionArgument{{Name: "host", Type: "ascii_identifier"}}
	cfg.Actions = []*config.Action{action}
	ex.RebuildActionMap()

	_, err := client.RestartAction(context.Background(), connect.NewRequest(&apiv1.RestartActionRequest{ExecutionTrackingId: trackingID}))
	require.Error(t, err)
	assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
}

func TestRestartActionRejectsEntityRemovedAfterReload(t *testing.T) {
	const entityType = "restarttestservers"

	entities.ClearEntitiesOfType(entityType)
	t.Cleanup(func() { entities.ClearEntitiesOfType(entityType) })
	entities.AddEntity(entityType, "web01", map[string]any{"name": "web01"})

	cfg := config.DefaultConfig()
	cfg.Actions = []*config.Action{{
		Title:         "Restart entity test",
		Exec:          []string{"echo", "entity-command"},
		Entity:        entityType,
		MaxConcurrent: 1,
		Timeout:       5,
	}}
	ex, client := newRestartTestServer(t, cfg)

	bindingID := entityBindingID(ex)
	require.NotEmpty(t, bindingID)

	trackingID := startAndWaitForBinding(t, client, ex, bindingID)

	entities.ClearEntitiesOfType(entityType)
	ex.RebuildActionMap()

	requireRestartNotFound(t, client, trackingID)
}
