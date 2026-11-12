package configissues_test

import (
	"testing"

	"github.com/OliveTin/OliveTin/internal/configissues"
	"github.com/stretchr/testify/assert"
)

func TestReportRuntimeDeduplicatesAndLists(t *testing.T) {
	configissues.ClearRuntime()
	configissues.Replace(nil)
	t.Cleanup(func() {
		configissues.ClearRuntime()
		configissues.Replace(nil)
	})

	issue := configissues.Issue{
		Severity: configissues.SeverityWarning,
		Code:     configissues.CodeDeprecatedAPICall,
		Message:  "deprecated",
		Source:   "StartActionByGet",
	}

	configissues.ReportRuntime(issue)
	configissues.ReportRuntime(issue)

	assert.Len(t, configissues.CopyRuntime(), 1)
	assert.Equal(t, 1, configissues.Count())
}
