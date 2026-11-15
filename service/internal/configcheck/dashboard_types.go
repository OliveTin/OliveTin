package configcheck

import (
	"fmt"
	"slices"

	"github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/configissues"
)

// knownDashboardComponentTypes are the dashboard component types the web UI
// can render. An empty type lets OliveTin pick one from the component's contents.
var knownDashboardComponentTypes = []string{
	"",
	"fieldset",
	"directory",
	"display",
	"link",
	"stdout-most-recent-execution",
}

func collectDashboardTypeIssues(cfg *config.Config) []configissues.Issue {
	out := make([]configissues.Issue, 0)
	visited := make(map[*config.DashboardComponent]bool)

	for _, dashboard := range cfg.Dashboards {
		out = append(out, dashboardTypeIssues(dashboard, dashboard, visited)...)
	}

	return out
}

func dashboardTypeIssues(dashboard, component *config.DashboardComponent, visited map[*config.DashboardComponent]bool) []configissues.Issue {
	if component == nil || visited[component] {
		return nil
	}

	visited[component] = true

	out := unknownDashboardTypeIssue(dashboard, component)

	for _, child := range component.Contents {
		out = append(out, dashboardTypeIssues(dashboard, child, visited)...)
	}

	return out
}

func unknownDashboardTypeIssue(dashboard, component *config.DashboardComponent) []configissues.Issue {
	if slices.Contains(knownDashboardComponentTypes, component.Type) {
		return nil
	}

	return []configissues.Issue{{
		Severity: configissues.SeverityWarning,
		Code:     configissues.CodeDashboardTypeUnknown,
		Message: fmt.Sprintf("Dashboard %q component %q has unknown type %q; a default type will be used instead",
			dashboard.Title, component.Title, component.Type),
		Source: component.Type,
	}}
}
