package grpcapi

import (
	pb "github.com/OliveTin/OliveTin/gen/grpc"
	acl "github.com/OliveTin/OliveTin/internal/acl"
	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/entities"
	executor "github.com/OliveTin/OliveTin/internal/executor"
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
		if res.Actions[i].Order == res.Actions[j].Order {
			return res.Actions[i].Title < res.Actions[j].Title
		} else {
			return res.Actions[i].Order < res.Actions[j].Order
		}
	})

	return res
}

func isAllowedExec(cfg *config.Config, user *acl.AuthenticatedUser, action *config.Action, actionBinding *executor.ActionBinding) bool {
	hasPermission := acl.IsAllowedExec(cfg, user, action)

	isEnabled := entities.ParseTemplateBoolWith(action.Enabled, actionBinding.Entity)

	return hasPermission && isEnabled
}

func buildAction(actionId string, actionBinding *executor.ActionBinding, user *acl.AuthenticatedUser) *pb.Action {
	action := actionBinding.Action

	btn := pb.Action{
		Id:           actionId,
		Title:        entities.ParseTemplateWith(action.Title, actionBinding.Entity),
		Icon:         action.Icon,
		CanExec:      isAllowedExec(cfg, user, action, actionBinding),
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

	entList := entities.GetEntities(entityTitle)

	for _, ent := range entList {
		ret = append(ret, &pb.ActionArgumentChoice{
			Value: entities.ParseTemplateWith(firstChoice.Value, ent),
			Title: entities.ParseTemplateWithArgs(firstChoice.Title, ent, nil),
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
