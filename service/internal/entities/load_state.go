package entities

import "sync"

var (
	loadAttemptedMu sync.RWMutex
	loadAttempted   = map[string]bool{}
)

// MarkEntityLoadAttempted records that OliveTin has tried to load this entity
// type from its entity file (successfully or not).
func MarkEntityLoadAttempted(entityName string) {
	if entityName == "" {
		return
	}
	loadAttemptedMu.Lock()
	defer loadAttemptedMu.Unlock()
	loadAttempted[entityName] = true
}

// HasEntityLoadAttempted reports whether a load has been attempted for the entity type.
func HasEntityLoadAttempted(entityName string) bool {
	loadAttemptedMu.RLock()
	defer loadAttemptedMu.RUnlock()
	return loadAttempted[entityName]
}

// ResetEntityLoadAttempts clears load-attempt tracking (used by tests).
func ResetEntityLoadAttempts() {
	loadAttemptedMu.Lock()
	defer loadAttemptedMu.Unlock()
	loadAttempted = map[string]bool{}
}
