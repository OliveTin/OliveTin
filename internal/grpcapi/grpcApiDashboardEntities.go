package grpcapi

import (
	pb "github.com/OliveTin/OliveTin/gen/grpc"
	config "github.com/OliveTin/OliveTin/internal/config"
	sv "github.com/OliveTin/OliveTin/internal/stringvariables"

	"fmt"
	"strconv"
)

func getEntityPrefix(entityTitle string, entityIndex int) string {
	return "entities." + entityTitle + "." + fmt.Sprintf("%v", entityIndex)
}

func buildEntityFieldsets(entityTitle string, tpl *config.DashboardComponent) []*pb.DashboardComponent {
	ret := make([]*pb.DashboardComponent, 0)

	entityCount, _ := strconv.Atoi(sv.Get("entities." + entityTitle + ".count"))

	for i := 0; i < entityCount; i++ {
		ret = append(ret, buildEntityFieldset(tpl, entityTitle, i))
	}

	return ret
}

func buildEntityFieldset(tpl *config.DashboardComponent, entityTitle string, entityIndex int) *pb.DashboardComponent {
	prefix := getEntityPrefix(entityTitle, entityIndex)

	return &pb.DashboardComponent{
		Title:    sv.ReplaceEntityVars(prefix, tpl.Title),
		Type:     "fieldset",
		Contents: buildEntityFieldsetContents(tpl.Contents, prefix),
	}
}

func buildEntityFieldsetContents(contents []config.DashboardComponent, prefix string) []*pb.DashboardComponent {
	ret := make([]*pb.DashboardComponent, 0)

	for _, subitem := range contents {
		clone := &pb.DashboardComponent{}

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
