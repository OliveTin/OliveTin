package api

import (
	"sort"

	apiv1 "github.com/OliveTin/OliveTin/gen/olivetin/api/v1"
	authpublic "github.com/OliveTin/OliveTin/internal/auth/authpublic"
	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/entities"
	"github.com/OliveTin/OliveTin/internal/executor"
	"github.com/OliveTin/OliveTin/internal/tpl"
)

type relatedActionCandidate struct {
	binding   *executor.ActionBinding
	prefilled map[string]string
}

func (api *oliveTinAPI) relatedActionsForEntity(user *authpublic.AuthenticatedUser, entityType string, entity *entities.Entity) []*apiv1.EntityRelatedAction {
	renderRequest := api.createDashboardRenderRequest(user, entityType, entity.UniqueKey)
	populateActiveBindingStates(renderRequest)

	candidates := collectRelatedActionCandidates(api, user, entityType, entity)
	sortRelatedActionCandidates(candidates)

	return buildEntityRelatedActions(candidates, renderRequest)
}

func collectRelatedActionCandidates(api *oliveTinAPI, user *authpublic.AuthenticatedUser, entityType string, entity *entities.Entity) []relatedActionCandidate {
	seen := make(map[string]bool)
	candidates := make([]relatedActionCandidate, 0)

	api.executor.MapActionBindingsLock.RLock()
	defer api.executor.MapActionBindingsLock.RUnlock()

	for _, binding := range api.executor.MapActionBindings {
		tryAppendRelatedCandidate(&candidates, seen, api, user, entityType, entity, binding)
	}

	return candidates
}

func tryAppendRelatedCandidate(candidates *[]relatedActionCandidate, seen map[string]bool, api *oliveTinAPI, user *authpublic.AuthenticatedUser, entityType string, entity *entities.Entity, binding *executor.ActionBinding) {
	prefilled, ok := relatedPrefillForBinding(binding, entityType, entity)
	if !ok || !bindingViewableForRelated(seen, api, user, binding) {
		return
	}

	seen[binding.ID] = true
	*candidates = append(*candidates, relatedActionCandidate{
		binding:   binding,
		prefilled: prefilled,
	})
}

func bindingViewableForRelated(seen map[string]bool, api *oliveTinAPI, user *authpublic.AuthenticatedUser, binding *executor.ActionBinding) bool {
	return binding != nil && binding.Action != nil && !seen[binding.ID] && api.userCanViewAction(user, binding.Action)
}

func relatedPrefillForBinding(binding *executor.ActionBinding, entityType string, entity *entities.Entity) (map[string]string, bool) {
	if isEntityBoundBindingFor(binding, entityType, entity) {
		return nil, true
	}

	return argumentEntityPrefill(binding, entityType, entity)
}

func argumentEntityPrefill(binding *executor.ActionBinding, entityType string, entity *entities.Entity) (map[string]string, bool) {
	if binding == nil || binding.Entity != nil || binding.Action == nil {
		return nil, false
	}

	prefilled := buildPrefilledArgumentsForEntity(binding.Action, entityType, entity)
	return prefilled, len(prefilled) > 0
}

func isEntityBoundBindingFor(binding *executor.ActionBinding, entityType string, entity *entities.Entity) bool {
	if entity == nil || !bindingHasEntity(binding) {
		return false
	}

	return binding.Action.Entity == entityType && binding.Entity.UniqueKey == entity.UniqueKey
}

func bindingHasEntity(binding *executor.ActionBinding) bool {
	return binding != nil && binding.Entity != nil && binding.Action != nil
}

func buildPrefilledArgumentsForEntity(action *config.Action, entityType string, entity *entities.Entity) map[string]string {
	prefilled := make(map[string]string)

	for i := range action.Arguments {
		arg := &action.Arguments[i]
		if arg.Entity != entityType || len(arg.Choices) != 1 {
			continue
		}

		prefilled[arg.Name] = tpl.ParseTemplateOfActionBeforeExec(arg.Choices[0].Value, entity)
	}

	return prefilled
}

func sortRelatedActionCandidates(candidates []relatedActionCandidate) {
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].binding.ConfigOrder != candidates[j].binding.ConfigOrder {
			return candidates[i].binding.ConfigOrder < candidates[j].binding.ConfigOrder
		}

		return candidates[i].binding.ID < candidates[j].binding.ID
	})
}

func buildEntityRelatedActions(candidates []relatedActionCandidate, rr *DashboardRenderRequest) []*apiv1.EntityRelatedAction {
	result := make([]*apiv1.EntityRelatedAction, 0, len(candidates))

	for _, candidate := range candidates {
		action := buildAction(candidate.binding, rr)
		if action == nil {
			continue
		}

		result = append(result, &apiv1.EntityRelatedAction{
			Action:             action,
			PrefilledArguments: candidate.prefilled,
		})
	}

	return result
}
