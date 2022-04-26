package grpcapi

import (
	"crypto/md5"
	"fmt"
	pb "github.com/OliveTin/OliveTin/gen/grpc"
	acl "github.com/OliveTin/OliveTin/internal/acl"
	config "github.com/OliveTin/OliveTin/internal/config"
)

func actionsCfgToPb(cfgActions []config.Action, user *acl.User) *pb.GetDashboardComponentsResponse {
	res := &pb.GetDashboardComponentsResponse{}

	for _, action := range cfgActions {
		if !acl.IsAllowedView(cfg, user, &action) {
			continue
		}

		btn := actionCfgToPb(action, user)
		res.Actions = append(res.Actions, btn)
	}

	return res
}

func actionCfgToPb(action config.Action, user *acl.User) *pb.Action {
	btn := pb.Action{
		Id:      fmt.Sprintf("%x", md5.Sum([]byte(action.Title))),
		Title:   action.Title,
		Icon:    action.Icon,
		CanExec: acl.IsAllowedExec(cfg, user, &action),
	}

	for _, cfgArg := range action.Arguments {
		pbArg := pb.ActionArgument{
			Name:         cfgArg.Name,
			Title:        cfgArg.Title,
			Type:         cfgArg.Type,
			DefaultValue: cfgArg.Default,
			Choices:      buildChoices(cfgArg.Choices),
		}

		btn.Arguments = append(btn.Arguments, &pbArg)
	}

	return &btn
}

func buildChoices(choices []config.ActionArgumentChoice) []*pb.ActionArgumentChoice {
	ret := []*pb.ActionArgumentChoice{}

	for _, cfgChoice := range choices {
		pbChoice := pb.ActionArgumentChoice{
			Value: cfgChoice.Value,
			Title: cfgChoice.Title,
		}

		ret = append(ret, &pbChoice)
	}

	return ret
}
