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

func setupHostEntityTestData(t *testing.T) {
	t.Helper()
	entities.ClearEntitiesOfType("host")
	entities.AddEntity("host", "0", map[string]any{"name": "stuffbox", "hostname": "192.168.66.8"})
	entities.AddEntity("host", "1", map[string]any{"name": "lurker", "hostname": "192.168.66.1"})
	t.Cleanup(func() {
		entities.ClearEntitiesOfType("host")
	})
}

func buildRelatedActionsTestConfig(t *testing.T) (*config.Config, *authpublic.AuthenticatedUser, *authpublic.AuthenticatedUser) {
	t.Helper()

	cfg := config.DefaultConfig()
	cfg.DefaultPermissions.View = false
	cfg.DefaultPermissions.Exec = false

	cfg.Actions = append(cfg.Actions,
		&config.Action{
			Title:  "Secret Entity Action",
			Shell:  "echo secret",
			Entity: "host",
		},
		&config.Action{
			Title:  "Hidden Host Action",
			Shell:  "echo hidden",
			Entity: "host",
			Hidden: true,
		},
		&config.Action{
			ID:    "run_playbook",
			Title: "Run Automation Playbook",
			Shell: "host '{{ ansible_host }}'",
			Arguments: []config.ActionArgument{
				{
					Name:   "ansible_host",
					Title:  "Host",
					Entity: "host",
					Choices: []config.ActionArgumentChoice{
						{Title: "{{ host.name }} ({{ host.hostname }})", Value: "{{ host.hostname }}"},
					},
				},
			},
		},
		&config.Action{
			Title:  "Public Host Action",
			Shell:  "echo public",
			Entity: "host",
		},
	)

	cfg.AccessControlLists = append(cfg.AccessControlLists,
		&config.AccessControlList{
			Name:             "restricted",
			MatchUsernames:   []string{"low"},
			AddToEveryAction: true,
			Permissions:      config.PermissionsList{View: false, Exec: false, Logs: false, Kill: false},
		},
		&config.AccessControlList{
			Name:             "full",
			MatchUsernames:   []string{"admin"},
			AddToEveryAction: true,
			Permissions:      config.PermissionsList{View: true, Exec: true, Logs: true, Kill: true},
		},
	)

	cfg.Entities = []*config.EntityFile{
		{File: "hosts.yaml", Name: "host", Icon: "ssh"},
	}
	cfg.Sanitize()

	lowUser := &authpublic.AuthenticatedUser{Username: "low", Acls: []string{"restricted"}}
	adminUser := &authpublic.AuthenticatedUser{Username: "admin", Acls: []string{"full"}}

	return cfg, lowUser, adminUser
}

func getEntityRelatedActionTitles(t *testing.T, api *oliveTinAPI, user *authpublic.AuthenticatedUser, entityType, entityKey string) []string {
	t.Helper()

	entity, ok := entities.GetEntityInstances(entityType)[entityKey]
	require.True(t, ok, "entity %s/%s must exist", entityType, entityKey)

	related := api.relatedActionsForEntity(user, entityType, entity)
	titles := make([]string, 0, len(related))
	for _, item := range related {
		if item.Action != nil {
			titles = append(titles, item.Action.Title)
		}
	}
	return titles
}

func getEntityRelatedBindingIDs(t *testing.T, api *oliveTinAPI, user *authpublic.AuthenticatedUser, entityType, entityKey string) []string {
	t.Helper()

	entity, ok := entities.GetEntityInstances(entityType)[entityKey]
	require.True(t, ok, "entity %s/%s must exist", entityType, entityKey)

	related := api.relatedActionsForEntity(user, entityType, entity)
	ids := make([]string, 0, len(related))
	for _, item := range related {
		if item.Action != nil {
			ids = append(ids, item.Action.BindingId)
		}
	}
	return ids
}

func TestGetEntityRelatedActionsDeniesRestrictedView(t *testing.T) {
	setupHostEntityTestData(t)
	cfg, lowUser, _ := buildRelatedActionsTestConfig(t)

	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()
	api := newServer(ex)

	ids := getEntityRelatedBindingIDs(t, api, lowUser, "host", "0")
	assert.Empty(t, ids)

	titles := getEntityRelatedActionTitles(t, api, lowUser, "host", "0")
	assert.Empty(t, titles)
}

