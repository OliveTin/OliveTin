package grpcapi

import (
	pb "github.com/OliveTin/OliveTin/gen/grpc"
	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/entities"
)

func buildEntityFieldsets(entityTitle string, tpl *config.DashboardComponent) []*pb.DashboardComponent {
	ret := make([]*pb.DashboardComponent, 0)

	entities := entities.GetEntities(entityTitle)

	for _, ent := range entities {
		ret = append(ret, buildEntityFieldset(tpl, ent))
	}

	return ret
}

func buildEntityFieldset(tpl *config.DashboardComponent, ent interface{}) *pb.DashboardComponent {
	return &pb.DashboardComponent{
		Title:    entities.ParseTemplateWith(tpl.Title, ent),
		Type:     "fieldset",
		Contents: buildEntityFieldsetContents(tpl.Contents, ent),
		CssClass: "foobar",
	}
}

func buildEntityFieldsetContents(contents []config.DashboardComponent, ent interface{}) []*pb.DashboardComponent {
	ret := make([]*pb.DashboardComponent, 0)

	for _, subitem := range contents {
		clone := &pb.DashboardComponent{}
		clone.CssClass = "blat"

		if subitem.Type == "" || subitem.Type == "link" {
			clone.Type = "link"
			clone.Title = entities.ParseTemplateWith(subitem.Title, ent)
		} else {
			clone.Title = entities.ParseTemplateWith(subitem.Title, ent)
			clone.Type = subitem.Type
		}

		ret = append(ret, clone)
	}

	return ret
}
