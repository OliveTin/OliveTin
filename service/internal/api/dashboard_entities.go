package api

import (
	apiv1 "github.com/OliveTin/OliveTin/gen/olivetin/api/v1"
	config "github.com/OliveTin/OliveTin/internal/config"
	entities "github.com/OliveTin/OliveTin/internal/entities"
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

func buildEntityFieldset(tpl *config.DashboardComponent, ent *entities.Entity, rr *DashboardRenderRequest) *apiv1.DashboardComponent {
	return &apiv1.DashboardComponent{
		Title:     entities.ParseTemplateWith(tpl.Title, ent),
		Type:      "fieldset",
		Contents:  removeFieldsetIfHasNoLinks(buildEntityFieldsetContents(tpl.Contents, ent, tpl.Entity, rr)),
		CssClass:  entities.ParseTemplateWith(tpl.CssClass, ent),
		Action:    rr.findAction(tpl.Title),
		EntityType: tpl.Entity,
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
	clone.CssClass = entities.ParseTemplateWith(subitem.CssClass, ent)

	if subitem.Type == "" || subitem.Type == "link" {
		clone.Type = "link"
		clone.Title = entities.ParseTemplateWith(subitem.Title, ent)
		clone.Action = rr.findActionForEntity(subitem.Title, ent)
	} else {
		clone.Title = entities.ParseTemplateWith(subitem.Title, ent)
		clone.Type = subitem.Type
		
		if clone.Type == "directory" && ent != nil && entityType != "" {
			clone.EntityType = entityType
			clone.EntityKey = ent.UniqueKey
		}
		
		if len(subitem.Contents) > 0 {
			clone.Contents = buildEntityFieldsetContents(subitem.Contents, ent, entityType, rr)
		}
	}

	return clone
}
