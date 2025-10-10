package api

import (
	"sort"

	apiv1 "github.com/OliveTin/OliveTin/gen/olivetin/api/v1"
	config "github.com/OliveTin/OliveTin/internal/config"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

func renderDashboard(rr *DashboardRenderRequest, dashboardTitle string) *apiv1.Dashboard {
	if dashboardTitle == "default" {
		return buildDefaultDashboard(rr)
	}

	return findAndRenderDashboard(rr, dashboardTitle)
}

func findAndRenderDashboard(rr *DashboardRenderRequest, dashboardTitle string) *apiv1.Dashboard {
	for _, dashboard := range rr.cfg.Dashboards {
		if dashboard.Title != dashboardTitle {
			continue
		}

		if len(dashboard.Contents) == 0 {
			logEmptyDashboard(dashboard.Title, rr.AuthenticatedUser.Username)
			return nil
		}

		return buildDashboardFromConfig(dashboard, rr)
	}

	return nil
}

func logEmptyDashboard(dashboardTitle, username string) {
	log.WithFields(log.Fields{
		"dashboard": dashboardTitle,
		"username":  username,
	}).Debugf("Dashboard has no readable contents, so it will not be visible in the web ui")
}

func buildDashboardFromConfig(dashboard *config.DashboardComponent, rr *DashboardRenderRequest) *apiv1.Dashboard {
	return &apiv1.Dashboard{
		Title:    dashboard.Title,
		Contents: sortActions(removeNulls(getDashboardComponentContents(dashboard, rr))),
	}
}

//gocyclo:ignore
func buildDefaultDashboard(rr *DashboardRenderRequest) *apiv1.Dashboard {
	db := &apiv1.Dashboard{
		Title:    "Actions",
		Contents: make([]*apiv1.DashboardComponent, 0),
	}

	fieldset := &apiv1.DashboardComponent{
		Type:     "fieldset",
		Title:    "Actions",
		Contents: make([]*apiv1.DashboardComponent, 0),
	}

	for _, binding := range rr.ex.MapActionIdToBinding {
		if binding.Action.Hidden {
			continue
		}

		if binding.IsOnDashboard {
			continue
		}

		action := buildAction(binding, rr)

		fieldset.Contents = append(fieldset.Contents, &apiv1.DashboardComponent{
			Type:   "link",
			Title:  action.Title,
			Icon:   action.Icon,
			Action: action,
		})
	}

	if len(fieldset.Contents) > 0 {
		fieldset.Contents = sortActions(fieldset.Contents)
		db.Contents = append(db.Contents, fieldset)
	}

	return db
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
	rootFieldset := createRootFieldset()

	for _, subitem := range dashboard.Contents {
		processDashboardSubitem(subitem, rr, &ret, rootFieldset)
	}

	return appendRootFieldsetIfNeeded(ret, rootFieldset)
}

func createRootFieldset() *apiv1.DashboardComponent {
	return &apiv1.DashboardComponent{
		Type:     "fieldset",
		Title:    "Actions",
		Contents: make([]*apiv1.DashboardComponent, 0),
	}
}

func processDashboardSubitem(subitem *config.DashboardComponent, rr *DashboardRenderRequest, ret *[]*apiv1.DashboardComponent, rootFieldset *apiv1.DashboardComponent) {
	if subitem.Type != "fieldset" {
		rootFieldset.Contents = append(rootFieldset.Contents, buildDashboardComponentSimple(subitem, rr))
		return
	}

	if subitem.Entity != "" {
		*ret = append(*ret, buildEntityFieldsets(subitem.Entity, subitem, rr)...)
	} else {
		*ret = append(*ret, buildDashboardComponentSimple(subitem, rr))
	}
}

func appendRootFieldsetIfNeeded(ret []*apiv1.DashboardComponent, rootFieldset *apiv1.DashboardComponent) []*apiv1.DashboardComponent {
	if len(rootFieldset.Contents) > 0 {
		ret = append(ret, rootFieldset)
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
