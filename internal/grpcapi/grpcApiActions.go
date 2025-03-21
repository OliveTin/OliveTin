package grpcapi

import (
	apiv1 "github.com/OliveTin/OliveTin/gen/grpc/olivetin/api/v1"
	acl "github.com/OliveTin/OliveTin/internal/acl"
	config "github.com/OliveTin/OliveTin/internal/config"
	executor "github.com/OliveTin/OliveTin/internal/executor"
	sv "github.com/OliveTin/OliveTin/internal/stringvariables"
	"sort"
)

func buildDashboardResponse(ex *executor.Executor, cfg *config.Config, user *acl.AuthenticatedUser) *apiv1.GetDashboardComponentsResponse {
	res := &apiv1.GetDashboardComponentsResponse{
		AuthenticatedUser:         user.Username,
		AuthenticatedUserProvider: user.Provider,
	}

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

func buildAction(actionId string, actionBinding *executor.ActionBinding, user *acl.AuthenticatedUser) *apiv1.Action {
	action := actionBinding.Action

	btn := apiv1.Action{
		Id:           actionId,
		Title:        sv.ReplaceEntityVars(actionBinding.EntityPrefix, action.Title),
		Icon:         action.Icon,
		CanExec:      acl.IsAllowedExec(cfg, user, action),
		PopupOnStart: action.PopupOnStart,
		Order:        int32(actionBinding.ConfigOrder),
	}

	for _, cfgArg := range action.Arguments {
		pbArg := apiv1.ActionArgument{
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

func buildChoices(arg config.ActionArgument) []*apiv1.ActionArgumentChoice {
	if arg.Entity != "" && len(arg.Choices) == 1 {
		return buildChoicesEntity(arg.Choices[0], arg.Entity)
	} else {
		return buildChoicesSimple(arg.Choices)
	}
}

func buildChoicesEntity(firstChoice config.ActionArgumentChoice, entityTitle string) []*apiv1.ActionArgumentChoice {
	ret := []*apiv1.ActionArgumentChoice{}

	entityCount := sv.GetEntityCount(entityTitle)

	for i := 0; i < entityCount; i++ {
		prefix := sv.GetEntityPrefix(entityTitle, i)

		ret = append(ret, &apiv1.ActionArgumentChoice{
			Value: sv.ReplaceEntityVars(prefix, firstChoice.Value),
			Title: sv.ReplaceEntityVars(prefix, firstChoice.Title),
		})
	}

	return ret
}

func buildChoicesSimple(choices []config.ActionArgumentChoice) []*apiv1.ActionArgumentChoice {
	ret := []*apiv1.ActionArgumentChoice{}

	for _, cfgChoice := range choices {
		pbChoice := apiv1.ActionArgumentChoice{
			Value: cfgChoice.Value,
			Title: cfgChoice.Title,
		}

		ret = append(ret, &pbChoice)
	}

	return ret
}
