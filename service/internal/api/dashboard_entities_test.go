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

func TestBuildEntityFieldsetDisplayRendersEntityHtmlTitleAndCssClass(t *testing.T) {
	entities.ClearEntitiesOfType("html_display")
	defer entities.ClearEntitiesOfType("html_display")

	entities.AddEntity("html_display", "0", map[string]any{
		"content": "<div class=\"content\">test</div>",
	})

	cfg := config.DefaultConfig()
	cfg.Dashboards = []*config.DashboardComponent{
		{
			Title: "Stream status",
			Contents: []*config.DashboardComponent{
				{
					Title:  "Compare result",
					Type:   "fieldset",
					Entity: "html_display",
					Contents: []*config.DashboardComponent{
						{
							Type:     "display",
							CssClass: "full_screen",
							Title:    "{{ html_display.content }}",
						},
					},
				},
			},
		},
	}

	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()

	rr := &DashboardRenderRequest{
		cfg: cfg,
		ex:  ex,
	}

	fieldsets := buildEntityFieldsets("html_display", cfg.Dashboards[0].Contents[0], rr)
	require.Len(t, fieldsets, 1)

	display := findComponentByType(fieldsets[0].Contents, "display")
	require.NotNil(t, display)
	assert.Equal(t, "full_screen", display.CssClass)
	assert.Equal(t, "<div class=\"content\">test</div>", display.Title)
}

func findComponentByType(components []*apiv1.DashboardComponent, componentType string) *apiv1.DashboardComponent {
	for _, component := range components {
		if component.Type == componentType {
			return component
		}

		if found := findNestedComponent(component, componentType); found != nil {
			return found
		}
	}

	return nil
}

func findNestedComponent(component *apiv1.DashboardComponent, componentType string) *apiv1.DashboardComponent {
	if len(component.Contents) == 0 {
		return nil
	}

	return findComponentByType(component.Contents, componentType)
}

func TestGetDashboardEntityDisplayHtmlTitle(t *testing.T) {
	entities.ClearEntitiesOfType("html_display")
	defer entities.ClearEntitiesOfType("html_display")

	entities.AddEntity("html_display", "0", map[string]any{
		"content": "<div class=\"content\">test</div>",
	})

	cfg := config.DefaultConfig()
	cfg.Dashboards = []*config.DashboardComponent{
		{
			Title: "Html Dashboard",
			Contents: []*config.DashboardComponent{
				{
					Title:  "Compare result",
					Type:   "fieldset",
					Entity: "html_display",
					Contents: []*config.DashboardComponent{
						{
							Type:     "display",
							CssClass: "full_screen",
							Title:    "{{ html_display.content }}",
						},
					},
				},
			},
		},
	}

	ts, client := getNewTestServerAndClient(cfg)
	defer ts.Close()

	resp, err := client.GetDashboard(context.Background(), connect.NewRequest(&apiv1.GetDashboardRequest{
		Title: "Html Dashboard",
	}))
	require.NoError(t, err)

	display := findComponentByType(resp.Msg.Dashboard.Contents, "display")
	require.NotNil(t, display)
	assert.Equal(t, "full_screen", display.CssClass)
	assert.Equal(t, "<div class=\"content\">test</div>", display.Title)
}
