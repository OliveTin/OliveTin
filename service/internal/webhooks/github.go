package webhooks

import (
	"github.com/OliveTin/OliveTin/internal/config"
)

// ApplyGitHubTemplate applies GitHub-specific template configurations
// This allows users to use simple template names instead of configuring everything manually
func ApplyGitHubTemplate(cfg *config.WebhookConfig, template string) {
	applier := getTemplateApplier(template)
	if applier != nil {
		applier(cfg)
	}
}

type templateApplier func(*config.WebhookConfig)

func getTemplateApplier(template string) templateApplier {
	templateMap := map[string]templateApplier{
		"github-push":         applyGitHubPushTemplate,
		"github-pr":           applyGitHubPRTemplate,
		"github-pull-request": applyGitHubPRTemplate,
		"github-release":      applyGitHubReleaseTemplate,
		"github-workflow":     applyGitHubWorkflowTemplate,
	}
	return templateMap[template]
}

func setDefaultAuth(cfg *config.WebhookConfig) {
	if cfg.AuthHeader == "" {
		cfg.AuthHeader = "X-Hub-Signature-256"
	}
	if cfg.AuthType == "" {
		cfg.AuthType = "hmac-sha256"
	}
}

func ensureMatchHeaders(cfg *config.WebhookConfig) {
	if len(cfg.MatchHeaders) == 0 {
		cfg.MatchHeaders = make(map[string]string)
	}
}

func ensureExtract(cfg *config.WebhookConfig) {
	if len(cfg.Extract) == 0 {
		cfg.Extract = make(map[string]string)
	}
}

func setExtractIfMissing(cfg *config.WebhookConfig, key, value string) {
	if _, exists := cfg.Extract[key]; !exists {
		cfg.Extract[key] = value
	}
}

func setMatchHeaderIfMissing(cfg *config.WebhookConfig, key, value string) {
	if _, exists := cfg.MatchHeaders[key]; !exists {
		cfg.MatchHeaders[key] = value
	}
}

func applyGitHubPushTemplate(cfg *config.WebhookConfig) {
	setDefaultAuth(cfg)
	ensureMatchHeaders(cfg)
	setMatchHeaderIfMissing(cfg, "X-GitHub-Event", "push")
	ensureExtract(cfg)
	setExtractIfMissing(cfg, "git_repository", "$.repository.full_name")
	setExtractIfMissing(cfg, "git_ref", "$.ref")
	setExtractIfMissing(cfg, "git_commit", "$.head_commit.id")
	setExtractIfMissing(cfg, "git_branch", "$.ref")
	setExtractIfMissing(cfg, "git_message", "$.head_commit.message")
	setExtractIfMissing(cfg, "git_author", "$.head_commit.author.name")
}

func applyGitHubPRTemplate(cfg *config.WebhookConfig) {
	setDefaultAuth(cfg)
	ensureMatchHeaders(cfg)
	setMatchHeaderIfMissing(cfg, "X-GitHub-Event", "pull_request")
	ensureExtract(cfg)
	setExtractIfMissing(cfg, "pr_number", "$.number")
	setExtractIfMissing(cfg, "pr_title", "$.pull_request.title")
	setExtractIfMissing(cfg, "pr_author", "$.pull_request.user.login")
	setExtractIfMissing(cfg, "pr_action", "$.action")
	setExtractIfMissing(cfg, "git_repository", "$.repository.full_name")
	setExtractIfMissing(cfg, "pr_state", "$.pull_request.state")
	setExtractIfMissing(cfg, "pr_head_sha", "$.pull_request.head.sha")
}

func applyGitHubReleaseTemplate(cfg *config.WebhookConfig) {
	setDefaultAuth(cfg)
	ensureMatchHeaders(cfg)
	setMatchHeaderIfMissing(cfg, "X-GitHub-Event", "release")
	ensureExtract(cfg)
	setExtractIfMissing(cfg, "release_action", "$.action")
	setExtractIfMissing(cfg, "release_tag", "$.release.tag_name")
	setExtractIfMissing(cfg, "release_name", "$.release.name")
	setExtractIfMissing(cfg, "git_repository", "$.repository.full_name")
	setExtractIfMissing(cfg, "release_author", "$.release.author.login")
}

func applyGitHubWorkflowTemplate(cfg *config.WebhookConfig) {
	setDefaultAuth(cfg)
	ensureMatchHeaders(cfg)
	setMatchHeaderIfMissing(cfg, "X-GitHub-Event", "workflow_run")
	ensureExtract(cfg)
	setExtractIfMissing(cfg, "workflow_name", "$.workflow_run.name")
	setExtractIfMissing(cfg, "workflow_status", "$.workflow_run.status")
	setExtractIfMissing(cfg, "workflow_conclusion", "$.workflow_run.conclusion")
	setExtractIfMissing(cfg, "git_repository", "$.repository.full_name")
	setExtractIfMissing(cfg, "git_commit", "$.workflow_run.head_sha")
	setExtractIfMissing(cfg, "git_branch", "$.workflow_run.head_branch")
}
