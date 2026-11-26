package webhooks

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/executor"
)

func unauthenticatedWebhookRequest(t *testing.T) *http.Request {
	t.Helper()

	return httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/webhooks?demo=open", strings.NewReader(`{}`))
}

func TestWebhookWithoutAuthTypeIsRejected(t *testing.T) {
	verifier := NewAuthVerifier(config.WebhookConfig{})

	assert.False(t, verifier.Verify(unauthenticatedWebhookRequest(t), []byte(`{}`)))
}

func TestWebhookWithAuthTypeNoneAcceptsUnauthenticatedRequests(t *testing.T) {
	verifier := NewAuthVerifier(config.WebhookConfig{AuthType: config.WebhookAuthTypeNone})

	assert.True(t, verifier.Verify(unauthenticatedWebhookRequest(t), []byte(`{}`)))
}

func TestGitHubTemplateWithoutAuthTypeStillRequiresSignature(t *testing.T) {
	webhook := config.WebhookConfig{Template: "github-push", Secret: "s3cret"}
	ApplyGitHubTemplate(&webhook, webhook.Template)

	assert.Equal(t, "hmac-sha256", webhook.AuthType)
	assert.False(t, NewAuthVerifier(webhook).Verify(unauthenticatedWebhookRequest(t), []byte(`{}`)))
}

func webhookExecutionCount(t *testing.T, authType string) int {
	t.Helper()

	cfg := config.DefaultConfig()
	cfg.Actions = []*config.Action{{
		ID:    "open-webhook",
		Title: "Open webhook",
		Exec:  []string{"echo", "ran"},
		ExecOnWebhook: []config.WebhookConfig{{
			AuthType:   authType,
			MatchQuery: map[string]string{"demo": "open"},
		}},
	}}

	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()

	rec := httptest.NewRecorder()
	NewWebhookHandler(cfg, ex).HandleWebhook(rec, unauthenticatedWebhookRequest(t))
	require.Equal(t, http.StatusOK, rec.Code)

	time.Sleep(100 * time.Millisecond)
	logs, _ := ex.GetLogTrackingIds(0, 100)

	return len(logs)
}

func TestWebhookHandlerDoesNotRunActionWithoutAuthType(t *testing.T) {
	assert.Equal(t, 0, webhookExecutionCount(t, ""), "a webhook with no authType must not start its action")
}

func TestWebhookHandlerRunsActionWithAuthTypeNone(t *testing.T) {
	assert.Equal(t, 1, webhookExecutionCount(t, config.WebhookAuthTypeNone))
}
