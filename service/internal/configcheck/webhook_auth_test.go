package configcheck_test

import (
	"testing"

	"github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/configcheck"
	"github.com/OliveTin/OliveTin/internal/configissues"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func webhookAuthIssues(webhooks ...config.WebhookConfig) []configissues.Issue {
	configissues.BeginConfigLoad()
	cfg := config.DefaultConfig()
	cfg.Actions = []*config.Action{{ID: "deploy", Title: "Deploy", Exec: []string{"true"}, ExecOnWebhook: webhooks}}

	configcheck.Rebuild(cfg)

	return issuesWithCode(configissues.List(), configissues.CodeWebhookAuthTypeUnset)
}

func TestRebuildReportsWebhookWithoutAuthType(t *testing.T) {
	issues := webhookAuthIssues(config.WebhookConfig{MatchQuery: map[string]string{"demo": "x"}})

	require.Len(t, issues, 1)
	assert.Equal(t, configissues.SeverityError, issues[0].Severity)
	assert.Equal(t, "Deploy", issues[0].ActionTitle)
	assert.Contains(t, issues[0].Message, "authType: none")
}

func TestRebuildAcceptsWebhookWithExplicitAuthType(t *testing.T) {
	assert.Empty(t, webhookAuthIssues(
		config.WebhookConfig{AuthType: config.WebhookAuthTypeNone},
		config.WebhookConfig{AuthType: "bearer", Secret: "token"},
	))
}

func TestRebuildAcceptsWebhookTemplateWithoutAuthType(t *testing.T) {
	assert.Empty(t, webhookAuthIssues(config.WebhookConfig{Template: "github-push", Secret: "s3cret"}))
}

func TestRebuildReportsEachWebhookWithoutAuthType(t *testing.T) {
	issues := webhookAuthIssues(config.WebhookConfig{}, config.WebhookConfig{AuthType: "bearer"}, config.WebhookConfig{})

	require.Len(t, issues, 2)
	assert.Equal(t, "execOnWebhook[0]", issues[0].Source)
	assert.Equal(t, "execOnWebhook[2]", issues[1].Source)
}

func TestRebuildReportsUnknownWebhookTemplateWithoutAuthType(t *testing.T) {
	webhook := config.WebhookConfig{Template: "github-pus", Secret: "s3cret"}

	assert.Len(t, webhookAuthIssues(webhook), 1, "a misspelt template must not excuse a missing authType")

	templateIssues := issuesWithCode(configissues.List(), configissues.CodeWebhookTemplateUnknown)
	require.Len(t, templateIssues, 1)
	assert.Contains(t, templateIssues[0].Message, "github-pus")
	assert.Contains(t, templateIssues[0].Message, "github-push")
	assert.Equal(t, "execOnWebhook[0]", templateIssues[0].Source)
}

func TestRebuildReportsUnknownWebhookTemplateEvenWithAuthType(t *testing.T) {
	webhook := config.WebhookConfig{Template: "gitlab-push", AuthType: "bearer", Secret: "token"}

	assert.Empty(t, webhookAuthIssues(webhook))
	assert.Len(t, issuesWithCode(configissues.List(), configissues.CodeWebhookTemplateUnknown), 1)
}

func TestRebuildAcceptsEveryKnownWebhookTemplate(t *testing.T) {
	for _, name := range config.WebhookTemplateNames {
		assert.Empty(t, webhookAuthIssues(config.WebhookConfig{Template: name, Secret: "s3cret"}), name)
		assert.Empty(t, issuesWithCode(configissues.List(), configissues.CodeWebhookTemplateUnknown), name)
	}
}
