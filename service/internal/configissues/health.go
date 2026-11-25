package configissues

import (
	"crypto/sha256"
	"encoding/hex"
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
// It is a hash so that it always fits armature-health's key limit and is stored
// unchanged; a longer key would be shortened, and stop matching.
func issueKey(issue Issue) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{
		issue.Code,
		issue.ActionID,
		issue.ActionTitle,
		issue.ArgumentName,
		issue.Source,
		issue.ConfigFile,
		issue.Message,
	}, "\x00")))
	return hex.EncodeToString(sum[:])
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
		// armature-health cuts titles at 512 characters, so the whole message is
		// also kept in the detail, which holds up to 16,000.
		Detail: issue.Message,
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

// toHealthIssues lists each problem once. When the same problem is collected
// as both a warning and an error, the error is kept, at the position the
// problem was first collected.
func toHealthIssues(issues []Issue) []health.Issue {
	positions := make(map[string]int, len(issues))
	out := make([]health.Issue, 0, len(issues))
	for _, issue := range issues {
		converted := toHealthIssue(issue)
		position, seen := positions[converted.Key]
		if !seen {
			positions[converted.Key] = len(out)
			out = append(out, converted)
			continue
		}
		if severityRank(string(converted.Severity)) > severityRank(string(out[position].Severity)) {
			out[position] = converted
		}
	}
	return out
}

func severityRank(severity string) int {
	if severity == string(health.SeverityError) {
		return 1
	}
	return 0
}

func fromHealthIssue(view health.IssueView) Issue {
	return Issue{
		Severity:     view.Severity,
		Code:         view.Attributes[attrCode],
		Message:      issueMessage(view),
		ActionID:     view.Attributes[attrActionID],
		ActionTitle:  view.Attributes[attrActionTitle],
		ArgumentName: view.Attributes[attrArgumentName],
		Source:       view.Attributes[attrSource],
		ConfigFile:   view.Attributes[attrConfigFile],
	}
}

func issueMessage(view health.IssueView) string {
	if view.Detail == "" {
		return view.Title
	}
	return view.Detail
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
	views := orderedIssueViews(stored)
	out := make([]Issue, 0, len(views))
	for _, view := range views {
		out = append(out, fromHealthIssue(view))
	}
	return out
}

func orderedIssueViews(stored []health.IssueView) []health.IssueView {
	views := uniqueIssueViews(stored)

	orderMu.Lock()
	defer orderMu.Unlock()

	sort.SliceStable(views, func(left, right int) bool {
		return issueSortsBefore(views[left], views[right])
	})
	return views
}

func uniqueIssueViews(stored []health.IssueView) []health.IssueView {
	views := slices.Clone(stored)
	sort.SliceStable(views, func(left, right int) bool {
		return sourceRank(views[left].CheckID) < sourceRank(views[right].CheckID)
	})

	// The same problem can be open in more than one source, with a different
	// severity in each. It is listed once, with the highest severity.
	positions := make(map[string]int, len(views))
	out := make([]health.IssueView, 0, len(views))
	for _, view := range views {
		position, seen := positions[view.IssueKey]
		if !seen {
			positions[view.IssueKey] = len(out)
			out = append(out, view)
			continue
		}
		if severityRank(view.Severity) > severityRank(out[position].Severity) {
			out[position] = view
		}
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
