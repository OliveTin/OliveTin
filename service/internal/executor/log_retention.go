package executor

import (
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
	evicted := make(map[*InternalLogEntry]bool, excess)

	for _, trackingID := range e.logsTrackingIdsByDate {
		if excess > 0 && canEvictLogEntry(e.logs[trackingID], now) {
			e.removeLogLocked(trackingID, evicted)
			excess--
			continue
		}

		kept = append(kept, trackingID)
	}

	e.logsTrackingIdsByDate = kept
	e.removeEvictedFromBindingsLocked(evicted)
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

func (e *Executor) removeLogLocked(trackingID string, evicted map[*InternalLogEntry]bool) {
	entry := e.logs[trackingID]
	delete(e.logs, trackingID)

	if entry == nil || entry.Binding == nil {
		return
	}

	evicted[entry] = true
}

// Each affected binding's list is rebuilt once per eviction, not once per
// evicted entry, so evicting many entries of one binding stays linear.
func (e *Executor) removeEvictedFromBindingsLocked(evicted map[*InternalLogEntry]bool) {
	bindingIDs := make(map[string]bool)
	for entry := range evicted {
		bindingIDs[entry.Binding.ID] = true
	}

	for bindingID := range bindingIDs {
		e.removeEvictedFromBindingLocked(bindingID, evicted)
	}
}

func (e *Executor) removeEvictedFromBindingLocked(bindingID string, evicted map[*InternalLogEntry]bool) {
	// Build a new slice rather than deleting in place, because callers may still
	// be reading the previous slice after releasing the lock.
	remaining := make([]*InternalLogEntry, 0, len(e.LogsByBindingId[bindingID]))
	for _, entry := range e.LogsByBindingId[bindingID] {
		if !evicted[entry] {
			remaining = append(remaining, entry)
		}
	}

	if len(remaining) == 0 {
		delete(e.LogsByBindingId, bindingID)
		return
	}

	e.LogsByBindingId[bindingID] = remaining
}
