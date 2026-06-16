package executor

import (
	"fmt"
	"slices"
	"sync"

	config "github.com/OliveTin/OliveTin/internal/config"
	log "github.com/sirupsen/logrus"
)

type groupLimit struct {
	name          string
	maxConcurrent int
}

type queuedExecution struct {
	req *ExecutionRequest
	wg  *sync.WaitGroup
}

func actionGroupLimits(req *ExecutionRequest) []groupLimit {
	if !hasActionGroupContext(req) {
		return nil
	}

	limits := make([]groupLimit, 0, len(req.Binding.Action.Groups))

	for _, groupName := range req.Binding.Action.Groups {
		if limit, ok := groupLimitFromConfig(req.Cfg, groupName); ok {
			limits = append(limits, limit)
		}
	}

	return limits
}

func hasActionGroupContext(req *ExecutionRequest) bool {
	return req != nil && req.Binding != nil && req.Binding.Action != nil && req.Cfg != nil
}

func groupLimitFromConfig(cfg *config.Config, groupName string) (groupLimit, bool) {
	group, found := cfg.ActionGroups[groupName]
	if !found || group == nil || group.MaxConcurrent < 1 {
		return groupLimit{}, false
	}

	return groupLimit{name: groupName, maxConcurrent: group.MaxConcurrent}, true
}

func actionNeedsGroupLimit(req *ExecutionRequest) bool {
	return len(actionGroupLimits(req)) > 0
}

func actionInGroup(action *config.Action, groupName string) bool {
	if action == nil {
		return false
	}

	return slices.Contains(action.Groups, groupName)
}

func (e *Executor) countActiveInGroup(groupName string) int {
	e.logmutex.RLock()
	defer e.logmutex.RUnlock()

	return e.countActiveInGroupLocked(groupName)
}

func (e *Executor) countActiveInGroupLocked(groupName string) int {
	count := 0

	for _, logEntry := range e.logs {
		if logEntryIsActiveInGroup(logEntry, groupName) {
			count++
		}
	}

	return count
}

func logEntryIsActiveInGroup(logEntry *InternalLogEntry, groupName string) bool {
	if inactiveLogEntry(logEntry) {
		return false
	}

	return actionInGroup(logEntry.Binding.Action, groupName)
}

func inactiveLogEntry(logEntry *InternalLogEntry) bool {
	if logEntry == nil {
		return true
	}

	return logEntryIsInactive(logEntry)
}

func logEntryIsInactive(logEntry *InternalLogEntry) bool {
	if logEntry.ExecutionFinished || logEntry.Queued {
		return true
	}

	return logEntry.Binding == nil || logEntry.Binding.Action == nil
}

func (e *Executor) groupsHaveCapacityForActive(req *ExecutionRequest) bool {
	for _, limit := range actionGroupLimits(req) {
		if e.countActiveInGroup(limit.name) >= (limit.maxConcurrent + 1) {
			return false
		}
	}

	return true
}

func (e *Executor) groupsHaveCapacityForQueued(req *ExecutionRequest) bool {
	for _, limit := range actionGroupLimits(req) {
		if e.countActiveInGroup(limit.name) >= limit.maxConcurrent {
			return false
		}
	}

	return true
}

func firstFullGroupName(e *Executor, req *ExecutionRequest) string {
	for _, limit := range actionGroupLimits(req) {
		if e.countActiveInGroup(limit.name) >= (limit.maxConcurrent + 1) {
			return limit.name
		}
	}

	return ""
}

func firstFullGroupNameLocked(e *Executor, req *ExecutionRequest) string {
	for _, limit := range actionGroupLimits(req) {
		if e.countActiveInGroupLocked(limit.name) >= (limit.maxConcurrent + 1) {
			return limit.name
		}
	}

	return ""
}

func (e *Executor) queueRequest(req *ExecutionRequest, wg *sync.WaitGroup) {
	e.groupQueueMu.Lock()

	var groupName string

	req.mutateLogEntry(func(entry *InternalLogEntry) {
		groupName = firstFullGroupNameLocked(e, req)
		entry.Queued = true
		entry.QueuedForGroup = groupName
		entry.Output = fmt.Sprintf("Queued waiting for action group %q", groupName)
	})

	e.groupQueue = append(e.groupQueue, &queuedExecution{req: req, wg: wg})
	e.groupQueueMu.Unlock()

	e.drainGroupQueue()

	log.WithFields(log.Fields{
		"actionTitle": req.logEntry.ActionTitle,
		"groupName":   groupName,
	}).Infof("Action queued due to action group concurrency limit")
}

func (e *Executor) drainGroupQueue() {
	e.groupQueueMu.Lock()

	if len(e.groupQueue) == 0 {
		e.groupQueueMu.Unlock()
		return
	}

	next := e.groupQueue[0]
	if !e.groupsHaveCapacityForQueued(next.req) {
		e.groupQueueMu.Unlock()
		return
	}

	e.groupQueue = e.groupQueue[1:]

	next.req.mutateLogEntry(func(entry *InternalLogEntry) {
		entry.Queued = false
		entry.QueuedForGroup = ""
	})

	e.groupQueueMu.Unlock()

	go e.runDequeuedExecution(next)
}

func (e *Executor) runDequeuedExecution(queued *queuedExecution) {
	req := queued.req

	req.skipRequestRegistration = true

	e.runExecutionSteps(req)
	e.finishExecChain(req)
	queued.wg.Done()
}
