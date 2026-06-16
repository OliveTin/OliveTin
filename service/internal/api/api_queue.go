package api

import (
	ctx "context"
	"sort"

	"connectrpc.com/connect"
	apiv1 "github.com/OliveTin/OliveTin/gen/olivetin/api/v1"
	"github.com/OliveTin/OliveTin/internal/auth"
	authpublic "github.com/OliveTin/OliveTin/internal/auth/authpublic"
	"github.com/OliveTin/OliveTin/internal/executor"
)

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
	grouped := make(map[string]*apiv1.ExecutionQueueGroup)

	for _, entry := range active {
		bindingID := entry.GetBindingId()
		group := grouped[bindingID]
		if group == nil {
			group = newExecutionQueueGroup(entry)
			grouped[bindingID] = group
		}

		group.Entries = append(group.Entries, api.internalLogEntryToPb(entry, user))
	}

	groups := make([]*apiv1.ExecutionQueueGroup, 0, len(grouped))
	for _, group := range grouped {
		sortQueueEntries(group.Entries)
		group.ActiveCount = int32(len(group.Entries))
		groups = append(groups, group)
	}

	sortExecutionQueueGroups(groups)
	return groups
}

func newExecutionQueueGroup(entry *executor.InternalLogEntry) *apiv1.ExecutionQueueGroup {
	group := &apiv1.ExecutionQueueGroup{
		BindingId:    entry.GetBindingId(),
		ActionTitle:  entry.ActionTitle,
		ActionIcon:   entry.ActionIcon,
		EntityPrefix: entry.EntityPrefix,
	}

	if entry.Binding != nil && entry.Binding.Action != nil {
		group.MaxConcurrent = int32(entry.Binding.Action.MaxConcurrent)
	}

	return group
}

func sortQueueEntries(entries []*apiv1.LogEntry) {
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].DatetimeStarted < entries[j].DatetimeStarted
	})
}

func sortExecutionQueueGroups(groups []*apiv1.ExecutionQueueGroup) {
	sort.Slice(groups, func(i, j int) bool {
		left := groups[i].ActionTitle
		right := groups[j].ActionTitle
		if left == right {
			return groups[i].EntityPrefix < groups[j].EntityPrefix
		}
		return left < right
	})
}
