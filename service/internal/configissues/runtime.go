package configissues

import (
	"context"

	health "github.com/jamesread/armature-health/v2"
	log "github.com/sirupsen/logrus"
)

// ReportRuntime records an issue discovered while OliveTin is running, rather
// than from the configuration (for example, a call to a deprecated API
// method). Unlike Report, runtime issues are kept when the issue list is
// rebuilt on config reload or entity changes. They are cleared on restart.
func ReportRuntime(issue Issue) {
	raise(sourceRuntime, issue)
}

// CopyRuntime returns a copy of the runtime issues.
func CopyRuntime() []Issue {
	stored, err := service.ListIssues(context.Background(), false, true)
	if err != nil {
		log.WithError(err).Error("Could not list runtime issues")
		return []Issue{}
	}

	runtime := make([]health.IssueView, 0)
	for _, view := range stored {
		if view.CheckID == sourceRuntime {
			runtime = append(runtime, view)
		}
	}
	return orderedIssues(runtime)
}

// ClearRuntime removes all runtime issues. Intended for tests.
func ClearRuntime() {
	setSource(sourceRuntime, nil)
}
