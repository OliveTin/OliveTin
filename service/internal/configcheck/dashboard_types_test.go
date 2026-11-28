package configcheck_test

import (
	"testing"

	"github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/configcheck"
	"github.com/OliveTin/OliveTin/internal/configissues"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRebuildUnknownDashboardComponentType(t *testing.T) {
	configissues.BeginConfigLoad()
	cfg := config.DefaultConfig()
	cfg.Dashboards = []*config.DashboardComponent{
		{
			Title: "Containers",
			Contents: []*config.DashboardComponent{
				{
					Title:  "Container {{ .CurrentEntity.Names }}",
					Entity: "container",
					Type:   "myexample",
					Contents: []*config.DashboardComponent{
						{Title: "Start", Type: "display"},
					},
				},
			},
		},
	}

	configcheck.Rebuild(cfg)

	issues := issuesWithCode(configissues.List(), configissues.CodeDashboardTypeUnknown)
	require.Len(t, issues, 1)
	assert.Equal(t, "myexample", issues[0].Source)
	assert.Contains(t, issues[0].Message, "Containers")
}

func TestRebuildKnownDashboardComponentTypes(t *testing.T) {
	configissues.BeginConfigLoad()
	cfg := config.DefaultConfig()
	cfg.Dashboards = []*config.DashboardComponent{
		{
			Title: "Everything",
			Contents: []*config.DashboardComponent{
				{Title: "a"},
				{Title: "b", Type: "fieldset"},
				{Title: "c", Type: "directory"},
				{Title: "d", Type: "display"},
				{Title: "e", Type: "link"},
				{Title: "f", Type: "stdout-most-recent-execution"},
			},
		},
	}

	configcheck.Rebuild(cfg)

	assert.False(t, hasCode(configissues.List(), configissues.CodeDashboardTypeUnknown))
}

func issuesWithCode(issues []configissues.Issue, code string) []configissues.Issue {
	out := make([]configissues.Issue, 0)
	for _, issue := range issues {
		if issue.Code == code {
			out = append(out, issue)
		}
	}
	return out
}
