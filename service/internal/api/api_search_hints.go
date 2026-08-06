package api

import (
	"sort"

	apiv1 "github.com/OliveTin/OliveTin/gen/olivetin/api/v1"
	acl "github.com/OliveTin/OliveTin/internal/acl"
	authpublic "github.com/OliveTin/OliveTin/internal/auth/authpublic"
	"github.com/OliveTin/OliveTin/internal/entities"
	executor "github.com/OliveTin/OliveTin/internal/executor"
	"github.com/OliveTin/OliveTin/internal/tpl"
)

const (
	maxSearchHintActions         = 100
	maxSearchHintEntitiesPerType = 50
)

// buildSearchHints returns lightweight search index data for Init.
// Dashboards are not included; clients index Init.root_dashboard_entries.
// Visibility matches action/entity view ACL; omitted entirely when login is required.
func (api *oliveTinAPI) buildSearchHints(user *authpublic.AuthenticatedUser) *apiv1.SearchHints {
	return &apiv1.SearchHints{
		Entities: api.buildEntitySearchHints(user),
		Actions:  api.buildActionSearchHints(user),
	}
}

func (api *oliveTinAPI) buildEntitySearchHints(user *authpublic.AuthenticatedUser) []*apiv1.EntitySearchHint {
	hints := entities.ListSearchHints()
	out := make([]*apiv1.EntitySearchHint, 0, len(hints))

	for _, hint := range hints {
		if pb := api.entitySearchHintIfAllowed(user, hint); pb != nil {
			out = append(out, pb)
		}
	}

	sortEntitySearchHints(out)
	return capEntitySearchHintsPerType(out, maxSearchHintEntitiesPerType)
}

func (api *oliveTinAPI) entitySearchHintIfAllowed(user *authpublic.AuthenticatedUser, hint entities.SearchHint) *apiv1.EntitySearchHint {
	if hint.UniqueKey == "" || hint.Type == "" {
		return nil
	}

	if !api.userCanViewEntityType(user, hint.Type) {
		return nil
	}

	return &apiv1.EntitySearchHint{
		Title:     hint.Title,
		Type:      hint.Type,
		UniqueKey: hint.UniqueKey,
	}
}

func sortEntitySearchHints(hints []*apiv1.EntitySearchHint) {
	sort.SliceStable(hints, func(i, j int) bool {
		if hints[i].Type != hints[j].Type {
			return hints[i].Type < hints[j].Type
		}

		if hints[i].UniqueKey != hints[j].UniqueKey {
			return hints[i].UniqueKey < hints[j].UniqueKey
		}

		return hints[i].Title < hints[j].Title
	})
}

func capEntitySearchHintsPerType(hints []*apiv1.EntitySearchHint, perType int) []*apiv1.EntitySearchHint {
	if perType < 1 || len(hints) == 0 {
		return hints
	}

	counts := make(map[string]int)
	out := make([]*apiv1.EntitySearchHint, 0, len(hints))

	for _, hint := range hints {
		if counts[hint.Type] >= perType {
			continue
		}

		counts[hint.Type]++
		out = append(out, hint)
	}

	return out
}

func (api *oliveTinAPI) buildActionSearchHints(user *authpublic.AuthenticatedUser) []*apiv1.ActionSearchHint {
	candidates := api.collectViewableActionBindings(user)
	sortActionSearchCandidates(candidates)

	if len(candidates) > maxSearchHintActions {
		candidates = candidates[:maxSearchHintActions]
	}

	out := make([]*apiv1.ActionSearchHint, 0, len(candidates))
	for _, candidate := range candidates {
		out = append(out, &apiv1.ActionSearchHint{
			Title:     candidate.title,
			BindingId: candidate.bindingID,
		})
	}

	return out
}

type actionSearchCandidate struct {
	title     string
	bindingID string
	hasEntity bool
}

func (api *oliveTinAPI) collectViewableActionBindings(user *authpublic.AuthenticatedUser) []actionSearchCandidate {
	api.executor.MapActionBindingsLock.RLock()
	defer api.executor.MapActionBindingsLock.RUnlock()

	candidates := make([]actionSearchCandidate, 0)
	for _, binding := range api.executor.MapActionBindings {
		if candidate, ok := actionSearchCandidateFromBinding(api, user, binding); ok {
			candidates = append(candidates, candidate)
		}
	}

	return candidates
}

func actionSearchCandidateFromBinding(api *oliveTinAPI, user *authpublic.AuthenticatedUser, binding *executor.ActionBinding) (actionSearchCandidate, bool) {
	if !isSearchableActionBinding(binding) {
		return actionSearchCandidate{}, false
	}

	if !acl.IsAllowedView(api.cfg, user, binding.Action) {
		return actionSearchCandidate{}, false
	}

	if !api.bindingEntityTypeAllowed(user, binding) {
		return actionSearchCandidate{}, false
	}

	return actionSearchCandidate{
		title:     tpl.ParseTemplateOfActionBeforeExec(binding.Action.Title, binding.Entity),
		bindingID: binding.ID,
		hasEntity: binding.Entity != nil,
	}, true
}

func isSearchableActionBinding(binding *executor.ActionBinding) bool {
	return binding != nil && binding.Action != nil && binding.ID != "" && !binding.Action.Hidden
}

func sortActionSearchCandidates(candidates []actionSearchCandidate) {
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].hasEntity != candidates[j].hasEntity {
			return !candidates[i].hasEntity
		}

		if candidates[i].title != candidates[j].title {
			return candidates[i].title < candidates[j].title
		}

		return candidates[i].bindingID < candidates[j].bindingID
	})
}
