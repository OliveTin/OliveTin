package grpcapi

import (
	apiv1 "github.com/OliveTin/OliveTin/gen/grpc/olivetin/api/v1"
	config "github.com/OliveTin/OliveTin/internal/config"
	sv "github.com/OliveTin/OliveTin/internal/stringvariables"
	"golang.org/x/exp/slices"
)

func buildEntityFieldsets(entityTitle string, tpl *config.DashboardComponent, rr *DashboardRenderRequest) []*apiv1.DashboardComponent {
	ret := make([]*apiv1.DashboardComponent, 0)

	entityCount := sv.GetEntityCount(entityTitle)

	for i := range entityCount {
		fs := buildEntityFieldset(tpl, entityTitle, i, rr)

		if len(fs.Contents) > 0 {
			ret = append(ret, fs)
		} else {
			// If the fieldset has no contents, we don't want to show it
			continue
		}
	}

	return ret
}

func buildEntityFieldset(tpl *config.DashboardComponent, entityTitle string, entityIndex int, rr *DashboardRenderRequest) *apiv1.DashboardComponent {
	prefix := sv.GetEntityPrefix(entityTitle, entityIndex)

	return &apiv1.DashboardComponent{
		Title:    sv.ReplaceEntityVars(prefix, tpl.Title),
		Type:     "fieldset",
		Contents: removeFieldsetsWithoutLinks(buildEntityFieldsetContents(tpl.Contents, prefix, rr)),
		CssClass: sv.ReplaceEntityVars(prefix, tpl.CssClass),
	}
}

func removeFieldsetsWithoutLinks(contents []*apiv1.DashboardComponent) []*apiv1.DashboardComponent {
	hasLinks := false

	for _, subitem := range contents {
		if subitem.Type == "link" {
			hasLinks = true
			break
		}
	}

	if hasLinks {
		return contents
	}

	return nil
}

func buildEntityFieldsetContents(contents []config.DashboardComponent, prefix string, rr *DashboardRenderRequest) []*apiv1.DashboardComponent {
	ret := make([]*apiv1.DashboardComponent, 0)

	for _, subitem := range contents {
		c := cloneItem(&subitem, prefix, rr)

		if c != nil {
			ret = append(ret, c)
		}
	}

	return ret
}

func cloneItem(subitem *config.DashboardComponent, prefix string, rr *DashboardRenderRequest) *apiv1.DashboardComponent {
	clone := &apiv1.DashboardComponent{}
	clone.CssClass = sv.ReplaceEntityVars(prefix, subitem.CssClass)

	if subitem.Type == "" || subitem.Type == "link" {
		clone.Type = "link"
		clone.Title = sv.ReplaceEntityVars(prefix, subitem.Title)

		if !slices.Contains(rr.AllowedActionTitles, clone.Title) {
			return nil
		}
	} else {
		clone.Title = sv.ReplaceEntityVars(prefix, subitem.Title)
		clone.Type = subitem.Type
	}

	return clone
}
