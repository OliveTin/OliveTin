package api

import (
	apiv1 "github.com/OliveTin/OliveTin/gen/olivetin/api/v1"
	authpublic "github.com/OliveTin/OliveTin/internal/auth/authpublic"
	"github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/configissues"
)

func configIssueCountForUser(api *oliveTinAPI, user *authpublic.AuthenticatedUser) int32 {
	return int32(len(api.buildConfigIssuesForUser(user)))
}

func (api *oliveTinAPI) buildConfigIssuesForUser(user *authpublic.AuthenticatedUser) []*apiv1.ConfigIssue {
	if !userCanSeeDiagnostics(user) {
		return []*apiv1.ConfigIssue{}
	}

	return filterConfigIssuesForUser(api, user, configissues.List())
}

func userCanSeeDiagnostics(user *authpublic.AuthenticatedUser) bool {
	return user != nil && user.EffectivePolicy != nil && user.EffectivePolicy.ShowDiagnostics
}

func filterConfigIssuesForUser(api *oliveTinAPI, user *authpublic.AuthenticatedUser, src []configissues.Issue) []*apiv1.ConfigIssue {
	out := make([]*apiv1.ConfigIssue, 0, len(src))
	for _, issue := range src {
		if api.includeConfigIssueForUser(user, issue) {
			out = append(out, toProtoConfigIssue(issue))
		}
	}
	return out
}

func (api *oliveTinAPI) includeConfigIssueForUser(user *authpublic.AuthenticatedUser, issue configissues.Issue) bool {
	if issue.ActionID == "" {
		return true
	}
	return api.canViewActionID(user, issue.ActionID)
}

func (api *oliveTinAPI) canViewActionID(user *authpublic.AuthenticatedUser, actionID string) bool {
	action := findActionInList(api.cfg.Actions, actionID)
	if action == nil {
		return false
	}
	return api.userCanViewAction(user, action)
}

func findActionInList(actions []*config.Action, actionID string) *config.Action {
	for _, action := range actions {
		if actionHasID(action, actionID) {
			return action
		}
	}
	return nil
}

func actionHasID(action *config.Action, actionID string) bool {
	return action != nil && action.ID == actionID
}

func toProtoConfigIssue(issue configissues.Issue) *apiv1.ConfigIssue {
	return &apiv1.ConfigIssue{
		Severity:     issue.Severity,
		Code:         issue.Code,
		Message:      issue.Message,
		ActionId:     issue.ActionID,
		ActionTitle:  issue.ActionTitle,
		ArgumentName: issue.ArgumentName,
		Source:       issue.Source,
		ConfigFile:   issue.ConfigFile,
	}
}
