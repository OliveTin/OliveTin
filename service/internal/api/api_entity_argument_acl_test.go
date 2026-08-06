package api

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiv1 "github.com/OliveTin/OliveTin/gen/olivetin/api/v1"
	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/entities"
	"github.com/OliveTin/OliveTin/internal/executor"
)

func buildEntityArgumentGuessConfig() *config.Config {
	cfg := config.DefaultConfig()
	cfg.DefaultPermissions.View = false
	cfg.DefaultPermissions.Exec = false
	cfg.AccessControlLists = []*config.AccessControlList{
		{
			Name:           "ops",
			MatchUsernames: []string{"admin"},
			Permissions:    config.PermissionsList{View: true, Exec: true},
		},
		{
			Name:             "everyone",
			MatchUsernames:   []string{"guest", "admin"},
			Permissions:      config.PermissionsList{View: true, Exec: true},
			AddToEveryAction: true,
		},
	}
	cfg.Entities = []*config.EntityFile{
		{Name: "servers", File: "servers.yaml", Acls: []string{"ops"}},
	}
	cfg.Actions = []*config.Action{
		{
			ID:    "reboot-server",
			Title: "Reboot server",
			Shell: "echo reboot '{{ target }}'",
			Arguments: []config.ActionArgument{
				{
					Name:   "target",
					Title:  "Server",
					Entity: "servers",
					Choices: []config.ActionArgumentChoice{
						{Title: "{{ servers.name }}", Value: "{{ servers.name }}"},
					},
				},
			},
		},
	}
	cfg.Sanitize()
	return cfg
}

func seedEntityArgumentGuessEntities(t *testing.T) {
	t.Helper()
	entities.ClearEntitiesOfType("servers")
	t.Cleanup(func() {
		entities.ClearEntitiesOfType("servers")
	})
	entities.AddEntity("servers", "0", map[string]any{"name": "web01"})
	entities.AddEntity("servers", "1", map[string]any{"name": "db01"})
}

func TestStartActionRejectsGuessedRestrictedEntityArgument(t *testing.T) {
	seedEntityArgumentGuessEntities(t)
	cfg := buildEntityArgumentGuessConfig()
	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()
	ts, client := getNewTestServerAndClientWithExecutor(cfg, ex)
	defer ts.Close()

	_, err := client.StartAction(context.Background(), connect.NewRequest(&apiv1.StartActionRequest{
		BindingId: "reboot-server",
		Arguments: []*apiv1.StartActionArgument{
			{Name: "target", Value: "web01"},
		},
	}))
	require.Error(t, err)
	assert.Equal(t, connect.CodePermissionDenied, connect.CodeOf(err))
}

func TestStartActionRejectsUnknownEntityArgumentValue(t *testing.T) {
	seedEntityArgumentGuessEntities(t)
	cfg := buildEntityArgumentGuessConfig()
	// Make servers unrestricted so guests can view the type but not invent values.
	cfg.Entities[0].Acls = nil
	cfg.Sanitize()

	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()
	ts, client := getNewTestServerAndClientWithExecutor(cfg, ex)
	defer ts.Close()

	_, err := client.StartAction(context.Background(), connect.NewRequest(&apiv1.StartActionRequest{
		BindingId: "reboot-server",
		Arguments: []*apiv1.StartActionArgument{
			{Name: "target", Value: "not-a-real-server"},
		},
	}))
	require.Error(t, err)
	assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
}

func TestStartActionAllowsListedEntityArgumentValue(t *testing.T) {
	seedEntityArgumentGuessEntities(t)
	cfg := buildEntityArgumentGuessConfig()
	cfg.Entities[0].Acls = nil
	cfg.Sanitize()

	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()
	ts, client := getNewTestServerAndClientWithExecutor(cfg, ex)
	defer ts.Close()

	resp, err := client.StartAction(context.Background(), connect.NewRequest(&apiv1.StartActionRequest{
		BindingId: "reboot-server",
		Arguments: []*apiv1.StartActionArgument{
			{Name: "target", Value: "web01"},
		},
	}))
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Msg.ExecutionTrackingId)
}

func TestChecklistEntityValuesAllowedRejectsBlankOnlyInput(t *testing.T) {
	allowed := map[string]struct{}{"web01": {}, "db01": {}}

	assert.False(t, checklistEntityValuesAllowed(",,,", allowed))
	assert.False(t, checklistEntityValuesAllowed(" , ", allowed))
	assert.False(t, checklistEntityValuesAllowed("", allowed),
		"all-blank checklist parts are rejected here; empty string is accepted by the caller separately")
	assert.True(t, checklistEntityValuesAllowed("web01", allowed))
	assert.True(t, checklistEntityValuesAllowed("web01, db01", allowed))
	assert.False(t, checklistEntityValuesAllowed("web01, unknown", allowed))
}

func TestStartActionRejectsMalformedMultiChoiceEntityArgument(t *testing.T) {
	seedEntityArgumentGuessEntities(t)
	cfg := buildEntityArgumentGuessConfig()
	// After sanitize, force an invalid entity+multi-choice shape that would
	// previously skip ACL and fall through to static UI choices.
	cfg.Actions[0].Arguments[0].Choices = []config.ActionArgumentChoice{
		{Title: "{{ servers.name }}", Value: "{{ servers.name }}"},
		{Title: "web01", Value: "web01"},
	}

	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()
	ts, client := getNewTestServerAndClientWithExecutor(cfg, ex)
	defer ts.Close()

	_, err := client.StartAction(context.Background(), connect.NewRequest(&apiv1.StartActionRequest{
		BindingId: "reboot-server",
		Arguments: []*apiv1.StartActionArgument{
			{Name: "target", Value: "web01"},
		},
	}))
	require.Error(t, err)
	assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
}
