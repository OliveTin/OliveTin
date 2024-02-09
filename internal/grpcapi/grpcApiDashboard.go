package grpcapi

import (
	pb "github.com/OliveTin/OliveTin/gen/grpc"
	config "github.com/OliveTin/OliveTin/internal/config"
)

func dashboardCfgToPb(res *pb.GetDashboardComponentsResponse, dashboards []config.DashboardItem) {
	for _, dashboard := range dashboards {
		res.Dashboards = append(res.Dashboards, &pb.DashboardItem{
			Type:     "dashboard",
			Title:    dashboard.Title,
			Contents: getDashboardContents(&dashboard),
		})
	}
}

func getDashboardContents(dashboard *config.DashboardItem) []*pb.DashboardItem {
	ret := make([]*pb.DashboardItem, 0)

	for _, subitem := range dashboard.Contents {
		newitem := &pb.DashboardItem{
			Title: subitem.Title,
			Type:  subitem.Type,
		}

		if len(subitem.Contents) > 0 {
			if newitem.Type != "fieldset" {
				newitem.Type = "directory"
			}

			newitem.Contents = getDashboardContents(&subitem)
		} else {
			newitem.Type = "link"
			newitem.Link = subitem.Link
		}

		ret = append(ret, newitem)
	}

	return ret
}
