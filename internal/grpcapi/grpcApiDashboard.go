package grpcapi

import (
	pb "github.com/OliveTin/OliveTin/gen/grpc"
	config "github.com/OliveTin/OliveTin/internal/config"
)

func dashboardCfgToPb(res *pb.GetDashboardComponentsResponse, dashboards []*config.DashboardComponent) {
	for _, dashboard := range dashboards {
		res.Dashboards = append(res.Dashboards, &pb.DashboardComponent{
			Type:     "dashboard",
			Title:    dashboard.Title,
			Contents: getDashboardComponentContents(dashboard),
		})
	}
}

func getDashboardComponentContents(dashboard *config.DashboardComponent) []*pb.DashboardComponent {
	ret := make([]*pb.DashboardComponent, 0)

	for _, subitem := range dashboard.Contents {
		if subitem.Type == "fieldset" && subitem.Entity != "" {
			ret = append(ret, buildEntityFieldsets(subitem.Entity, &subitem)...)
			continue
		}

		newitem := &pb.DashboardComponent{
			Title:    subitem.Title,
			Type:     getDashboardComponentType(&subitem),
			Contents: getDashboardComponentContents(&subitem),
		}

		ret = append(ret, newitem)
	}

	return ret
}

func getDashboardComponentType(item *config.DashboardComponent) string {
	if len(item.Contents) > 0 {
		if item.Type != "fieldset" {
			return "directory"
		}

		return "fieldset"
	} else if item.Type == "display" {
		return "display"
	}

	return "link"
}
