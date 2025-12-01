package api

import (
	"sort"

	apiv1 "github.com/OliveTin/OliveTin/gen/olivetin/api/v1"
	config "github.com/OliveTin/OliveTin/internal/config"
	entities "github.com/OliveTin/OliveTin/internal/entities"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

func renderDashboard(rr *DashboardRenderRequest, dashboardTitle string) *apiv1.Dashboard {
	if dashboardTitle == "default" {
		return buildDefaultDashboard(rr)
	}

	return findAndRenderDashboard(rr, dashboardTitle)
}

func getEntityFromRequest(rr *DashboardRenderRequest) *entities.Entity {
	if rr.EntityType == "" || rr.EntityKey == "" {
		return nil
	}

	entityInstances := entities.GetEntityInstances(rr.EntityType)
	if entity, ok := entityInstances[rr.EntityKey]; ok {
		return entity
	}

	return nil
}

func findAndRenderDashboard(rr *DashboardRenderRequest, dashboardTitle string) *apiv1.Dashboard {
	if dashboard := findDashboardByTitle(rr, dashboardTitle); dashboard != nil {
		return renderDashboardIfValid(dashboard, rr)
	}

	return renderDirectoryDashboard(rr, dashboardTitle)
}

func findDashboardByTitle(rr *DashboardRenderRequest, dashboardTitle string) *config.DashboardComponent {
	for _, dashboard := range rr.cfg.Dashboards {
		if dashboard.Title == dashboardTitle {
			return dashboard
		}
	}
	return nil
}

func renderDashboardIfValid(dashboard *config.DashboardComponent, rr *DashboardRenderRequest) *apiv1.Dashboard {
	if len(dashboard.Contents) == 0 {
		logEmptyDashboard(dashboard.Title, rr.AuthenticatedUser.Username)
		return nil
	}
	return buildDashboardFromConfig(dashboard, rr)
}

func renderDirectoryDashboard(rr *DashboardRenderRequest, dashboardTitle string) *apiv1.Dashboard {
	directoryComponent := findDirectoryComponent(rr, dashboardTitle)
	if directoryComponent == nil {
		return nil
	}

	entity := getEntityFromRequest(rr)
	return buildDashboardFromConfigWithEntity(directoryComponent, rr, entity)
}

func findDirectoryComponent(rr *DashboardRenderRequest, title string) *config.DashboardComponent {
	for _, dashboard := range rr.cfg.Dashboards {
		if component := searchDirectoryInComponent(dashboard, title); component != nil {
			return component
		}
	}
	return nil
}

func searchDirectoryInComponent(component *config.DashboardComponent, title string) *config.DashboardComponent {
	if isMatchingDirectory(component, title) {
		return component
	}

	return searchDirectoryInSubcomponents(component.Contents, title)
}

func isMatchingDirectory(component *config.DashboardComponent, title string) bool {
	return component.Title == title && len(component.Contents) > 0 && component.Type != "fieldset"
}

func searchDirectoryInSubcomponents(contents []*config.DashboardComponent, title string) *config.DashboardComponent {
	for _, subitem := range contents {
		if found := searchDirectoryInComponent(subitem, title); found != nil {
			return found
		}
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
	return buildDashboardFromConfigWithEntity(dashboard, rr, nil)
}

func buildDashboardFromConfigWithEntity(dashboard *config.DashboardComponent, rr *DashboardRenderRequest, entity *entities.Entity) *apiv1.Dashboard {
	return &apiv1.Dashboard{
		Title:    dashboard.Title,
		Contents: sortActions(removeNulls(getDashboardComponentContentsWithEntity(dashboard, rr, entity))),
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
		if components[i].Action == nil || components[j].Action == nil {
			return components[i].Title < components[j].Title
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
	return getDashboardComponentContentsWithEntity(dashboard, rr, nil)
}

func getDashboardComponentContentsWithEntity(dashboard *config.DashboardComponent, rr *DashboardRenderRequest, entity *entities.Entity) []*apiv1.DashboardComponent {
	ret := make([]*apiv1.DashboardComponent, 0)
	rootFieldset := createRootFieldset()

	for _, subitem := range dashboard.Contents {
		processDashboardSubitemWithEntity(subitem, rr, &ret, rootFieldset, entity)
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
	processDashboardSubitemWithEntity(subitem, rr, ret, rootFieldset, nil)
}

func processDashboardSubitemWithEntity(subitem *config.DashboardComponent, rr *DashboardRenderRequest, ret *[]*apiv1.DashboardComponent, rootFieldset *apiv1.DashboardComponent, entity *entities.Entity) {
	if subitem.Type != "fieldset" {
		rootFieldset.Contents = append(rootFieldset.Contents, buildDashboardComponentSimpleWithEntity(subitem, rr, entity))
		return
	}

	if subitem.Entity != "" {
		*ret = append(*ret, buildEntityFieldsets(subitem.Entity, subitem, rr)...)
	} else {
		*ret = append(*ret, buildDashboardComponentSimpleWithEntity(subitem, rr, entity))
	}
}

func appendRootFieldsetIfNeeded(ret []*apiv1.DashboardComponent, rootFieldset *apiv1.DashboardComponent) []*apiv1.DashboardComponent {
	if len(rootFieldset.Contents) > 0 {
		ret = append(ret, rootFieldset)
	}
	return ret
}

func buildDashboardComponentSimple(subitem *config.DashboardComponent, rr *DashboardRenderRequest) *apiv1.DashboardComponent {
	return buildDashboardComponentSimpleWithEntity(subitem, rr, nil)
}

func buildDashboardComponentSimpleWithEntity(subitem *config.DashboardComponent, rr *DashboardRenderRequest, entity *entities.Entity) *apiv1.DashboardComponent {
	var contents []*apiv1.DashboardComponent

	if len(subitem.Contents) > 0 {
		contents = getDashboardComponentContentsWithEntity(subitem, rr, entity)
	}

	action := rr.findActionForEntity(subitem.Title, entity)
	componentType := getDashboardComponentType(subitem, action)

	title := subitem.Title
	if entity != nil {
		title = entities.ParseTemplateWith(subitem.Title, entity)
	}

	newitem := &apiv1.DashboardComponent{
		Title:    title,
		Type:     componentType,
		Contents: contents,
		Icon:     getDashboardComponentIcon(subitem, rr.cfg),
		CssClass: subitem.CssClass,
		Action:   action,
	}

	return newitem
}

func getDashboardComponentIcon(item *config.DashboardComponent, cfg *config.Config) string {
	if item.Icon == "" {
		return cfg.DefaultIconForDirectories
	}

	return item.Icon
}

func getDashboardComponentType(item *config.DashboardComponent, action *apiv1.Action) string {
	if hasContents(item) {
		return getTypeForComponentWithContents(item)
	}

	if isAllowedType(item.Type) {
		return item.Type
	}

	return getDefaultType(action)
}

func hasContents(item *config.DashboardComponent) bool {
	return len(item.Contents) > 0
}

func getTypeForComponentWithContents(item *config.DashboardComponent) string {
	if item.Type != "fieldset" {
		return "directory"
	}
	return "fieldset"
}

func isAllowedType(itemType string) bool {
	allowedTypes := []string{
		"stdout-most-recent-execution",
		"display",
	}
	return slices.Contains(allowedTypes, itemType)
}

func getDefaultType(action *apiv1.Action) string {
	if action == nil {
		return "display"
	}
	return "link"
}
