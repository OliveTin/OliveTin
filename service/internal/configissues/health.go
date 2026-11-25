package configissues

import (
	"slices"
	"sort"
	"strings"
	"sync"

	health "github.com/jamesread/armature-health/v2"
)

const (
	attrCode         = "code"
	attrActionID     = "action_id"
	attrActionTitle  = "action_title"
	attrArgumentName = "argument_name"
	attrSource       = "source"
	attrConfigFile   = "config_file"
)

var (
	orderMu sync.Mutex
	// orderBySource keeps the position of each issue in the last list set for a
	// source, so List keeps the order the issues were collected in.
	orderBySource = map[string]map[string]int{}
)

// issueKey identifies the same problem across rebuilds. Severity is left out
// so that a change in severity updates the issue rather than raising a new one.
func issueKey(issue Issue) string {
	return strings.Join([]string{
		issue.Code,
		issue.ActionID,
		issue.ActionTitle,
		issue.ArgumentName,
		issue.Source,
		issue.ConfigFile,
		issue.Message,
	}, "\x00")
}

// armature-health treats anything that is not a warning or an error as info.
// OliveTin has always treated such issues as warnings.
func healthSeverity(severity string) health.Severity {
	if severity == SeverityError {
		return health.SeverityError
	}
	return health.SeverityWarning
}

func toHealthIssue(issue Issue) health.Issue {
	return health.Issue{
		Key:      issueKey(issue),
		Severity: healthSeverity(issue.Severity),
		Title:    issue.Message,
		Attributes: map[string]string{
			attrCode:         issue.Code,
			attrActionID:     issue.ActionID,
			attrActionTitle:  issue.ActionTitle,
			attrArgumentName: issue.ArgumentName,
			attrSource:       issue.Source,
			attrConfigFile:   issue.ConfigFile,
		},
	}
}

func toHealthIssues(issues []Issue) []health.Issue {
	seen := make(map[string]bool, len(issues))
	out := make([]health.Issue, 0, len(issues))
	for _, issue := range issues {
		converted := toHealthIssue(issue)
		if seen[converted.Key] {
			continue
		}
		seen[converted.Key] = true
		out = append(out, converted)
	}
	return out
}

func fromHealthIssue(view health.IssueView) Issue {
	return Issue{
		Severity:     view.Severity,
		Code:         view.Attributes[attrCode],
		Message:      view.Title,
		ActionID:     view.Attributes[attrActionID],
		ActionTitle:  view.Attributes[attrActionTitle],
		ArgumentName: view.Attributes[attrArgumentName],
		Source:       view.Attributes[attrSource],
		ConfigFile:   view.Attributes[attrConfigFile],
	}
}

func rememberOrder(sourceID string, issues []Issue) {
	positions := make(map[string]int, len(issues))
	for index, issue := range issues {
		key := issueKey(issue)
		if _, exists := positions[key]; !exists {
			positions[key] = index
		}
	}

	orderMu.Lock()
	defer orderMu.Unlock()
	orderBySource[sourceID] = positions
}

// orderedIssues lists each problem once, in the order OliveTin has always used:
// load-time issues, then runtime issues, then issues derived from the config.
// Within a source, issues keep the order they were collected in, and issues
// reported between rebuilds follow in the order they were reported.
func orderedIssues(stored []health.IssueView) []Issue {
	views := uniqueIssueViews(stored)

	orderMu.Lock()
	defer orderMu.Unlock()

	sort.SliceStable(views, func(left, right int) bool {
		return issueSortsBefore(views[left], views[right])
	})

	out := make([]Issue, 0, len(views))
	for _, view := range views {
		out = append(out, fromHealthIssue(view))
	}
	return out
}

func uniqueIssueViews(stored []health.IssueView) []health.IssueView {
	views := slices.Clone(stored)
	sort.SliceStable(views, func(left, right int) bool {
		return sourceRank(views[left].CheckID) < sourceRank(views[right].CheckID)
	})

	seen := make(map[string]bool, len(views))
	out := make([]health.IssueView, 0, len(views))
	for _, view := range views {
		if seen[view.IssueKey] {
			continue
		}
		seen[view.IssueKey] = true
		out = append(out, view)
	}
	return out
}

func issueSortsBefore(left, right health.IssueView) bool {
	if sourceRank(left.CheckID) != sourceRank(right.CheckID) {
		return sourceRank(left.CheckID) < sourceRank(right.CheckID)
	}

	leftPosition, leftPushed := orderBySource[left.CheckID][left.IssueKey]
	rightPosition, rightPushed := orderBySource[right.CheckID][right.IssueKey]
	if leftPushed && rightPushed {
		return leftPosition < rightPosition
	}
	if leftPushed != rightPushed {
		return leftPushed
	}
	return left.ID < right.ID
}

func sourceRank(sourceID string) int {
	return slices.Index(sourceOrder, sourceID)
}
