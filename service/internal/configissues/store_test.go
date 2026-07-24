package configissues_test

import (
	"testing"

	"github.com/OliveTin/OliveTin/internal/configissues"
	"github.com/stretchr/testify/assert"
)

func TestStoreReplaceListCount(t *testing.T) {
	configissues.BeginConfigLoad()
	configissues.Replace([]configissues.Issue{{
		Severity: configissues.SeverityError,
		Code:     configissues.CodeTemplateParse,
		Message:  "broken",
	}})

	assert.Equal(t, 1, configissues.Count())
	assert.Equal(t, configissues.CodeTemplateParse, configissues.List()[0].Code)

	configissues.Replace(nil)
	assert.Equal(t, 0, configissues.Count())
}

func TestStickyCopy(t *testing.T) {
	configissues.BeginConfigLoad()
	configissues.ReportSticky(configissues.Issue{
		Code:    configissues.CodeEnvUnset,
		Message: "missing",
		Source:  "FOO",
	})

	copied := configissues.CopySticky()
	assert.Len(t, copied, 1)
	assert.Equal(t, configissues.CodeEnvUnset, copied[0].Code)

	configissues.BeginConfigLoad()
	assert.Empty(t, configissues.CopySticky())
}
