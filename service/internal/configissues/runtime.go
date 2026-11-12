package configissues

import "sync"

var (
	runtimeMu     sync.Mutex
	runtimeIssues []Issue
)

// ReportRuntime records an issue discovered while OliveTin is running, rather
// than from the configuration (for example, a call to a deprecated API
// method). Unlike Report, runtime issues are kept when the issue list is
// rebuilt on config reload or entity changes. They are cleared on restart.
func ReportRuntime(issue Issue) {
	addRuntime(issue)
	Report(issue)
}

func addRuntime(issue Issue) {
	runtimeMu.Lock()
	defer runtimeMu.Unlock()

	fingerprint := issueFingerprint(issue)
	for _, existing := range runtimeIssues {
		if issueFingerprint(existing) == fingerprint {
			return
		}
	}

	runtimeIssues = append(runtimeIssues, issue)
}

// CopyRuntime returns a copy of the runtime issues.
func CopyRuntime() []Issue {
	runtimeMu.Lock()
	defer runtimeMu.Unlock()
	out := make([]Issue, len(runtimeIssues))
	copy(out, runtimeIssues)
	return out
}

// ClearRuntime removes all runtime issues. Intended for tests.
func ClearRuntime() {
	runtimeMu.Lock()
	defer runtimeMu.Unlock()
	runtimeIssues = nil
}
