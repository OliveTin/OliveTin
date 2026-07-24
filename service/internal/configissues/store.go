package configissues

import (
	"sync"

	log "github.com/sirupsen/logrus"
)

var (
	storeMu sync.RWMutex
	issues  []Issue

	stickyMu sync.Mutex
	sticky   []Issue
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

func copySticky() []Issue {
	stickyMu.Lock()
	defer stickyMu.Unlock()
	out := make([]Issue, len(sticky))
	copy(out, sticky)
	return out
}

// CopySticky returns a copy of sticky load-time issues.
func CopySticky() []Issue {
	return copySticky()
}

// Replace sets the current issue list and logs only newly appeared issues.
func Replace(next []Issue) {
	storeMu.Lock()
	defer storeMu.Unlock()

	previous := make(map[string]struct{}, len(issues))
	for _, issue := range issues {
		previous[issueFingerprint(issue)] = struct{}{}
	}

	issues = make([]Issue, len(next))
	copy(issues, next)

	for _, issue := range issues {
		if _, exists := previous[issueFingerprint(issue)]; exists {
			continue
		}
		logNewIssue(issue)
	}
}

func issueFingerprint(issue Issue) string {
	return issue.Severity + "\x00" +
		issue.Code + "\x00" +
		issue.ActionID + "\x00" +
		issue.ActionTitle + "\x00" +
		issue.ArgumentName + "\x00" +
		issue.Source + "\x00" +
		issue.ConfigFile + "\x00" +
		issue.Message
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

// Report adds a live issue if it is not already present and logs it when new.
// Use this for runtime discoveries that happen outside Rebuild (for example
// filesystem watcher setup failures).
func Report(issue Issue) {
	storeMu.Lock()
	defer storeMu.Unlock()

	fp := issueFingerprint(issue)
	for _, existing := range issues {
		if issueFingerprint(existing) == fp {
			return
		}
	}

	issues = append(issues, issue)
	logNewIssue(issue)
}

// List returns a copy of the current issues.
func List() []Issue {
	storeMu.RLock()
	defer storeMu.RUnlock()
	out := make([]Issue, len(issues))
	copy(out, issues)
	return out
}

// Count returns the number of current issues.
func Count() int {
	storeMu.RLock()
	defer storeMu.RUnlock()
	return len(issues)
}
