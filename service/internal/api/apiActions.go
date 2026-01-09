package api

import (
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	apiv1 "github.com/OliveTin/OliveTin/gen/olivetin/api/v1"
	acl "github.com/OliveTin/OliveTin/internal/acl"
	authpublic "github.com/OliveTin/OliveTin/internal/auth/authpublic"
	config "github.com/OliveTin/OliveTin/internal/config"
	entities "github.com/OliveTin/OliveTin/internal/entities"
	executor "github.com/OliveTin/OliveTin/internal/executor"
)

type DashboardRenderRequest struct {
	AuthenticatedUser *authpublic.AuthenticatedUser
	cfg               *config.Config
	ex                *executor.Executor
	EntityType        string
	EntityKey         string
}

func (rr *DashboardRenderRequest) findAction(title string) *apiv1.Action {
	return rr.findActionForEntity(title, nil)
}

func (rr *DashboardRenderRequest) findActionForEntity(title string, entity *entities.Entity) *apiv1.Action {
	rr.ex.MapActionIdToBindingLock.RLock()
	defer rr.ex.MapActionIdToBindingLock.RUnlock()

	for _, binding := range rr.ex.MapActionIdToBinding {
		if binding.Action.Title != title {
			continue
		}

		if matchesEntity(binding, entity) {
			return buildAction(binding, rr)
		}
	}

	return nil
}

func matchesEntity(binding *executor.ActionBinding, entity *entities.Entity) bool {
	if entity == nil {
		return binding.Entity == nil
	}

	return binding.Entity != nil && binding.Entity.UniqueKey == entity.UniqueKey
}

func buildEffectivePolicy(policy *config.ConfigurationPolicy) *apiv1.EffectivePolicy {
	ret := &apiv1.EffectivePolicy{
		ShowDiagnostics: policy.ShowDiagnostics,
		ShowLogList:     policy.ShowLogList,
	}

	return ret
}

func evaluateEnabledExpression(action *config.Action, entity *entities.Entity) bool {
	if action.EnabledExpression == "" {
		return true
	}

	result := entities.ParseTemplateWith(action.EnabledExpression, entity)
	result = strings.TrimSpace(result)

	if result == "" {
		return false
	}

	if isTemplateError(result, action) {
		return false
	}

	return evaluateResultValue(result)
}

func isTemplateError(result string, action *config.Action) bool {
	if !strings.HasPrefix(result, "tpl ") || !strings.Contains(result, "error") {
		return false
	}

	log.WithFields(log.Fields{
		"actionTitle":       action.Title,
		"enabledExpression": action.EnabledExpression,
		"result":            result,
	}).Warn("enabledExpression template evaluation failed, treating as disabled")
	return true
}

func evaluateResultValue(result string) bool {
	if strings.EqualFold(result, "true") {
		return true
	}

	if num, err := strconv.Atoi(result); err == nil {
		return num != 0
	}

	return false
}

func buildAction(actionBinding *executor.ActionBinding, rr *DashboardRenderRequest) *apiv1.Action {
	action := actionBinding.Action

	aclCanExec := acl.IsAllowedExec(rr.cfg, rr.AuthenticatedUser, action)
	enabledExprCanExec := evaluateEnabledExpression(action, actionBinding.Entity)

	btn := apiv1.Action{
		BindingId:    actionBinding.ID,
		Title:        entities.ParseTemplateWith(action.Title, actionBinding.Entity),
		Icon:         entities.ParseTemplateWith(action.Icon, actionBinding.Entity),
		CanExec:      aclCanExec && enabledExprCanExec,
		PopupOnStart: action.PopupOnStart,
		Order:        int32(actionBinding.ConfigOrder),
		Timeout:      int32(action.Timeout),
	}

	for _, cfgArg := range action.Arguments {
		pbArg := apiv1.ActionArgument{
			Name:                  cfgArg.Name,
			Title:                 cfgArg.Title,
			Type:                  cfgArg.Type,
			Description:           cfgArg.Description,
			DefaultValue:          cfgArg.Default,
			Choices:               buildChoices(cfgArg),
			Suggestions:           cfgArg.Suggestions,
			SuggestionsBrowserKey: cfgArg.SuggestionsBrowserKey,
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
