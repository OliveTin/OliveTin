package configcheck_test

import (
	"testing"

	"github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/configcheck"
	"github.com/OliveTin/OliveTin/internal/configissues"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func unsignedArgIssues(webhook config.WebhookConfig, argNames ...string) []configissues.Issue {
	configissues.BeginConfigLoad()

	args := make([]config.ActionArgument, 0, len(argNames))
	for _, name := range argNames {
		args = append(args, config.ActionArgument{Name: name, Type: "ascii_identifier"})
	}

	cfg := config.DefaultConfig()
	cfg.Actions = []*config.Action{{
		ID:            "deploy",
		Title:         "Deploy",
		Exec:          []string{"true"},
		Arguments:     args,
		ExecOnWebhook: []config.WebhookConfig{webhook},
	}}

	configcheck.Rebuild(cfg)

	return issuesWithCode(configissues.List(), configissues.CodeWebhookUnsignedArg)
}

func hmacWebhook() config.WebhookConfig {
	return config.WebhookConfig{AuthType: "hmac-sha256", Secret: "s3cret"}
}

func TestRebuildWarnsForUnsignedArgumentsOnHmacWebhook(t *testing.T) {
	issues := unsignedArgIssues(hmacWebhook(), "webhook_header_x-target", "webhook_query", "webhook_path", "branch")

	require.Len(t, issues, 3)
	names := []string{issues[0].ArgumentName, issues[1].ArgumentName, issues[2].ArgumentName}
	assert.ElementsMatch(t, []string{"webhook_header_x-target", "webhook_query", "webhook_path"}, names)
	assert.Equal(t, configissues.SeverityWarning, issues[0].Severity)
	assert.Contains(t, issues[0].Message, "extract")
}

func TestRebuildWarnsForUnsignedArgumentsOnHmacSha1AndTemplateWebhooks(t *testing.T) {
	assert.Len(t, unsignedArgIssues(config.WebhookConfig{AuthType: "hmac-sha1", Secret: "s"}, "webhook_query"), 1)
	assert.Len(t, unsignedArgIssues(config.WebhookConfig{Template: "github-push", Secret: "s"}, "webhook_query"), 1)
}

func TestRebuildDoesNotWarnForArgumentsTakenFromTheBody(t *testing.T) {
	assert.Empty(t, unsignedArgIssues(hmacWebhook(), "branch", "webhook_method"))
}

func TestRebuildDoesNotWarnForWebhooksThatAuthenticateTheWholeRequest(t *testing.T) {
	for _, webhook := range []config.WebhookConfig{
		{AuthType: "bearer", Secret: "token"},
		{AuthType: "basic", Secret: "user:pass"},
		{AuthType: config.WebhookAuthTypeNone},
	} {
		assert.Empty(t, unsignedArgIssues(webhook, "webhook_query"), webhook.AuthType)
	}
}

func TestRebuildDoesNotTreatUnknownWebhookTemplateAsHmac(t *testing.T) {
	assert.Empty(t, unsignedArgIssues(config.WebhookConfig{Template: "github-pus", Secret: "s"}, "webhook_query"))
}
