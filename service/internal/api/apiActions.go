package api

import (
	apiv1 "github.com/OliveTin/OliveTin/gen/olivetin/api/v1"
	acl "github.com/OliveTin/OliveTin/internal/acl"
	config "github.com/OliveTin/OliveTin/internal/config"
	entities "github.com/OliveTin/OliveTin/internal/entities"
	executor "github.com/OliveTin/OliveTin/internal/executor"
)

type DashboardRenderRequest struct {
	AuthenticatedUser *acl.AuthenticatedUser
	cfg               *config.Config
	ex                *executor.Executor
}

func (rr *DashboardRenderRequest) findAction(title string) *apiv1.Action {
	for id, binding := range rr.ex.MapActionIdToBinding {
		if binding.Action.Title == title {
			return buildAction(id, binding, rr)
		}
	}

	return nil
}

func buildDashboardResponse(ex *executor.Executor, cfg *config.Config, user *acl.AuthenticatedUser, dashboardTitle string) *apiv1.GetDashboardResponse {
	res := &apiv1.GetDashboardResponse{}

	rr := &DashboardRenderRequest{
		AuthenticatedUser: user,
		cfg:               cfg,
		ex:                ex,
	}

	res.Dashboard = dashboardCfgToPb(rr, dashboardTitle)

	return res
}

func buildEffectivePolicy(policy *config.ConfigurationPolicy) *apiv1.EffectivePolicy {
	ret := &apiv1.EffectivePolicy{
		ShowDiagnostics: policy.ShowDiagnostics,
		ShowLogList:     policy.ShowLogList,
	}

	return ret
}

func buildAction(bindingId string, actionBinding *executor.ActionBinding, rr *DashboardRenderRequest) *apiv1.Action {
	action := actionBinding.Action

	btn := apiv1.Action{
		BindingId:    bindingId,
		Title:        entities.ParseTemplateWith(action.Title, actionBinding.Entity),
		Icon:         entities.ParseTemplateWith(action.Icon, actionBinding.Entity),
		CanExec:      acl.IsAllowedExec(rr.cfg, rr.AuthenticatedUser, action),
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

	entList := entities.GetEntityInstances(entityTitle)

	for _, ent := range entList {
		ret = append(ret, &apiv1.ActionArgumentChoice{
			Value: entities.ParseTemplateWith(firstChoice.Value, ent),
			Title: entities.ParseTemplateWith(firstChoice.Title, ent),
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
