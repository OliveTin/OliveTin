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

func TestGetEntitiesPaginatesAndFiltersInstances(t *testing.T) {
	entities.ClearEntitiesOfType("server")
	entities.AddEntity("server", "0", map[string]any{"name": "alpha", "hostname": "alpha.example.com", "ip": "10.0.0.1"})
	entities.AddEntity("server", "1", map[string]any{"name": "beta", "hostname": "beta.example.com", "ip": "10.0.0.2"})
	entities.AddEntity("server", "2", map[string]any{"name": "gamma", "hostname": "gamma.example.com", "ip": "10.0.0.3"})
	t.Cleanup(func() {
		entities.ClearEntitiesOfType("server")
	})

	cfg := config.DefaultConfig()
	cfg.Entities = []*config.EntityFile{
		{
			Name: "server",
			Properties: []config.EntityProperty{
				{Name: "hostname", Title: "Hostname"},
				{Name: "ip", Title: "IP"},
			},
		},
	}
	cfg.Sanitize()

	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()
	ts, client := getNewTestServerAndClientWithExecutor(cfg, ex)
	defer ts.Close()

	filteredResp, err := client.GetEntities(context.Background(), connect.NewRequest(&apiv1.GetEntitiesRequest{
		EntityType: "server",
		Filter:     "beta",
		Page:       1,
		PageSize:   10,
	}))
	require.NoError(t, err)
	require.Len(t, filteredResp.Msg.EntityDefinitions, 1)
	assert.Equal(t, int32(1), filteredResp.Msg.EntityDefinitions[0].TotalInstances)
	require.Len(t, filteredResp.Msg.EntityDefinitions[0].Instances, 1)
	assert.Equal(t, "beta.example.com", filteredResp.Msg.EntityDefinitions[0].Instances[0].Fields["hostname"])

	pagedResp, err := client.GetEntities(context.Background(), connect.NewRequest(&apiv1.GetEntitiesRequest{
		EntityType: "server",
		Page:       2,
		PageSize:   1,
	}))
	require.NoError(t, err)
	require.Len(t, pagedResp.Msg.EntityDefinitions, 1)
	assert.Equal(t, int32(3), pagedResp.Msg.EntityDefinitions[0].TotalInstances)
	require.Len(t, pagedResp.Msg.EntityDefinitions[0].Instances, 1)
	assert.Equal(t, "1", pagedResp.Msg.EntityDefinitions[0].Instances[0].UniqueKey)
}

func TestGetEntitiesUnfilteredIncludesConfiguredProperties(t *testing.T) {
	entities.ClearEntitiesOfType("server")
	entities.AddEntity("server", "0", map[string]any{"name": "alpha", "hostname": "alpha.example.com", "ip": "10.0.0.1"})
	t.Cleanup(func() {
		entities.ClearEntitiesOfType("server")
	})

	cfg := config.DefaultConfig()
	cfg.Entities = []*config.EntityFile{
		{
			Name: "server",
			Properties: []config.EntityProperty{
				{Name: "hostname", Title: "Hostname"},
			},
		},
	}
	cfg.Sanitize()

	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()
	ts, client := getNewTestServerAndClientWithExecutor(cfg, ex)
	defer ts.Close()

	resp, err := client.GetEntities(context.Background(), connect.NewRequest(&apiv1.GetEntitiesRequest{}))
	require.NoError(t, err)

	serverDef := findEntityDefinition(resp.Msg.EntityDefinitions, "server")
	require.NotNil(t, serverDef)
	require.Len(t, serverDef.Properties, 1)
	assert.Equal(t, "hostname", serverDef.Properties[0].Name)
	assert.Equal(t, int32(1), serverDef.TotalInstances)
	assert.Empty(t, serverDef.Instances)
}

func TestPaginateEntityInstancesHandlesLargePageValues(t *testing.T) {
	instances := []*apiv1.Entity{
		{UniqueKey: "0"},
		{UniqueKey: "1"},
	}

	assert.Empty(t, paginateEntityInstances(instances, 1<<30, 1))
	assert.Empty(t, paginateEntityInstances(instances, 2, 1<<30))
	assert.Equal(t, []*apiv1.Entity{{UniqueKey: "1"}}, paginateEntityInstances(instances, 2, 1))
}

func findEntityDefinition(definitions []*apiv1.EntityDefinition, title string) *apiv1.EntityDefinition {
	for _, definition := range definitions {
		if definition.Title == title {
			return definition
		}
	}

	return nil
}

func TestGetEntityRestrictsFieldsToConfiguredProperties(t *testing.T) {
	entities.ClearEntitiesOfType("server")
	entities.AddEntity("server", "0", map[string]any{
		"name":     "alpha",
		"hostname": "alpha.example.com",
		"ip":       "10.0.0.1",
		"groups":   []string{"admins"},
	})
	t.Cleanup(func() {
		entities.ClearEntitiesOfType("server")
	})

	cfg := config.DefaultConfig()
	cfg.Entities = []*config.EntityFile{
		{
			Name: "server",
			Properties: []config.EntityProperty{
				{Name: "hostname", Title: "Hostname"},
				{Name: "ip", Title: "IP"},
			},
		},
	}
	cfg.Sanitize()

	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()
	ts, client := getNewTestServerAndClientWithExecutor(cfg, ex)
	defer ts.Close()

	resp, err := client.GetEntity(context.Background(), connect.NewRequest(&apiv1.GetEntityRequest{
		Type:      "server",
		UniqueKey: "0",
	}))
	require.NoError(t, err)
	require.NotNil(t, resp.Msg)

	assert.Equal(t, "alpha.example.com", resp.Msg.Fields["hostname"])
	assert.Equal(t, "10.0.0.1", resp.Msg.Fields["ip"])
	assert.NotContains(t, resp.Msg.Fields, "groups")
	assert.NotContains(t, resp.Msg.Fields, "name")
}
