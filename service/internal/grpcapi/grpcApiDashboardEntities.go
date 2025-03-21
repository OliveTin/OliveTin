package grpcapi

import (
	apiv1 "github.com/OliveTin/OliveTin/gen/grpc/olivetin/api/v1"
	config "github.com/OliveTin/OliveTin/internal/config"
	sv "github.com/OliveTin/OliveTin/internal/stringvariables"
)

func buildEntityFieldsets(entityTitle string, tpl *config.DashboardComponent) []*apiv1.DashboardComponent {
	ret := make([]*apiv1.DashboardComponent, 0)

	entityCount := sv.GetEntityCount(entityTitle)

	for i := 0; i < entityCount; i++ {
		ret = append(ret, buildEntityFieldset(tpl, entityTitle, i))
	}

	return ret
}

func buildEntityFieldset(tpl *config.DashboardComponent, entityTitle string, entityIndex int) *apiv1.DashboardComponent {
	prefix := sv.GetEntityPrefix(entityTitle, entityIndex)

	return &apiv1.DashboardComponent{
		Title:    sv.ReplaceEntityVars(prefix, tpl.Title),
		Type:     "fieldset",
		Contents: buildEntityFieldsetContents(tpl.Contents, prefix),
		CssClass: sv.ReplaceEntityVars(prefix, tpl.CssClass),
	}
}

func buildEntityFieldsetContents(contents []config.DashboardComponent, prefix string) []*apiv1.DashboardComponent {
	ret := make([]*apiv1.DashboardComponent, 0)

	for _, subitem := range contents {
		clone := &apiv1.DashboardComponent{}
		clone.CssClass = sv.ReplaceEntityVars(prefix, subitem.CssClass)

		if subitem.Type == "" || subitem.Type == "link" {
			clone.Type = "link"
			clone.Title = sv.ReplaceEntityVars(prefix, subitem.Title)
		} else {
			clone.Title = sv.ReplaceEntityVars(prefix, subitem.Title)
			clone.Type = subitem.Type
		}

		ret = append(ret, clone)
	}

	return ret
}
