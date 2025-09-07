package api

import (
	"sort"

	apiv1 "github.com/OliveTin/OliveTin/gen/olivetin/api/v1"
	config "github.com/OliveTin/OliveTin/internal/config"
	"golang.org/x/exp/slices"
)

func dashboardCfgToPb(rr *DashboardRenderRequest, dashboardTitle string) *apiv1.Dashboard {
	if dashboardTitle == "default" {
		return buildDefaultDashboard(rr)
	}

	for _, dashboard := range rr.cfg.Dashboards {
		if dashboard.Title != dashboardTitle {
			continue
		}

		return &apiv1.Dashboard{
			Title:    dashboard.Title,
			Contents: sortActions(removeNulls(getDashboardComponentContents(dashboard, rr))),
		}

		/*
			if len(pbdb.Contents) == 0 {
				log.WithFields(log.Fields{
					"dashboard": dashboard.Title,
					"username":  rr.AuthenticatedUser.Username,
				}).Debugf("Dashboard has no readable contents, so it will not be visible in the web ui")
				continue
			}
		*/
	}

	return nil
}

//gocyclo:ignore
func buildDefaultDashboard(rr *DashboardRenderRequest) *apiv1.Dashboard {
	fieldset := &apiv1.DashboardComponent{
		Type:     "fieldset",
		Contents: make([]*apiv1.DashboardComponent, 0),
		Title:    "Default",
	}

	actions := make([]*apiv1.Action, 0)

	for id, binding := range rr.ex.MapActionIdToBinding {
		if binding.Action.Hidden {
			continue
		}

		if binding.IsOnDashboard {
			continue
		}

		actions = append(actions, buildAction(id, binding, rr))
	}

	for _, action := range actions {
		fieldset.Contents = append(fieldset.Contents, &apiv1.DashboardComponent{
			Type:   "link",
			Title:  action.Title,
			Icon:   action.Icon,
			Action: action,
		})
	}

	fieldset.Contents = sortActions(fieldset.Contents)

	return &apiv1.Dashboard{
		Title:    "Default",
		Contents: []*apiv1.DashboardComponent{fieldset},
	}
}

func sortActions(components []*apiv1.DashboardComponent) []*apiv1.DashboardComponent {
	sort.Slice(components, func(i, j int) bool {
		if components[i].Action == nil {
			return false
		}

		if components[j].Action == nil {
			return true
		}

		if components[i].Action.Order == components[j].Action.Order {
			return components[i].Action.Title < components[j].Action.Title
		} else {
			return components[i].Action.Order < components[j].Action.Order
		}
	})

	return components
}

func removeNulls(components []*apiv1.DashboardComponent) []*apiv1.DashboardComponent {
	ret := make([]*apiv1.DashboardComponent, 0)

	for _, component := range components {
		if component == nil {
			continue
		}

		ret = append(ret, component)
	}

	return ret
}

func getDashboardComponentContents(dashboard *config.DashboardComponent, rr *DashboardRenderRequest) []*apiv1.DashboardComponent {
	ret := make([]*apiv1.DashboardComponent, 0)

	for _, subitem := range dashboard.Contents {
		if subitem.Type == "fieldset" && subitem.Entity != "" {
			ret = append(ret, buildEntityFieldsets(subitem.Entity, subitem, rr)...)
		} else {
			ret = append(ret, buildDashboardComponentSimple(subitem, rr))
		}
	}

	return ret
}

func buildDashboardComponentSimple(subitem *config.DashboardComponent, rr *DashboardRenderRequest) *apiv1.DashboardComponent {
	newitem := &apiv1.DashboardComponent{
		Title:    subitem.Title,
		Type:     getDashboardComponentType(subitem),
		Contents: getDashboardComponentContents(subitem, rr),
		Icon:     getDashboardComponentIcon(subitem, rr.cfg),
		CssClass: subitem.CssClass,
		Action:   rr.findAction(subitem.Title),
	}

	return newitem
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
