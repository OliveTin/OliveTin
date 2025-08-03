package api

import (
	apiv1 "github.com/OliveTin/OliveTin/gen/olivetin/api/v1"
	acl "github.com/OliveTin/OliveTin/internal/acl"
	config "github.com/OliveTin/OliveTin/internal/config"
	executor "github.com/OliveTin/OliveTin/internal/executor"
	sv "github.com/OliveTin/OliveTin/internal/stringvariables"
	log "github.com/sirupsen/logrus"
)

type DashboardRenderRequest struct {
	AuthenticatedUser   *acl.AuthenticatedUser
	AllowedActionTitles []string `json:"allows_action_titles"`
	cfg                 *config.Config
	ex                  *executor.Executor

	usedActions map[string]bool
}

func (rr *DashboardRenderRequest) findAction(title string) *apiv1.Action {
	for _, action := range rr.cfg.Actions {
		log.Infof("Checking action %s against %s", title, action.Title)
		if action.Title == title {
			return buildAction(action.ID, nil, rr)
		}
	}

	return nil
}

func buildDashboardResponse(ex *executor.Executor, cfg *config.Config, user *acl.AuthenticatedUser) *apiv1.GetDashboardComponentsResponse {
	res := &apiv1.GetDashboardComponentsResponse{
		AuthenticatedUser:         user.Username,
		AuthenticatedUserProvider: user.Provider,
	}

	/*
		sort.Slice(res.Actions, func(i, j int) bool {
			if res.Actions[i].Order == res.Actions[j].Order {
				return res.Actions[i].Title < res.Actions[j].Title
			} else {
				return res.Actions[i].Order < res.Actions[j].Order
			}
		})
	*/

	rr := &DashboardRenderRequest{
		AuthenticatedUser: user,
		//		AllowedActionTitles: getActionTitles(res.Actions),
		cfg:         cfg,
		ex:          ex,
		usedActions: make(map[string]bool),
	}

	res.EffectivePolicy = buildEffectivePolicy(user.EffectivePolicy)
	res.Dashboards = dashboardCfgToPb(rr)

	return res
}

func getActionTitles(actions []*apiv1.Action) []string {
	titles := make([]string, 0, len(actions))

	for _, action := range actions {
		titles = append(titles, action.Title)
	}

	return titles
}

func buildEffectivePolicy(policy *config.ConfigurationPolicy) *apiv1.EffectivePolicy {
	ret := &apiv1.EffectivePolicy{
		ShowDiagnostics: policy.ShowDiagnostics,
		ShowLogList:     policy.ShowLogList,
	}

	return ret
}

func buildAction(actionId string, actionBinding *executor.ActionBinding, rr *DashboardRenderRequest) *apiv1.Action {
	action := actionBinding.Action

	btn := apiv1.Action{
		Id:           actionId,
		Title:        sv.ReplaceEntityVars(actionBinding.EntityPrefix, action.Title),
		Icon:         sv.ReplaceEntityVars(actionBinding.EntityPrefix, action.Icon),
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
