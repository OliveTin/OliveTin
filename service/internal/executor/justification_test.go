package executor

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/OliveTin/OliveTin/internal/auth"
	config "github.com/OliveTin/OliveTin/internal/config"
)

func TestResolveJustificationUsesProvidedValue(t *testing.T) {
	cfg := config.DefaultConfig()
	action := &config.Action{Title: "Send email", Justification: config.JustificationRequiredNoTemplate, Shell: "echo hi"}
	cfg.Actions = append(cfg.Actions, action)
	ex := DefaultExecutor(cfg)
	ex.RebuildActionMap()

	req := &ExecutionRequest{
		Binding:           ex.FindBindingWithNoEntity(action),
		Justification:     "New user registration foo@example.com",
		AuthenticatedUser: auth.UserGuest(cfg),
		Cfg:               cfg,
	}
	req.logEntry = &InternalLogEntry{}

	assert.Equal(t, "New user registration foo@example.com", ResolveJustification(req))
}

func TestResolveJustificationCronDefault(t *testing.T) {
	cfg := config.DefaultConfig()
	action := &config.Action{Title: "Nightly backup", Justification: config.JustificationRequiredNoTemplate, Shell: "echo hi"}
	cfg.Actions = append(cfg.Actions, action)
	ex := DefaultExecutor(cfg)
	ex.RebuildActionMap()

	req := &ExecutionRequest{
		Binding:           ex.FindBindingWithNoEntity(action),
		AuthenticatedUser: auth.UserFromSystem(cfg, "cron"),
		Cfg:               cfg,
	}

	assert.Equal(t, justificationCron, ResolveJustification(req))
}

func TestResolveJustificationStartupDefault(t *testing.T) {
	cfg := config.DefaultConfig()
	action := &config.Action{Title: "Init", Justification: config.JustificationRequiredNoTemplate, Shell: "echo hi"}
	cfg.Actions = append(cfg.Actions, action)
	ex := DefaultExecutor(cfg)
	ex.RebuildActionMap()

	req := &ExecutionRequest{
		Binding:           ex.FindBindingWithNoEntity(action),
		AuthenticatedUser: auth.UserFromSystem(cfg, "startup"),
		Cfg:               cfg,
	}

	assert.Equal(t, justificationStartup, ResolveJustification(req))
}

func TestResolveJustificationWebhookDefault(t *testing.T) {
	cfg := config.DefaultConfig()
	action := &config.Action{Title: "Deploy", Justification: config.JustificationRequiredNoTemplate, Exec: []string{"echo", "deploy"}}
	cfg.Actions = append(cfg.Actions, action)
	ex := DefaultExecutor(cfg)
	ex.RebuildActionMap()

	req := &ExecutionRequest{
		Binding:           ex.FindBindingWithNoEntity(action),
		AuthenticatedUser: auth.UserFromSystem(cfg, "webhook"),
		Cfg:               cfg,
	}

	assert.Equal(t, justificationWebhook, ResolveJustification(req))
}

func TestResolveJustificationEmptyWhenNotRequired(t *testing.T) {
	cfg := config.DefaultConfig()
	action := &config.Action{Title: "Ping", Shell: "echo hi"}
	cfg.Actions = append(cfg.Actions, action)
	ex := DefaultExecutor(cfg)
	ex.RebuildActionMap()

	req := &ExecutionRequest{
		Binding:           ex.FindBindingWithNoEntity(action),
		AuthenticatedUser: auth.UserGuest(cfg),
		Cfg:               cfg,
	}

	assert.Empty(t, ResolveJustification(req))
}

func TestJustificationNotPassedToShellArgs(t *testing.T) {
	cfg := config.DefaultConfig()
	action := &config.Action{
		Title:         "Echo",
		Justification: config.JustificationRequiredNoTemplate,
		Shell:         "echo {{ message }}",
		Arguments: []config.ActionArgument{
			{Name: "message", Type: "ascii_sentence"},
		},
	}
	cfg.Actions = append(cfg.Actions, action)
	ex := DefaultExecutor(cfg)
	ex.RebuildActionMap()

	req := &ExecutionRequest{
		Binding: ex.FindBindingWithNoEntity(action),
		Arguments: map[string]string{
			"message":       "hello",
			"justification": "should be stripped",
		},
		Justification:     "audit reason",
		AuthenticatedUser: auth.UserGuest(cfg),
		Cfg:               cfg,
	}
	req.logEntry = &InternalLogEntry{}

	filterToDefinedArgumentsOnly(req)

	assert.Equal(t, "hello", req.Arguments["message"])
	assert.Empty(t, req.Arguments["justification"])
}

func TestIsSystemExecution(t *testing.T) {
	cfg := config.DefaultConfig()
	assert.True(t, IsSystemExecution(auth.UserFromSystem(cfg, "cron")))
	assert.False(t, IsSystemExecution(auth.UserGuest(cfg)))
}
