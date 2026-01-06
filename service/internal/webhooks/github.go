package webhooks

import (
	"github.com/OliveTin/OliveTin/internal/config"
)

// ApplyGitHubTemplate applies GitHub-specific template configurations
// This allows users to use simple template names instead of configuring everything manually
func ApplyGitHubTemplate(cfg *config.WebhookConfig, template string) {
	switch template {
	case "github-push":
		applyGitHubPushTemplate(cfg)
	case "github-pr", "github-pull-request":
		applyGitHubPRTemplate(cfg)
	case "github-release":
		applyGitHubReleaseTemplate(cfg)
	case "github-workflow":
		applyGitHubWorkflowTemplate(cfg)
	}
}

func applyGitHubPushTemplate(cfg *config.WebhookConfig) {
	if cfg.AuthHeader == "" {
		cfg.AuthHeader = "X-Hub-Signature-256"
	}
	if cfg.AuthType == "" {
		cfg.AuthType = "hmac-sha256"
	}
	if len(cfg.MatchHeaders) == 0 {
		cfg.MatchHeaders = make(map[string]string)
	}
	if _, exists := cfg.MatchHeaders["X-GitHub-Event"]; !exists {
		cfg.MatchHeaders["X-GitHub-Event"] = "push"
	}
	if len(cfg.Extract) == 0 {
		cfg.Extract = make(map[string]string)
	}
	if _, exists := cfg.Extract["git_repository"]; !exists {
		cfg.Extract["git_repository"] = "$.repository.full_name"
	}
	if _, exists := cfg.Extract["git_ref"]; !exists {
		cfg.Extract["git_ref"] = "$.ref"
	}
	if _, exists := cfg.Extract["git_commit"]; !exists {
		cfg.Extract["git_commit"] = "$.head_commit.id"
	}
	if _, exists := cfg.Extract["git_branch"]; !exists {
		cfg.Extract["git_branch"] = "$.ref"
	}
	if _, exists := cfg.Extract["git_message"]; !exists {
		cfg.Extract["git_message"] = "$.head_commit.message"
	}
	if _, exists := cfg.Extract["git_author"]; !exists {
		cfg.Extract["git_author"] = "$.head_commit.author.name"
	}
}

func applyGitHubPRTemplate(cfg *config.WebhookConfig) {
	if cfg.AuthHeader == "" {
		cfg.AuthHeader = "X-Hub-Signature-256"
	}
	if cfg.AuthType == "" {
		cfg.AuthType = "hmac-sha256"
	}
	if len(cfg.MatchHeaders) == 0 {
		cfg.MatchHeaders = make(map[string]string)
	}
	if _, exists := cfg.MatchHeaders["X-GitHub-Event"]; !exists {
		cfg.MatchHeaders["X-GitHub-Event"] = "pull_request"
	}
	if len(cfg.Extract) == 0 {
		cfg.Extract = make(map[string]string)
	}
	if _, exists := cfg.Extract["pr_number"]; !exists {
		cfg.Extract["pr_number"] = "$.number"
	}
	if _, exists := cfg.Extract["pr_title"]; !exists {
		cfg.Extract["pr_title"] = "$.pull_request.title"
	}
	if _, exists := cfg.Extract["pr_author"]; !exists {
		cfg.Extract["pr_author"] = "$.pull_request.user.login"
	}
	if _, exists := cfg.Extract["pr_action"]; !exists {
		cfg.Extract["pr_action"] = "$.action"
	}
	if _, exists := cfg.Extract["git_repository"]; !exists {
		cfg.Extract["git_repository"] = "$.repository.full_name"
	}
	if _, exists := cfg.Extract["pr_state"]; !exists {
		cfg.Extract["pr_state"] = "$.pull_request.state"
	}
	if _, exists := cfg.Extract["pr_head_sha"]; !exists {
		cfg.Extract["pr_head_sha"] = "$.pull_request.head.sha"
	}
}

func applyGitHubReleaseTemplate(cfg *config.WebhookConfig) {
	if cfg.AuthHeader == "" {
		cfg.AuthHeader = "X-Hub-Signature-256"
	}
	if cfg.AuthType == "" {
		cfg.AuthType = "hmac-sha256"
	}
	if len(cfg.MatchHeaders) == 0 {
		cfg.MatchHeaders = make(map[string]string)
	}
	if _, exists := cfg.MatchHeaders["X-GitHub-Event"]; !exists {
		cfg.MatchHeaders["X-GitHub-Event"] = "release"
	}
	if len(cfg.Extract) == 0 {
		cfg.Extract = make(map[string]string)
	}
	if _, exists := cfg.Extract["release_action"]; !exists {
		cfg.Extract["release_action"] = "$.action"
	}
	if _, exists := cfg.Extract["release_tag"]; !exists {
		cfg.Extract["release_tag"] = "$.release.tag_name"
	}
	if _, exists := cfg.Extract["release_name"]; !exists {
		cfg.Extract["release_name"] = "$.release.name"
	}
	if _, exists := cfg.Extract["git_repository"]; !exists {
		cfg.Extract["git_repository"] = "$.repository.full_name"
	}
	if _, exists := cfg.Extract["release_author"]; !exists {
		cfg.Extract["release_author"] = "$.release.author.login"
	}
}

func applyGitHubWorkflowTemplate(cfg *config.WebhookConfig) {
	if cfg.AuthHeader == "" {
		cfg.AuthHeader = "X-Hub-Signature-256"
	}
	if cfg.AuthType == "" {
		cfg.AuthType = "hmac-sha256"
	}
	if len(cfg.MatchHeaders) == 0 {
		cfg.MatchHeaders = make(map[string]string)
	}
	if _, exists := cfg.MatchHeaders["X-GitHub-Event"]; !exists {
		cfg.MatchHeaders["X-GitHub-Event"] = "workflow_run"
	}
	if len(cfg.Extract) == 0 {
		cfg.Extract = make(map[string]string)
	}
	if _, exists := cfg.Extract["workflow_name"]; !exists {
		cfg.Extract["workflow_name"] = "$.workflow_run.name"
	}
	if _, exists := cfg.Extract["workflow_status"]; !exists {
		cfg.Extract["workflow_status"] = "$.workflow_run.status"
	}
	if _, exists := cfg.Extract["workflow_conclusion"]; !exists {
		cfg.Extract["workflow_conclusion"] = "$.workflow_run.conclusion"
	}
	if _, exists := cfg.Extract["git_repository"]; !exists {
		cfg.Extract["git_repository"] = "$.repository.full_name"
	}
	if _, exists := cfg.Extract["git_commit"]; !exists {
		cfg.Extract["git_commit"] = "$.workflow_run.head_sha"
	}
	if _, exists := cfg.Extract["git_branch"]; !exists {
		cfg.Extract["git_branch"] = "$.workflow_run.head_branch"
	}
}
