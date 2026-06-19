package executor

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/entities"
)

func TestResolveOnDashboardsDefaultsToActionsDashboard(t *testing.T) {
	index := buildDashboardTargetIndex(&config.Config{})

	targets := resolveOnDashboards(index, "Lonely Action", "")

	require.Len(t, targets, 1)
	assert.Equal(t, "Actions", targets[0].Title)
	assert.Equal(t, "/", targets[0].Path)
}

func TestResolveOnDashboardsConfiguredDashboard(t *testing.T) {
	cfg := &config.Config{
		Actions: []*config.Action{
			{Title: "Restart"},
		},
		Dashboards: []*config.DashboardComponent{
			{
				Title: "Operations",
				Contents: []*config.DashboardComponent{
					{Title: "Restart"},
				},
			},
		},
	}

	index := buildDashboardTargetIndex(cfg)
	targets := resolveOnDashboards(index, "Restart", "")

	require.Len(t, targets, 1)
	assert.Equal(t, "Operations", targets[0].Title)
	assert.Equal(t, "/dashboards/Operations", targets[0].Path)
}

func TestResolveOnDashboardsEntityDirectory(t *testing.T) {
	cfg := &config.Config{
		Actions: []*config.Action{
			{Title: "Reboot", Entity: "host"},
		},
		Dashboards: []*config.DashboardComponent{
			{
				Title: "Servers",
				Contents: []*config.DashboardComponent{
					{
						Type:   "fieldset",
						Entity: "host",
						Contents: []*config.DashboardComponent{
							{
								Type:  "directory",
								Title: "Host Details",
								Contents: []*config.DashboardComponent{
									{Title: "Reboot"},
								},
							},
						},
					},
				},
			},
		},
	}

	entities.ClearEntitiesOfType("host")
	defer entities.ClearEntitiesOfType("host")
	entities.AddEntity("host", "host-1", map[string]any{"title": "Host 1"})

	index := buildDashboardTargetIndex(cfg)
	targets := resolveOnDashboards(index, "Reboot", "host-1")

	require.Len(t, targets, 1)
	assert.Equal(t, "Host Details", targets[0].Title)
	assert.Equal(t, "host", targets[0].EntityType)
	assert.Equal(t, "host-1", targets[0].EntityKey)
	assert.Equal(t, "/dashboards/Host Details/host/host-1", targets[0].Path)
}

func TestRebuildActionMapStoresOnDashboards(t *testing.T) {
	cfg := &config.Config{
		Actions: []*config.Action{
			{Title: "Only Default"},
			{Title: "Configured"},
		},
		Dashboards: []*config.DashboardComponent{
			{
				Title: "Custom",
				Contents: []*config.DashboardComponent{
					{Title: "Configured"},
				},
			},
		},
	}

	ex := DefaultExecutor(cfg)
	ex.RebuildActionMap()

	defaultBinding := ex.FindBindingWithNoEntity(cfg.Actions[0])
	require.NotNil(t, defaultBinding)
	require.Len(t, defaultBinding.OnDashboards, 1)
	assert.Equal(t, "Actions", defaultBinding.OnDashboards[0].Title)
	assert.False(t, defaultBinding.IsOnConfiguredDashboard())

	configuredBinding := ex.FindBindingWithNoEntity(cfg.Actions[1])
	require.NotNil(t, configuredBinding)
	require.Len(t, configuredBinding.OnDashboards, 1)
	assert.Equal(t, "Custom", configuredBinding.OnDashboards[0].Title)
	assert.True(t, configuredBinding.IsOnConfiguredDashboard())
}

func TestResolveOnDashboardsMultipleDashboards(t *testing.T) {
	cfg := &config.Config{
		Actions: []*config.Action{
			{Title: "Shared"},
		},
		Dashboards: []*config.DashboardComponent{
			{
				Title: "One",
				Contents: []*config.DashboardComponent{
					{Title: "Shared"},
				},
			},
			{
				Title: "Two",
				Contents: []*config.DashboardComponent{
					{Title: "Shared"},
				},
			},
		},
	}

	index := buildDashboardTargetIndex(cfg)
	targets := resolveOnDashboards(index, "Shared", "")

	require.Len(t, targets, 2)
	assert.Equal(t, "One", targets[0].Title)
	assert.Equal(t, "Two", targets[1].Title)
}
