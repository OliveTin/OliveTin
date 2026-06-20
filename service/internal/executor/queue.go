package executor

import (
	acl "github.com/OliveTin/OliveTin/internal/acl"
	authpublic "github.com/OliveTin/OliveTin/internal/auth/authpublic"
	config "github.com/OliveTin/OliveTin/internal/config"
)

func isActiveQueueEntry(entry *InternalLogEntry) bool {
	return entry != nil && !entry.ExecutionFinished
}

func isQueueEntryVisible(cfg *config.Config, user *authpublic.AuthenticatedUser, entry *InternalLogEntry) bool {
	if !isActiveQueueEntry(entry) || !isValidLogEntryForACL(entry) {
		return false
	}

	return acl.IsAllowedLogs(cfg, user, entry.Binding.Action)
}

// GetActiveExecutionsACL returns unfinished executions the user may view in the queue.
func (e *Executor) GetActiveExecutionsACL(cfg *config.Config, user *authpublic.AuthenticatedUser) []*InternalLogEntry {
	e.logmutex.RLock()
	defer e.logmutex.RUnlock()

	active := make([]*InternalLogEntry, 0)

	for _, trackingID := range e.logsTrackingIdsByDate {
		entry := e.logs[trackingID]
		if isQueueEntryVisible(cfg, user, entry) {
			active = append(active, entry)
		}
	}

	return active
}
