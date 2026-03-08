package api

import (
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	apiv1 "github.com/OliveTin/OliveTin/gen/olivetin/api/v1"
	acl "github.com/OliveTin/OliveTin/internal/acl"
	authpublic "github.com/OliveTin/OliveTin/internal/auth/authpublic"
	config "github.com/OliveTin/OliveTin/internal/config"
	entities "github.com/OliveTin/OliveTin/internal/entities"
	executor "github.com/OliveTin/OliveTin/internal/executor"
	"github.com/OliveTin/OliveTin/internal/tpl"
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

func bindingMatchesTitleAndEntity(binding *executor.ActionBinding, title string, entity *entities.Entity) bool {
	return binding != nil && binding.Action != nil && binding.Action.Title == title && matchesEntity(binding, entity)
}

func (rr *DashboardRenderRequest) findActionForEntity(title string, entity *entities.Entity) *apiv1.Action {
	rr.ex.MapActionBindingsLock.RLock()
	defer rr.ex.MapActionBindingsLock.RUnlock()

	for _, binding := range rr.ex.MapActionBindings {
		if !bindingMatchesTitleAndEntity(binding, title, entity) {
			continue
		}
		if !acl.IsAllowedView(rr.cfg, rr.AuthenticatedUser, binding.Action) {
			return nil
		}
		return buildAction(binding, rr)
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
		ShowDiagnostics:   policy.ShowDiagnostics,
		ShowLogList:       policy.ShowLogList,
		ShowVersionNumber: policy.ShowVersionNumber,
	}

	return ret
}

func evaluateEnabledExpression(action *config.Action, entity *entities.Entity) bool {
	if action.EnabledExpression == "" {
		return true
	}

	result := tpl.ParseTemplateOfActionBeforeExec(action.EnabledExpression, entity)
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

func getDefaultArgumentValue(cfgArg config.ActionArgument, entity *entities.Entity) string {
	defaultValue := cfgArg.Default

	if defaultValue != "" {
		defaultValue = tpl.ParseTemplateOfActionBeforeExec(defaultValue, entity)
	}

	return defaultValue
}

func formatRateLimitExpiry(expiryUnix int64) string {
	if expiryUnix <= 0 {
		return ""
	}
	return time.Unix(expiryUnix, 0).Format("2006-01-02 15:04:05")
}

func actionFromBinding(actionBinding *executor.ActionBinding) (*executor.ActionBinding, *config.Action) {
	if actionBinding == nil || actionBinding.Action == nil {
		return nil, nil
	}
	return actionBinding, actionBinding.Action
}

func buildAction(actionBinding *executor.ActionBinding, rr *DashboardRenderRequest) *apiv1.Action {
	binding, action := actionFromBinding(actionBinding)
	if binding == nil {
		return nil
	}

	aclCanExec := acl.IsAllowedExec(rr.cfg, rr.AuthenticatedUser, action)
	enabledExprCanExec := evaluateEnabledExpression(action, binding.Entity)
	datetimeRateLimitExpires := formatRateLimitExpiry(rr.ex.GetTimeUntilAvailable(binding))

	btn := apiv1.Action{
		BindingId:                binding.ID,
		Title:                    tpl.ParseTemplateOfActionBeforeExec(action.Title, binding.Entity),
		Icon:                     tpl.ParseTemplateOfActionBeforeExec(action.Icon, binding.Entity),
		CanExec:                  aclCanExec && enabledExprCanExec,
		PopupOnStart:             action.PopupOnStart,
		Order:                    int32(binding.ConfigOrder),
		Timeout:                  int32(action.Timeout),
		DatetimeRateLimitExpires: datetimeRateLimitExpires,
	}

	for _, cfgArg := range action.Arguments {
		pbArg := apiv1.ActionArgument{
			Name:                  cfgArg.Name,
			Title:                 cfgArg.Title,
			Type:                  cfgArg.Type,
			Description:           cfgArg.Description,
			DefaultValue:          getDefaultArgumentValue(cfgArg, binding.Entity),
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

	for _, ent := range entities.GetEntityInstancesOrdered(entityTitle) {
		ret = append(ret, &apiv1.ActionArgumentChoice{
			Value: tpl.ParseTemplateOfActionBeforeExec(firstChoice.Value, ent),
			Title: tpl.ParseTemplateOfActionBeforeExec(firstChoice.Title, ent),
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
