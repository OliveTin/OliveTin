package executor

import (
	"slices"
	"time"

	config "github.com/OliveTin/OliveTin/internal/config"
)

// The in-memory log history is capped so that requests cannot grow it without
// limit. The oldest finished entries are evicted first. Entries are never
// evicted while they are still running or queued, or while they still count
// towards their action's maxRate window, because evicting those would let
// unrelated requests reset an action's rate limit. Logs saved to disk are not
// affected.

func (e *Executor) takeNextLogIndexLocked() int64 {
	index := e.nextLogIndex
	e.nextLogIndex++
	return index
}

func (e *Executor) logHistoryLimit() int {
	if e.Cfg == nil || e.Cfg.LogHistoryLimit <= 0 {
		return config.DefaultLogHistoryLimit
	}

	return e.Cfg.LogHistoryLimit
}

// evictOldLogsLocked must be called with logmutex held for writing.
func (e *Executor) evictOldLogsLocked(now time.Time) {
	excess := len(e.logsTrackingIdsByDate) - e.logHistoryLimit()
	if excess <= 0 {
		return
	}

	kept := make([]string, 0, len(e.logsTrackingIdsByDate)-excess)

	for _, trackingID := range e.logsTrackingIdsByDate {
		if excess > 0 && canEvictLogEntry(e.logs[trackingID], now) {
			e.removeLogLocked(trackingID)
			excess--
			continue
		}

		kept = append(kept, trackingID)
	}

	e.logsTrackingIdsByDate = kept
}

func canEvictLogEntry(entry *InternalLogEntry, now time.Time) bool {
	if entry == nil {
		return true
	}

	if !entry.ExecutionFinished {
		return false
	}

	return !countsTowardsRateLimit(entry, now)
}

func countsTowardsRateLimit(entry *InternalLogEntry, now time.Time) bool {
	if entry.Binding == nil || entry.Binding.Action == nil {
		return false
	}

	window := longestRateWindow(entry.Binding.Action.MaxRate)
	if window <= 0 {
		return false
	}

	return shouldCountExecution(entry, now.Add(-window))
}

func longestRateWindow(rates []config.RateSpec) time.Duration {
	longest := time.Duration(0)

	for _, rate := range rates {
		longest = max(longest, parseDuration(rate))
	}

	return longest
}

func (e *Executor) removeLogLocked(trackingID string) {
	entry := e.logs[trackingID]
	delete(e.logs, trackingID)

	if entry == nil || entry.Binding == nil {
		return
	}

	e.removeLogFromBindingLocked(entry)
}

func (e *Executor) removeLogFromBindingLocked(entry *InternalLogEntry) {
	bindingID := entry.Binding.ID
	bindingLogs := e.LogsByBindingId[bindingID]

	index := slices.Index(bindingLogs, entry)
	if index < 0 {
		return
	}

	if len(bindingLogs) == 1 {
		delete(e.LogsByBindingId, bindingID)
		return
	}

	// Build a new slice rather than deleting in place, because callers may still
	// be reading the previous slice after releasing the lock.
	e.LogsByBindingId[bindingID] = slices.Concat(bindingLogs[:index], bindingLogs[index+1:])
}
