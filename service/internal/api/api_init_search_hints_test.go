package api

import (
	"context"
	"fmt"
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

func TestInitIncludesEntitySearchHints(t *testing.T) {
	entities.ClearEntitiesOfType("server")
	entities.ClearEntitiesOfType("database")
	t.Cleanup(func() {
		entities.ClearEntitiesOfType("server")
		entities.ClearEntitiesOfType("database")
	})

	entities.AddEntity("server", "0", map[string]any{
		"name":   "web01",
		"secret": "must-not-appear-in-search-hints",
	})
	entities.AddEntity("database", "db-1", map[string]any{
		"title": "postgres",
	})

	cfg := config.DefaultConfig()
	cfg.Features.HeaderSearch = true
	cfg.Sanitize()

	testExecutor := executor.DefaultExecutor(cfg)
	testExecutor.RebuildActionMap()
	testServer, client := getNewTestServerAndClientWithExecutor(cfg, testExecutor)
	defer testServer.Close()

	resp, err := client.Init(context.Background(), connect.NewRequest(&apiv1.InitRequest{}))
	require.NoError(t, err)
	require.NotNil(t, resp.Msg.Features)
	assert.True(t, resp.Msg.Features.HeaderSearch)
	require.NotNil(t, resp.Msg.SearchHints)

	byKey := map[string]*apiv1.EntitySearchHint{}
	for _, hint := range resp.Msg.SearchHints.Entities {
		byKey[hint.Type+":"+hint.UniqueKey] = hint
	}

	host, ok := byKey["server:0"]
	require.True(t, ok, "expected server:0 search hint")
	assert.Equal(t, "web01", host.Title)
	assert.Equal(t, "server", host.Type)
	assert.Equal(t, "0", host.UniqueKey)

	db, ok := byKey["database:db-1"]
	require.True(t, ok, "expected database:db-1 search hint")
	assert.Equal(t, "postgres", db.Title)
}

func TestInitOmitsSearchHintsWhenLoginRequired(t *testing.T) {
	entities.ClearEntitiesOfType("server")
	t.Cleanup(func() {
		entities.ClearEntitiesOfType("server")
	})

	entities.AddEntity("server", "0", map[string]any{"name": "web01"})

	cfg := config.DefaultConfig()
	cfg.AuthRequireGuestsToLogin = true
	cfg.Features.HeaderSearch = true
	cfg.Sanitize()

	testExecutor := executor.DefaultExecutor(cfg)
	testExecutor.RebuildActionMap()
	testServer, client := getNewTestServerAndClientWithExecutor(cfg, testExecutor)
	defer testServer.Close()

	resp, err := client.Init(context.Background(), connect.NewRequest(&apiv1.InitRequest{}))
	require.NoError(t, err)
	require.True(t, resp.Msg.LoginRequired)
	assert.Nil(t, resp.Msg.SearchHints)
}

func TestInitOmitsSearchHintsWhenHeaderSearchDisabled(t *testing.T) {
	entities.ClearEntitiesOfType("server")
	t.Cleanup(func() {
		entities.ClearEntitiesOfType("server")
	})

	entities.AddEntity("server", "0", map[string]any{"name": "web01"})

	cfg := config.DefaultConfig()
	cfg.Sanitize()
	require.False(t, cfg.Features.HeaderSearch)

	testExecutor := executor.DefaultExecutor(cfg)
	testExecutor.RebuildActionMap()
	testServer, client := getNewTestServerAndClientWithExecutor(cfg, testExecutor)
	defer testServer.Close()

	resp, err := client.Init(context.Background(), connect.NewRequest(&apiv1.InitRequest{}))
	require.NoError(t, err)
	require.NotNil(t, resp.Msg.Features)
	assert.False(t, resp.Msg.Features.HeaderSearch)
	assert.Nil(t, resp.Msg.SearchHints)
}

func TestBuildSearchHintsRespectsActionACL(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.DefaultPermissions.View = false
	cfg.DefaultPermissions.Exec = false
	cfg.Actions = []*config.Action{
		{ID: "public_action", Title: "Public Action", Shell: "echo public"},
		{ID: "secret_action", Title: "Secret Action", Shell: "echo secret", Acls: []string{"admins"}},
	}
	cfg.AccessControlLists = []*config.AccessControlList{
		{
			Name:             "everyone",
			MatchUsernames:   []string{"guest", "admin"},
			AddToEveryAction: false,
			Permissions:      config.PermissionsList{View: true, Exec: true},
		},
		{
			Name:           "admins",
			MatchUsernames: []string{"admin"},
			Permissions:    config.PermissionsList{View: true, Exec: true},
		},
	}
	cfg.Actions[0].Acls = []string{"everyone"}
	cfg.Sanitize()

	testExecutor := executor.DefaultExecutor(cfg)
	testExecutor.RebuildActionMap()
	api := newServer(testExecutor)

	guest := &authpublic.AuthenticatedUser{Username: "guest", Provider: "system"}
	guest.BuildUserAcls(cfg)
	admin := &authpublic.AuthenticatedUser{Username: "admin"}
	admin.BuildUserAcls(cfg)

	guestHints := api.buildSearchHints(guest)
	require.NotNil(t, guestHints)

	guestActionIDs := actionHintBindingIDs(guestHints.Actions)
	assert.Contains(t, guestActionIDs, "public_action")
	assert.NotContains(t, guestActionIDs, "secret_action")

	adminHints := api.buildSearchHints(admin)
	require.NotNil(t, adminHints)

	adminActionIDs := actionHintBindingIDs(adminHints.Actions)
	assert.Contains(t, adminActionIDs, "public_action")
	assert.Contains(t, adminActionIDs, "secret_action")
}

func TestBuildSearchHintsOmitsHiddenActions(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Actions = []*config.Action{
		{ID: "visible", Title: "Visible", Shell: "echo visible"},
		{ID: "hidden", Title: "Hidden", Shell: "echo hidden", Hidden: true},
	}
	cfg.Sanitize()

	testExecutor := executor.DefaultExecutor(cfg)
	testExecutor.RebuildActionMap()
	api := newServer(testExecutor)

	user := &authpublic.AuthenticatedUser{Username: "guest", Provider: "system"}
	user.BuildUserAcls(cfg)

	hints := api.buildSearchHints(user)
	require.NotNil(t, hints)

	ids := actionHintBindingIDs(hints.Actions)
	assert.Contains(t, ids, "visible")
	assert.NotContains(t, ids, "hidden")
}

func TestBuildSearchHintsCapsActionsAndEntitiesPerType(t *testing.T) {
	entities.ClearEntitiesOfType("cap_host")
	entities.ClearEntitiesOfType("cap_container")
	t.Cleanup(func() {
		entities.ClearEntitiesOfType("cap_host")
		entities.ClearEntitiesOfType("cap_container")
	})

	for i := 0; i < maxSearchHintEntitiesPerType+5; i++ {
		entities.AddEntity("cap_host", fmt.Sprintf("%03d", i), map[string]any{"name": fmt.Sprintf("host-%03d", i)})
		entities.AddEntity("cap_container", fmt.Sprintf("%03d", i), map[string]any{"name": fmt.Sprintf("ctr-%03d", i)})
	}

	cfg := config.DefaultConfig()
	cfg.Entities = []*config.EntityFile{
		{Name: "cap_host", File: "cap_host.yaml"},
		{Name: "cap_container", File: "cap_container.yaml"},
	}
	cfg.Actions = make([]*config.Action, 0, maxSearchHintActions+5)
	for i := 0; i < maxSearchHintActions+5; i++ {
		cfg.Actions = append(cfg.Actions, &config.Action{
			ID:    fmt.Sprintf("action-%03d", i),
			Title: fmt.Sprintf("Action %03d", i),
			Shell: "echo",
		})
	}
	cfg.Sanitize()

	testExecutor := executor.DefaultExecutor(cfg)
	testExecutor.RebuildActionMap()
	api := newServer(testExecutor)

	user := &authpublic.AuthenticatedUser{Username: "guest", Provider: "system"}
	user.BuildUserAcls(cfg)

	hints := api.buildSearchHints(user)
	require.NotNil(t, hints)
	assert.Len(t, hints.Actions, maxSearchHintActions)
	assert.Equal(t, maxSearchHintEntitiesPerType, countEntityHintsByType(hints.Entities, "cap_host"))
	assert.Equal(t, maxSearchHintEntitiesPerType, countEntityHintsByType(hints.Entities, "cap_container"))
}

func countEntityHintsByType(hints []*apiv1.EntitySearchHint, entityType string) int {
	count := 0
	for _, hint := range hints {
		if hint.Type == entityType {
			count++
		}
	}
	return count
}

func TestBuildSearchHintsPrefersNonEntityActions(t *testing.T) {
	entities.ClearEntitiesOfType("host")
	t.Cleanup(func() {
		entities.ClearEntitiesOfType("host")
	})

	for i := 0; i < 10; i++ {
		entities.AddEntity("host", fmt.Sprintf("%d", i), map[string]any{"name": fmt.Sprintf("host-%d", i)})
	}

	cfg := config.DefaultConfig()
	cfg.Actions = []*config.Action{
		{ID: "plain", Title: "Plain Action", Shell: "echo plain"},
		{Title: "Entity Action {{ name }}", Shell: "echo entity", Entity: "host"},
	}
	cfg.Sanitize()

	testExecutor := executor.DefaultExecutor(cfg)
	testExecutor.RebuildActionMap()
	api := newServer(testExecutor)

	user := &authpublic.AuthenticatedUser{Username: "guest", Provider: "system"}
	user.BuildUserAcls(cfg)

	hints := api.buildSearchHints(user)
	require.NotNil(t, hints)
	require.NotEmpty(t, hints.Actions)
	assert.Equal(t, "plain", hints.Actions[0].BindingId)
}

func actionHintBindingIDs(hints []*apiv1.ActionSearchHint) []string {
	ids := make([]string, 0, len(hints))
	for _, hint := range hints {
		ids = append(ids, hint.BindingId)
	}
	return ids
}
