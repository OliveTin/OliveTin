package api

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiv1 "github.com/OliveTin/OliveTin/gen/olivetin/api/v1"
	authpublic "github.com/OliveTin/OliveTin/internal/auth/authpublic"
	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/entities"
	"github.com/OliveTin/OliveTin/internal/executor"
)

func buildEntityAclTestConfig() *config.Config {
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
		{Name: "printers", File: "printers.yaml"},
		{Name: "servers", File: "servers.yaml", Acls: []string{"ops"}},
	}
	cfg.Actions = []*config.Action{
		{
			ID:     "restart-server",
			Title:  "Restart {{ servers.name }}",
			Shell:  "echo restart",
			Entity: "servers",
		},
		{
			ID:    "ping-printer",
			Title: "Ping printer",
			Shell: "echo ping",
		},
	}
	cfg.Dashboards = []*config.DashboardComponent{
		{
			Title: "Infra",
			Contents: []*config.DashboardComponent{
				{
					Title:  "{{ servers.name }}",
					Type:   "fieldset",
					Entity: "servers",
					Contents: []*config.DashboardComponent{
						{Title: "Restart {{ servers.name }}"},
					},
				},
			},
		},
	}
	cfg.Sanitize()
	return cfg
}

func seedEntityAclTestEntities(t *testing.T) {
	t.Helper()
	entities.ClearEntitiesOfType("printers")
	entities.ClearEntitiesOfType("servers")
	t.Cleanup(func() {
		entities.ClearEntitiesOfType("printers")
		entities.ClearEntitiesOfType("servers")
	})

	entities.AddEntity("printers", "p1", map[string]any{"name": "lobby"})
	entities.AddEntity("servers", "0", map[string]any{"name": "web01"})
}

func TestGetEntitiesOmitsRestrictedEntityTypes(t *testing.T) {
	seedEntityAclTestEntities(t)
	cfg := buildEntityAclTestConfig()
	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()
	api := newServer(ex)

	guest := &authpublic.AuthenticatedUser{Username: "guest", Provider: "system"}
	guest.BuildUserAcls(cfg)

	defs := api.buildEntityDefinitionsResponse(guest, &apiv1.GetEntitiesRequest{}, entities.GetEntities())
	titles := make([]string, 0, len(defs))
	for _, def := range defs {
		titles = append(titles, def.Title)
	}

	assert.Contains(t, titles, "printers")
	assert.NotContains(t, titles, "servers")
}

func TestGetEntityNotFoundForRestrictedType(t *testing.T) {
	seedEntityAclTestEntities(t)
	cfg := buildEntityAclTestConfig()
	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()
	ts, client := getNewTestServerAndClientWithExecutor(cfg, ex)
	defer ts.Close()

	// Guest Init/API without login uses guest user with default ACLs from UserFromApiCall.
	_, err := client.GetEntity(context.Background(), connect.NewRequest(&apiv1.GetEntityRequest{
		Type:      "servers",
		UniqueKey: "0",
	}))
	require.Error(t, err)
	assert.Equal(t, connect.CodeNotFound, connect.CodeOf(err))

	resp, err := client.GetEntity(context.Background(), connect.NewRequest(&apiv1.GetEntityRequest{
		Type:      "printers",
		UniqueKey: "p1",
	}))
	require.NoError(t, err)
	assert.Equal(t, "lobby", resp.Msg.Title)
}

func TestSearchHintsOmitRestrictedEntitiesAndEntityBoundActions(t *testing.T) {
	seedEntityAclTestEntities(t)
	cfg := buildEntityAclTestConfig()
	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()
	api := newServer(ex)

	guest := &authpublic.AuthenticatedUser{Username: "guest", Provider: "system"}
	guest.BuildUserAcls(cfg)
	admin := &authpublic.AuthenticatedUser{Username: "admin"}
	admin.BuildUserAcls(cfg)

	guestHints := api.buildSearchHints(guest)
	require.NotNil(t, guestHints)

	guestEntityKeys := make([]string, 0)
	for _, hint := range guestHints.Entities {
		guestEntityKeys = append(guestEntityKeys, hint.Type+":"+hint.UniqueKey)
	}
	assert.Contains(t, guestEntityKeys, "printers:p1")
	assert.NotContains(t, guestEntityKeys, "servers:0")

	guestActionIDs := actionHintBindingIDs(guestHints.Actions)
	for _, id := range guestActionIDs {
		assert.NotContains(t, id, "restart")
	}

	adminHints := api.buildSearchHints(admin)
	require.NotNil(t, adminHints)

	adminEntityKeys := make([]string, 0)
	for _, hint := range adminHints.Entities {
		adminEntityKeys = append(adminEntityKeys, hint.Type+":"+hint.UniqueKey)
	}
	assert.Contains(t, adminEntityKeys, "servers:0")
	assert.NotEmpty(t, adminHints.Actions)
}

func TestEntityFieldsetOmitsRestrictedEntityType(t *testing.T) {
	seedEntityAclTestEntities(t)
	cfg := buildEntityAclTestConfig()
	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()
	api := newServer(ex)

	guest := &authpublic.AuthenticatedUser{Username: "guest", Provider: "system"}
	guest.BuildUserAcls(cfg)
	admin := &authpublic.AuthenticatedUser{Username: "admin"}
	admin.BuildUserAcls(cfg)

	guestRR := api.createDashboardRenderRequest(guest, "", "")
	guestDB := renderDashboard(guestRR, "Infra")
	require.NotNil(t, guestDB)
	assert.Empty(t, guestDB.Contents, "restricted entity fieldsets must not leak instances")

	adminRR := api.createDashboardRenderRequest(admin, "", "")
	adminDB := renderDashboard(adminRR, "Infra")
	require.NotNil(t, adminDB)
	require.NotEmpty(t, adminDB.Contents)
}

func TestBuildChoicesEntityRespectsEntityACL(t *testing.T) {
	seedEntityAclTestEntities(t)
	cfg := buildEntityAclTestConfig()
	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()

	guest := &authpublic.AuthenticatedUser{Username: "guest", Provider: "system"}
	guest.BuildUserAcls(cfg)
	admin := &authpublic.AuthenticatedUser{Username: "admin"}
	admin.BuildUserAcls(cfg)

	arg := config.ActionArgument{
		Entity: "servers",
		Choices: []config.ActionArgumentChoice{
			{Title: "{{ servers.name }}", Value: "{{ servers.name }}"},
		},
	}

	guestRR := &DashboardRenderRequest{AuthenticatedUser: guest, cfg: cfg, ex: ex}
	assert.Empty(t, buildChoices(arg, guestRR))

	adminRR := &DashboardRenderRequest{AuthenticatedUser: admin, cfg: cfg, ex: ex}
	choices := buildChoices(arg, adminRR)
	require.Len(t, choices, 1)
	assert.Equal(t, "web01", choices[0].Value)
}
