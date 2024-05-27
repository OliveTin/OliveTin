package grpcapi

import (
	pb "github.com/OliveTin/OliveTin/gen/grpc"
	acl "github.com/OliveTin/OliveTin/internal/acl"
	config "github.com/OliveTin/OliveTin/internal/config"
	executor "github.com/OliveTin/OliveTin/internal/executor"
	sv "github.com/OliveTin/OliveTin/internal/stringvariables"
	"sort"
)

func buildDashboardResponse(ex *executor.Executor, cfg *config.Config, user *acl.AuthenticatedUser) *pb.GetDashboardComponentsResponse {
	res := &pb.GetDashboardComponentsResponse{}

	ex.MapActionIdToBindingLock.RLock()

	for actionId, actionBinding := range ex.MapActionIdToBinding {
		if !acl.IsAllowedView(cfg, user, actionBinding.Action) {
			continue
		}

		res.Actions = append(res.Actions, buildAction(actionId, actionBinding, user))
	}

	ex.MapActionIdToBindingLock.RUnlock()

	sort.Slice(res.Actions, func(i, j int) bool {
		return res.Actions[i].Order < res.Actions[j].Order

	})

	return res
}

func buildAction(actionId string, actionBinding *executor.ActionBinding, user *acl.AuthenticatedUser) *pb.Action {
	action := actionBinding.Action

	btn := pb.Action{
		Id:           actionId,
		Title:        sv.ReplaceEntityVars(actionBinding.EntityPrefix, action.Title),
		Icon:         action.Icon,
		CanExec:      acl.IsAllowedExec(cfg, user, action),
		PopupOnStart: action.PopupOnStart,
		Order:        int32(actionBinding.ConfigOrder),
	}

	for _, cfgArg := range action.Arguments {
		pbArg := pb.ActionArgument{
			Name:         cfgArg.Name,
			Title:        cfgArg.Title,
			Type:         cfgArg.Type,
			Description:  cfgArg.Description,
			DefaultValue: cfgArg.Default,
			Choices:      buildChoices(cfgArg),
			Suggestions:  cfgArg.Suggestions,
		}

		btn.Arguments = append(btn.Arguments, &pbArg)
	}

	return &btn
}

func buildChoices(arg config.ActionArgument) []*pb.ActionArgumentChoice {
	if arg.Entity != "" && len(arg.Choices) == 1 {
		return buildChoicesEntity(arg.Choices[0], arg.Entity)
	} else {
		return buildChoicesSimple(arg.Choices)
	}
}

func buildChoicesEntity(firstChoice config.ActionArgumentChoice, entityTitle string) []*pb.ActionArgumentChoice {
	ret := []*pb.ActionArgumentChoice{}

	entityCount := sv.GetEntityCount(entityTitle)

	for i := 0; i < entityCount; i++ {
		prefix := sv.GetEntityPrefix(entityTitle, i)

		ret = append(ret, &pb.ActionArgumentChoice{
			Value: sv.ReplaceEntityVars(prefix, firstChoice.Value),
			Title: sv.ReplaceEntityVars(prefix, firstChoice.Title),
		})
	}

	return ret
}

func buildChoicesSimple(choices []config.ActionArgumentChoice) []*pb.ActionArgumentChoice {
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
