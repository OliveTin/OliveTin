package configissues

import (
	"context"
	"sync"
	"time"

	health "github.com/jamesread/armature-health/v2"
	log "github.com/sirupsen/logrus"
)

// Issues are kept by armature-health in memory, grouped by where they come from.
// A rebuild replaces the "config" and "config.load" sources, and leaves "runtime"
// alone, so issues found while OliveTin is running survive config reloads.
const (
	sourceConfig     = "config"
	sourceConfigLoad = "config.load"
	sourceRuntime    = "runtime"
)

var sourceOrder = []string{sourceConfigLoad, sourceRuntime, sourceConfig}

// OliveTin only shows open issues, so resolved ones are not kept for long.
const resolvedRetention = time.Hour

var (
	service = health.NewService(nil, nil, nil, health.Options{
		SourceTitles: map[string]string{
			sourceConfig:     "Configuration",
			sourceConfigLoad: "Configuration loading",
			sourceRuntime:    "Runtime",
		},
		ResolvedRetention: resolvedRetention,
	})

	stickyMu sync.Mutex
	sticky   []Issue

	// writeMu makes each change and the logging of the issues it adds one step,
	// so concurrent reports cannot log the same new issue twice.
	writeMu sync.Mutex
)

// BeginConfigLoad clears sticky issues captured during config decode/include load.
func BeginConfigLoad() {
	stickyMu.Lock()
	defer stickyMu.Unlock()
	sticky = nil
}

// ReportSticky records an issue that cannot be re-derived after config decode
// (for example unset environment variables that were already expanded).
func ReportSticky(issue Issue) {
	stickyMu.Lock()
	defer stickyMu.Unlock()
	sticky = append(sticky, issue)
}

// CopySticky returns a copy of sticky load-time issues.
func CopySticky() []Issue {
	stickyMu.Lock()
	defer stickyMu.Unlock()
	out := make([]Issue, len(sticky))
	copy(out, sticky)
	return out
}

// Replace sets the issues derived from the current configuration. Issues that
// were already present are kept (and not logged again), and missing ones are
// resolved.
func Replace(next []Issue) {
	setSource(sourceConfig, next)
}

// ReplaceLoadIssues sets the issues found while loading configuration.
func ReplaceLoadIssues(next []Issue) {
	setSource(sourceConfigLoad, next)
}

// Report adds a live issue if it is not already present and logs it when new.
// Use this for runtime discoveries that happen outside Rebuild (for example
// filesystem watcher setup failures). The next Rebuild resolves it unless the
// rebuild finds it again.
func Report(issue Issue) {
	raise(sourceConfig, issue)
}

// List returns the current issues.
func List() []Issue {
	stored, err := service.ListIssues(context.Background(), false, true)
	if err != nil {
		log.WithError(err).Error("Could not list configuration issues")
		return []Issue{}
	}

	return orderedIssues(stored)
}

// Count returns the number of current issues.
func Count() int {
	return len(List())
}

func setSource(sourceID string, next []Issue) {
	writeMu.Lock()
	defer writeMu.Unlock()

	before := openSeverities()
	rememberOrder(sourceID, next)

	if err := service.SetIssues(context.Background(), sourceID, toHealthIssues(next)); err != nil {
		log.WithError(err).WithField("source", sourceID).Error("Could not save configuration issues")
		return
	}

	logNewIssues(before, next)
}

func raise(sourceID string, issue Issue) {
	writeMu.Lock()
	defer writeMu.Unlock()

	before := openSeverities()

	if err := service.Raise(context.Background(), sourceID, toHealthIssue(issue)); err != nil {
		log.WithError(err).WithField("source", sourceID).Error("Could not save configuration issue")
		return
	}

	logNewIssues(before, []Issue{issue})
}

// openSeverities returns the severity of every open issue, by key, across all sources.
func openSeverities() map[string]string {
	stored, err := service.ListIssues(context.Background(), false, true)
	if err != nil {
		log.WithError(err).Error("Could not list configuration issues")
		return map[string]string{}
	}

	severities := make(map[string]string, len(stored))
	for _, view := range stored {
		severities[view.IssueKey] = view.Severity
	}
	return severities
}

// logNewIssues logs issues that were not open before, or whose severity changed,
// in the order they were collected, so startup and reloads do not repeat the
// same warnings.
func logNewIssues(before map[string]string, issues []Issue) {
	for _, issue := range issues {
		key := issueKey(issue)
		severity := string(healthSeverity(issue.Severity))
		if before[key] == severity {
			continue
		}
		before[key] = severity
		logNewIssue(issue)
	}
}

func logNewIssue(issue Issue) {
	fields := log.Fields{
		"code":         issue.Code,
		"actionTitle":  issue.ActionTitle,
		"argumentName": issue.ArgumentName,
		"configFile":   issue.ConfigFile,
		"source":       issue.Source,
	}

	if issue.Severity == SeverityError {
		log.WithFields(fields).Error(issue.Message)
		return
	}
	log.WithFields(fields).Warn(issue.Message)
}
