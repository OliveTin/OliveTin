package grpcapi

import (
	apiv1 "github.com/OliveTin/OliveTin/gen/grpc/olivetin/api/v1"
	config "github.com/OliveTin/OliveTin/internal/config"
	"golang.org/x/exp/slices"
)

func dashboardCfgToPb(res *apiv1.GetDashboardComponentsResponse, dashboards []*config.DashboardComponent, cfg *config.Config) {
	for _, dashboard := range dashboards {
		res.Dashboards = append(res.Dashboards, &apiv1.DashboardComponent{
			Type:     "dashboard",
			Title:    dashboard.Title,
			Contents: getDashboardComponentContents(dashboard, cfg),
		})
	}
}

func getDashboardComponentContents(dashboard *config.DashboardComponent, cfg *config.Config) []*apiv1.DashboardComponent {
	ret := make([]*apiv1.DashboardComponent, 0)

	for _, subitem := range dashboard.Contents {
		if subitem.Type == "fieldset" && subitem.Entity != "" {
			ret = append(ret, buildEntityFieldsets(subitem.Entity, &subitem)...)
			continue
		}

		newitem := &apiv1.DashboardComponent{
			Title:    subitem.Title,
			Type:     getDashboardComponentType(&subitem),
			Contents: getDashboardComponentContents(&subitem, cfg),
			Icon:     getDashboardComponentIcon(&subitem, cfg),
			CssClass: subitem.CssClass,
		}

		ret = append(ret, newitem)
	}

	return ret
}

func getDashboardComponentIcon(item *config.DashboardComponent, cfg *config.Config) string {
	if item.Icon == "" {
		return cfg.DefaultIconForDirectories
	}

	return item.Icon
}

func getDashboardComponentType(item *config.DashboardComponent) string {
	allowedTypes := []string{
		"stdout-most-recent-execution",
		"display",
	}

	if len(item.Contents) > 0 {
		if item.Type != "fieldset" {
			return "directory"
		}

		return "fieldset"
	} else if slices.Contains(allowedTypes, item.Type) {
		return item.Type
	}

	return "link"
}
