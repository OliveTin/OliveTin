package executor

import (
	"fmt"
	"strings"

	authpublic "github.com/OliveTin/OliveTin/internal/auth/authpublic"
)

const (
	justificationCron       = "Triggered by cron"
	justificationStartup    = "Triggered by startup"
	justificationFileChange = "Triggered by file change"
	justificationCalendar   = "Triggered by calendar"
	justificationWebhook    = "Triggered by webhook"
)

var systemJustificationDefaults = map[string]string{
	"cron":      justificationCron,
	"startup":   justificationStartup,
	"fileindir": justificationFileChange,
	"calendar":  justificationCalendar,
	"webhook":   justificationWebhook,
}

func IsSystemExecution(user *authpublic.AuthenticatedUser) bool {
	if user == nil || user.Provider != "system" {
		return false
	}

	return user.Username != "guest"
}

func ResolveJustification(req *ExecutionRequest) string {
	provided := strings.TrimSpace(reqJustification(req))
	if provided != "" {
		return provided
	}

	if !actionRequiresJustification(req) {
		return ""
	}

	return defaultJustificationForRequest(req)
}

func actionRequiresJustification(req *ExecutionRequest) bool {
	return req != nil && req.Binding != nil && req.Binding.Action != nil && req.Binding.Action.RequiresJustification()
}

func defaultJustificationForRequest(req *ExecutionRequest) string {
	if req.TriggerDepth > 0 && req.logEntry != nil {
		return fmt.Sprintf("Triggered by action: %s", req.logEntry.ActionTitle)
	}

	if req.AuthenticatedUser == nil {
		return ""
	}

	return systemJustificationDefaults[req.AuthenticatedUser.Username]
}

func reqJustification(req *ExecutionRequest) string {
	if req == nil {
		return ""
	}
	return req.Justification
}
