package configissues

import (
	"context"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resetIssues(t *testing.T) {
	t.Helper()

	clear := func() {
		BeginConfigLoad()
		ReplaceLoadIssues(nil)
		Replace(nil)
		ClearRuntime()
	}

	clear()
	t.Cleanup(clear)
}

func captureLogs(t *testing.T) *logtest.Hook {
	t.Helper()

	hook := logtest.NewGlobal()
	t.Cleanup(func() {
		log.StandardLogger().ReplaceHooks(make(log.LevelHooks))
	})
	return hook
}

func issueLogEntries(hook *logtest.Hook) []*log.Entry {
	out := make([]*log.Entry, 0)
	for _, entry := range hook.AllEntries() {
		if _, ok := entry.Data["code"]; ok {
			out = append(out, entry)
		}
	}
	return out
}

func testIssue(code, message string) Issue {
	return Issue{Severity: SeverityWarning, Code: code, Message: message}
}

func messages(issues []Issue) []string {
	out := make([]string, 0, len(issues))
	for _, issue := range issues {
		out = append(out, issue.Message)
	}
	return out
}

func TestReplaceOnlyLogsNewIssues(t *testing.T) {
	resetIssues(t)
	hook := captureLogs(t)

	first := testIssue(CodeCronInvalid, "first")
	Replace([]Issue{first})
	Replace([]Issue{first})
	assert.Len(t, issueLogEntries(hook), 1, "an unchanged issue must not be logged again")

	Replace([]Issue{first, testIssue(CodeCronInvalid, "second")})
	entries := issueLogEntries(hook)
	require.Len(t, entries, 2)
	assert.Equal(t, "second", entries[1].Message)
	assert.Equal(t, log.WarnLevel, entries[1].Level)
	assert.Equal(t, CodeCronInvalid, entries[1].Data["code"])
}

func TestIssueIsLoggedAgainWhenItBecomesAnError(t *testing.T) {
	resetIssues(t)
	hook := captureLogs(t)

	issue := testIssue(CodeTemplateParse, "broken template")
	Replace([]Issue{issue})

	issue.Severity = SeverityError
	Replace([]Issue{issue})

	entries := issueLogEntries(hook)
	require.Len(t, entries, 2)
	assert.Equal(t, log.ErrorLevel, entries[1].Level)
	require.Len(t, List(), 1)
	assert.Equal(t, SeverityError, List()[0].Severity)
}

func TestIssueWithoutSeverityIsLoggedAsWarning(t *testing.T) {
	resetIssues(t)
	hook := captureLogs(t)

	Replace([]Issue{{Code: CodeCronInvalid, Message: "no severity"}})

	entries := issueLogEntries(hook)
	require.Len(t, entries, 1)
	assert.Equal(t, log.WarnLevel, entries[0].Level)
}

func TestRuntimeIssuesSurviveRebuildAndCountOccurrences(t *testing.T) {
	resetIssues(t)

	issue := Issue{Severity: SeverityWarning, Code: CodeDeprecatedAPICall, Message: "deprecated", Source: "StartActionByGet"}
	ReportRuntime(issue)
	ReportRuntime(issue)

	Replace(nil)
	ReplaceLoadIssues(nil)

	assert.Equal(t, []Issue{issue}, List())

	stored, err := service.ListIssues(context.Background(), false, true)
	require.NoError(t, err)
	require.Len(t, stored, 1)
	assert.Equal(t, 2, stored[0].Occurrences)
}

func TestReportedIssueIsResolvedByNextRebuild(t *testing.T) {
	resetIssues(t)

	Report(Issue{Severity: SeverityWarning, Code: CodeWatcherPath, Message: "cannot watch", ActionID: "a1"})
	assert.Equal(t, 1, Count())

	Replace(nil)
	assert.Equal(t, 0, Count())
}

func TestLoadIssuesAreClearedByTheNextLoad(t *testing.T) {
	resetIssues(t)

	BeginConfigLoad()
	ReportSticky(testIssue(CodeEnvUnset, "FOO is not set"))
	ReplaceLoadIssues(CopySticky())
	assert.Equal(t, 1, Count())

	BeginConfigLoad()
	ReplaceLoadIssues(CopySticky())
	assert.Equal(t, 0, Count())
}

func TestIssueInTwoSourcesIsListedOnce(t *testing.T) {
	resetIssues(t)

	issue := testIssue(CodeConfigDecode, "bad value")
	ReplaceLoadIssues([]Issue{issue})
	Report(issue)

	assert.Equal(t, []Issue{issue}, List())
}

func TestListKeepsCollectionOrder(t *testing.T) {
	resetIssues(t)

	ReplaceLoadIssues([]Issue{testIssue(CodeEnvUnset, "load")})
	ReportRuntime(testIssue(CodeDeprecatedAPICall, "runtime"))
	Replace([]Issue{testIssue(CodeCronInvalid, "config b"), testIssue(CodeCronInvalid, "config a")})
	Report(testIssue(CodeWatcherPath, "reported"))

	assert.Equal(t, []string{"load", "runtime", "config b", "config a", "reported"}, messages(List()))

	Replace([]Issue{testIssue(CodeCronInvalid, "config a"), testIssue(CodeCronInvalid, "config c"), testIssue(CodeCronInvalid, "config b")})

	assert.Equal(t, []string{"load", "runtime", "config a", "config c", "config b"}, messages(List()))
}

func TestLongIssuesKeepOrderAndAreNotLoggedAgain(t *testing.T) {
	resetIssues(t)
	hook := captureLogs(t)

	long := strings.Repeat("x", 600)
	issues := []Issue{
		testIssue(CodeConfigDecode, "b "+long),
		testIssue(CodeConfigDecode, "a "+long),
	}
	Replace(issues)
	Replace(issues)
	ReportRuntime(testIssue(CodeDeprecatedAPICall, long))
	ReportRuntime(testIssue(CodeDeprecatedAPICall, long))

	assert.Len(t, issueLogEntries(hook), 3, "an unchanged long issue must not be logged again")

	listed := List()
	require.Len(t, listed, 3)
	assert.True(t, strings.HasPrefix(listed[1].Message, "b "), "long issues must keep collection order")
	assert.True(t, strings.HasPrefix(listed[2].Message, "a "), "long issues must keep collection order")
}

func TestDuplicateIssueKeepsTheError(t *testing.T) {
	warning := testIssue(CodeConfigDecode, "bad value")
	failure := warning
	failure.Severity = SeverityError

	orders := map[string][]Issue{
		"warning first": {warning, failure},
		"error first":   {failure, warning},
	}
	for name, issues := range orders {
		t.Run(name, func(t *testing.T) {
			resetIssues(t)
			Replace(issues)
			assert.Equal(t, []Issue{failure}, List())
		})
	}
}

func TestIssueInTwoSourcesIsListedWithTheHighestSeverity(t *testing.T) {
	warning := testIssue(CodeConfigDecode, "bad value")
	failure := warning
	failure.Severity = SeverityError

	t.Run("error reported after a load warning", func(t *testing.T) {
		resetIssues(t)
		ReplaceLoadIssues([]Issue{warning})
		Report(failure)
		assert.Equal(t, []Issue{failure}, List())
	})

	t.Run("warning reported after a load error", func(t *testing.T) {
		resetIssues(t)
		ReplaceLoadIssues([]Issue{failure})
		Report(warning)
		assert.Equal(t, []Issue{failure}, List())
	})
}
