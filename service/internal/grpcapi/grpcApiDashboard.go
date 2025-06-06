package grpcapi

import (
	apiv1 "github.com/OliveTin/OliveTin/gen/grpc/olivetin/api/v1"
	config "github.com/OliveTin/OliveTin/internal/config"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

func dashboardCfgToPb(rr *DashboardRenderRequest) []*apiv1.DashboardComponent {
	ret := make([]*apiv1.DashboardComponent, 0)

	for _, dashboard := range cfg.Dashboards {
		pbdb := &apiv1.DashboardComponent{
			Type:     "dashboard",
			Title:    dashboard.Title,
			Contents: removeNulls(getDashboardComponentContents(dashboard, rr)),
		}

		if len(pbdb.Contents) == 0 {
			log.WithFields(log.Fields{
				"dashboard": dashboard.Title,
				"username":  rr.AuthenticatedUser.Username,
			}).Debugf("Dashboard has no readable contents, so it will not be visible in the web ui")
			continue
		}

		ret = append(ret, pbdb)
	}

	return ret
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
			ret = append(ret, buildEntityFieldsets(subitem.Entity, &subitem, rr)...)
		} else {
			ret = append(ret, buildDashboardComponentSimple(&subitem, rr))
		}
	}

	return ret
}

func buildDashboardComponentSimple(subitem *config.DashboardComponent, rr *DashboardRenderRequest) *apiv1.DashboardComponent {
	if !slices.Contains(rr.AllowedActionTitles, subitem.Title) {
		return nil
	}

	newitem := &apiv1.DashboardComponent{
		Title:    subitem.Title,
		Type:     getDashboardComponentType(subitem),
		Contents: getDashboardComponentContents(subitem, rr),
		Icon:     getDashboardComponentIcon(subitem, rr.cfg),
		CssClass: subitem.CssClass,
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
