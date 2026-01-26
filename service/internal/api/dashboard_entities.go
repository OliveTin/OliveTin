package api

import (
	apiv1 "github.com/OliveTin/OliveTin/gen/olivetin/api/v1"
	config "github.com/OliveTin/OliveTin/internal/config"
	entities "github.com/OliveTin/OliveTin/internal/entities"
	"github.com/OliveTin/OliveTin/internal/tpl"
	log "github.com/sirupsen/logrus"
)

func buildEntityFieldsets(entityTitle string, tpl *config.DashboardComponent, rr *DashboardRenderRequest) []*apiv1.DashboardComponent {
	ret := make([]*apiv1.DashboardComponent, 0)

	entities := entities.GetEntityInstances(entityTitle)

	for _, ent := range entities {
		fs := buildEntityFieldset(tpl, ent, rr)

		if len(fs.Contents) > 0 {
			ret = append(ret, fs)
		}
	}

	return ret
}

func buildEntityFieldset(component *config.DashboardComponent, ent *entities.Entity, rr *DashboardRenderRequest) *apiv1.DashboardComponent {
	return &apiv1.DashboardComponent{
		Title:      tpl.ParseTemplateWith(component.Title, ent),
		Type:       "fieldset",
		Contents:   removeFieldsetIfHasNoLinks(buildEntityFieldsetContents(component.Contents, ent, component.Entity, rr)),
		CssClass:   tpl.ParseTemplateWith(component.CssClass, ent),
		Action:     rr.findAction(component.Title),
		EntityType: component.Entity,
		EntityKey:  ent.UniqueKey,
	}
}

func removeFieldsetIfHasNoLinks(contents []*apiv1.DashboardComponent) []*apiv1.DashboardComponent {
	return contents
	/*
		for _, subitem := range contents {
			if subitem.Type == "link" {
				return contents
			}
		}

		log.Infof("removeFieldsetIfHasNoLinks: %+v", contents)

		return nil
	*/
}

func buildEntityFieldsetContents(contents []*config.DashboardComponent, ent *entities.Entity, entityType string, rr *DashboardRenderRequest) []*apiv1.DashboardComponent {
	ret := make([]*apiv1.DashboardComponent, 0)

	for _, subitem := range contents {
		c := cloneItem(subitem, ent, entityType, rr)

		log.Infof("cloneItem: %+v", c)

		if c != nil {
			ret = append(ret, c)
		}
	}

	return ret
}

func cloneItem(subitem *config.DashboardComponent, ent *entities.Entity, entityType string, rr *DashboardRenderRequest) *apiv1.DashboardComponent {
	clone := &apiv1.DashboardComponent{}
	clone.CssClass = tpl.ParseTemplateWith(subitem.CssClass, ent)

	if isLinkType(subitem.Type) {
		return cloneLinkItem(subitem, ent, clone, rr)
	}

	return cloneNonLinkItem(subitem, ent, entityType, clone, rr)
}

func isLinkType(itemType string) bool {
	return itemType == "" || itemType == "link"
}

func cloneLinkItem(subitem *config.DashboardComponent, ent *entities.Entity, clone *apiv1.DashboardComponent, rr *DashboardRenderRequest) *apiv1.DashboardComponent {
	clone.Type = "link"
	clone.Title = tpl.ParseTemplateWith(subitem.Title, ent)
	// Prefer an entity-specific action when available, but fall back to a
	// non-entity-scoped action with the same title. This allows inline actions
	// defined inside entity dashboards to work without requiring an explicit
	// entity binding.
	action := rr.findActionForEntity(subitem.Title, ent)
	if action == nil {
		action = rr.findAction(subitem.Title)
	}

	clone.Action = action
	return clone
}

func cloneNonLinkItem(subitem *config.DashboardComponent, ent *entities.Entity, entityType string, clone *apiv1.DashboardComponent, rr *DashboardRenderRequest) *apiv1.DashboardComponent {
	clone.Title = tpl.ParseTemplateWith(subitem.Title, ent)
	clone.Type = subitem.Type

	if isDirectoryWithEntity(clone.Type, ent, entityType) {
		clone.EntityType = entityType
		clone.EntityKey = ent.UniqueKey
	}

	if len(subitem.Contents) > 0 {
		clone.Contents = buildEntityFieldsetContents(subitem.Contents, ent, entityType, rr)
	}

	return clone
}

func isDirectoryWithEntity(itemType string, ent *entities.Entity, entityType string) bool {
	return itemType == "directory" && ent != nil && entityType != ""
}
