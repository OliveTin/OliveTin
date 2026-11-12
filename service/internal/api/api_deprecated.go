package api

import (
	"fmt"

	"github.com/OliveTin/OliveTin/internal/configissues"
)

// reportDeprecatedAPICall raises a warning on the Diagnostics page (once per
// method, until OliveTin restarts) when a deprecated API method is called, so
// admins can find and update the integrations still using it.
func reportDeprecatedAPICall(method, replacement string) {
	configissues.ReportRuntime(configissues.Issue{
		Severity: configissues.SeverityWarning,
		Code:     configissues.CodeDeprecatedAPICall,
		Message: fmt.Sprintf(
			"The deprecated API method %s was called. Use %s instead, as %s will be removed in a future version. See https://docs.olivetin.app/api/start_action.html#api-deprecated-byget",
			method, replacement, method),
		Source: method,
	})
}
