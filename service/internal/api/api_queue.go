package api

import (
	ctx "context"
	"sort"

	"connectrpc.com/connect"
	apiv1 "github.com/OliveTin/OliveTin/gen/olivetin/api/v1"
	"github.com/OliveTin/OliveTin/internal/auth"
	authpublic "github.com/OliveTin/OliveTin/internal/auth/authpublic"
	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/executor"
)

const defaultActionGroupName = "default"

type executionQueueBucketKey struct {
	groupName string
	bindingID string
}

func (api *oliveTinAPI) GetExecutionQueue(ctx ctx.Context, req *connect.Request[apiv1.GetExecutionQueueRequest]) (*connect.Response[apiv1.GetExecutionQueueResponse], error) {
	user := auth.UserFromApiCall(ctx, req, api.cfg)

	if err := api.checkDashboardAccess(user); err != nil {
		return nil, err
	}

	active := api.executor.GetActiveExecutionsACL(api.cfg, user)
	groups := buildExecutionQueueGroups(active, user, api)

	return connect.NewResponse(&apiv1.GetExecutionQueueResponse{
		Groups:      groups,
		TotalActive: int32(len(active)),
	}), nil
}

func buildExecutionQueueGroups(active []*executor.InternalLogEntry, user *authpublic.AuthenticatedUser, api *oliveTinAPI) []*apiv1.ExecutionQueueGroup {
	actionBuckets := make(map[executionQueueBucketKey]*apiv1.ExecutionQueueAction)

	for _, entry := range active {
		addActiveEntryToActionBuckets(actionBuckets, entry, api.cfg, user, api)
	}

	return buildExecutionQueueGroupsFromBuckets(actionBuckets, api.cfg)
}

func addActiveEntryToActionBuckets(
	buckets map[executionQueueBucketKey]*apiv1.ExecutionQueueAction,
	entry *executor.InternalLogEntry,
	cfg *config.Config,
	user *authpublic.AuthenticatedUser,
	api *oliveTinAPI,
) {
	for _, groupName := range enforcedActionGroupNames(entry, cfg) {
		key := executionQueueBucketKey{
			groupName: groupName,
			bindingID: entry.GetBindingId(),
		}

		action := buckets[key]
		if action == nil {
			action = newExecutionQueueAction(entry)
			buckets[key] = action
		}

		action.Entries = append(action.Entries, api.internalLogEntryToPb(entry, user))
	}
}

func finalizeExecutionQueueGroup(group *apiv1.ExecutionQueueGroup) {
	sortExecutionQueueActions(group.Actions)
	group.ActiveCount = sumExecutionQueueActionEntries(group.Actions)
	group.QueuedCount = countQueuedGroupEntries(group.Actions)
}

func buildExecutionQueueGroupsFromBuckets(
	buckets map[executionQueueBucketKey]*apiv1.ExecutionQueueAction,
	cfg *config.Config,
) []*apiv1.ExecutionQueueGroup {
	grouped := make(map[string]*apiv1.ExecutionQueueGroup)

	for key, action := range buckets {
		sortQueueEntries(action.Entries)
		action.ActiveCount = int32(len(action.Entries))

		group := grouped[key.groupName]
		if group == nil {
			group = newExecutionQueueGroup(key.groupName, cfg)
			grouped[key.groupName] = group
		}

		group.Actions = append(group.Actions, action)
	}

	groups := make([]*apiv1.ExecutionQueueGroup, 0, len(grouped))
	for _, group := range grouped {
		finalizeExecutionQueueGroup(group)
		groups = append(groups, group)
	}

	sortExecutionQueueGroups(groups)
	return groups
}

func hasExecutionQueueBinding(entry *executor.InternalLogEntry, cfg *config.Config) bool {
	return entry != nil && entry.Binding != nil && entry.Binding.Action != nil && cfg != nil
}

func collectEnforcedActionGroupNames(groups []string, cfg *config.Config) []string {
	names := make([]string, 0, len(groups))
	for _, groupName := range groups {
		if isEnforcedActionGroup(cfg, groupName) {
			names = append(names, groupName)
		}
	}
	return names
}

func enforcedActionGroupNames(entry *executor.InternalLogEntry, cfg *config.Config) []string {
	if !hasExecutionQueueBinding(entry, cfg) {
		return []string{defaultActionGroupName}
	}

	names := collectEnforcedActionGroupNames(entry.Binding.Action.Groups, cfg)
	if len(names) == 0 {
		return []string{defaultActionGroupName}
	}

	return names
}

func isEnforcedActionGroup(cfg *config.Config, groupName string) bool {
	group, found := cfg.ActionGroups[groupName]
	return found && group != nil && group.MaxConcurrent >= 1
}

func newExecutionQueueGroup(name string, cfg *config.Config) *apiv1.ExecutionQueueGroup {
	group := &apiv1.ExecutionQueueGroup{Name: name}
	if name == defaultActionGroupName {
		return group
	}

	actionGroup, found := cfg.ActionGroups[name]
	if !found || actionGroup == nil {
		return group
	}

	group.Icon = actionGroup.Icon
	group.MaxConcurrent = int32(actionGroup.MaxConcurrent)
	group.QueueSize = int32(actionGroup.QueueSize)
	return group
}

func newExecutionQueueAction(entry *executor.InternalLogEntry) *apiv1.ExecutionQueueAction {
	action := &apiv1.ExecutionQueueAction{
		BindingId:    entry.GetBindingId(),
		ActionTitle:  entry.ActionTitle,
		ActionIcon:   entry.ActionIcon,
		EntityPrefix: entry.EntityPrefix,
	}

	if entry.Binding != nil && entry.Binding.Action != nil {
		action.MaxConcurrent = int32(entry.Binding.Action.MaxConcurrent)
	}

	return action
}

func sumExecutionQueueActionEntries(actions []*apiv1.ExecutionQueueAction) int32 {
	var total int32

	for _, action := range actions {
		total += int32(len(action.Entries))
	}

	return total
}

func countQueuedGroupEntries(actions []*apiv1.ExecutionQueueAction) int32 {
	var total int32

	for _, action := range actions {
		for _, entry := range action.Entries {
			if entry.Queued {
				total++
			}
		}
	}

	return total
}

func sortQueueEntries(entries []*apiv1.LogEntry) {
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].DatetimeStarted < entries[j].DatetimeStarted
	})
}

func sortExecutionQueueActions(actions []*apiv1.ExecutionQueueAction) {
	sort.Slice(actions, func(i, j int) bool {
		left := actions[i].ActionTitle
		right := actions[j].ActionTitle
		if left == right {
			return actions[i].EntityPrefix < actions[j].EntityPrefix
		}

		return left < right
	})
}

func sortExecutionQueueGroups(groups []*apiv1.ExecutionQueueGroup) {
	sort.Slice(groups, func(i, j int) bool {
		left := groups[i].Name
		right := groups[j].Name

		if left == defaultActionGroupName {
			return false
		}

		if right == defaultActionGroupName {
			return true
		}

		return left < right
	})
}