func TestGetEntityRelatedActionsAllowsAdminView(t *testing.T) {
	setupHostEntityTestData(t)
	cfg, _, adminUser := buildRelatedActionsTestConfig(t)

	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()
	api := newServer(ex)

	titles := getEntityRelatedActionTitles(t, api, adminUser, "host", "0")
	assert.Contains(t, titles, "Run Automation Playbook")
	assert.Contains(t, titles, "Public Host Action")
	assert.Contains(t, titles, "Secret Entity Action")
}

func TestGetEntityRelatedActionsExcludesHiddenActions(t *testing.T) {
	setupHostEntityTestData(t)
	cfg, _, adminUser := buildRelatedActionsTestConfig(t)

	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()
	api := newServer(ex)

	titles := getEntityRelatedActionTitles(t, api, adminUser, "host", "0")
	assert.NotContains(t, titles, "Hidden Host Action")
}

func TestGetEntityRelatedActionsEntityBoundMatchesInstanceOnly(t *testing.T) {
	setupHostEntityTestData(t)
	cfg := config.DefaultConfig()
	cfg.Actions = append(cfg.Actions, &config.Action{
		Title:  "{{ host.name }} Wake",
		Shell:  "echo wake",
		Entity: "host",
	})

	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()
	api := newServer(ex)
	user := &authpublic.AuthenticatedUser{Username: "guest", Provider: "system"}

	titlesHost0 := getEntityRelatedActionTitles(t, api, user, "host", "0")
	assert.Len(t, titlesHost0, 1)
	assert.Equal(t, "stuffbox Wake", titlesHost0[0])

	titlesHost1 := getEntityRelatedActionTitles(t, api, user, "host", "1")
	assert.Len(t, titlesHost1, 1)
	assert.Equal(t, "lurker Wake", titlesHost1[0])
}

func TestGetEntityRelatedActionsPrefillsArgumentEntityValues(t *testing.T) {
	setupHostEntityTestData(t)
	cfg := config.DefaultConfig()
	cfg.Actions = append(cfg.Actions, &config.Action{
		ID:    "run_playbook",
		Title: "Run Automation Playbook",
		Shell: "host '{{ ansible_host }}'",
		Arguments: []config.ActionArgument{
			{
				Name:   "ansible_host",
				Title:  "Host",
				Entity: "host",
				Choices: []config.ActionArgumentChoice{
					{Title: "{{ host.name }} ({{ host.hostname }})", Value: "{{ host.hostname }}"},
				},
			},
		},
	})

	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()
	api := newServer(ex)
	user := &authpublic.AuthenticatedUser{Username: "guest", Provider: "system"}

	entity := entities.GetEntityInstances("host")["0"]
	related := api.relatedActionsForEntity(user, "host", entity)
	require.Len(t, related, 1)
	require.NotNil(t, related[0].Action)
	assert.Equal(t, "run_playbook", related[0].Action.BindingId)
	assert.Equal(t, "192.168.66.8", related[0].PrefilledArguments["ansible_host"])
}

func TestGetEntityDeniesGuestsWhenLoginRequired(t *testing.T) {
	setupHostEntityTestData(t)
	cfg := config.DefaultConfig()
	cfg.AuthRequireGuestsToLogin = true

	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()
	ts, client := getNewTestServerAndClientWithExecutor(cfg, ex)
	defer ts.Close()

	_, err := client.GetEntity(context.Background(), connect.NewRequest(&apiv1.GetEntityRequest{
		Type:      "host",
		UniqueKey: "0",
	}))
	require.Error(t, err)
	assert.Equal(t, connect.CodePermissionDenied, connect.CodeOf(err))
}

func TestGetEntityReturnsRelatedActionsForAdmin(t *testing.T) {
	setupHostEntityTestData(t)
	cfg, _, adminUser := buildRelatedActionsTestConfig(t)
	cfg.AuthHttpHeaderUsername = "X-Ot-User"

	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()
	ts, client := getNewTestServerAndClientWithExecutor(cfg, ex)
	defer ts.Close()

	req := connect.NewRequest(&apiv1.GetEntityRequest{
		Type:      "host",
		UniqueKey: "0",
	})
	req.Header().Set("X-Ot-User", adminUser.Username)

	resp, err := client.GetEntity(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp.Msg)
	assert.Equal(t, "&#128272;", resp.Msg.Icon)

	foundPlaybook := false
	for _, related := range resp.Msg.RelatedActions {
		if related.Action != nil && related.Action.BindingId == "run_playbook" {
			foundPlaybook = true
			assert.Equal(t, "192.168.66.8", related.PrefilledArguments["ansible_host"])
		}
	}
	assert.True(t, foundPlaybook, "admin should see argument-entity related action in GetEntity response")
}
