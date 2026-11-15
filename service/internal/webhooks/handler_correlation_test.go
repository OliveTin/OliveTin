package webhooks

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/OliveTin/OliveTin/internal/config"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func postWebhook(t *testing.T, cfg *config.Config) (*httptest.ResponseRecorder, *test.Hook) {
	t.Helper()
	hook := test.NewGlobal()
	t.Cleanup(func() { log.StandardLogger().ReplaceHooks(make(log.LevelHooks)) })

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/webhooks?source=ci", strings.NewReader(`{"ref":"main"}`))
	rec := httptest.NewRecorder()
	NewWebhookHandler(cfg, nil).HandleWebhook(rec, req)

	return rec, hook
}

func requireCorrelationID(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	id := rec.Header().Get(CorrelationIDHeader)
	_, err := uuid.Parse(id)
	require.NoError(t, err, "header should be a bare UUID, got %q", id)
	return id
}

func findLogEntry(hook *test.Hook, message string) *log.Entry {
	for _, entry := range hook.AllEntries() {
		if entry.Message == message {
			return entry
		}
	}
	return nil
}

func TestWebhookNoMatchReturnsCorrelationIDInHeaderAndLog(t *testing.T) {
	rec, hook := postWebhook(t, config.DefaultConfig())

	assert.Equal(t, http.StatusOK, rec.Code)
	id := requireCorrelationID(t, rec)

	entry := findLogEntry(hook, "Webhook: no matching actions")
	require.NotNil(t, entry)
	assert.Equal(t, id, entry.Data["correlationId"])
}

func TestWebhookMatchUsesSameCorrelationIDForEveryLogLine(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Actions = []*config.Action{{
		Title: "Deploy",
		ExecOnWebhook: []config.WebhookConfig{{
			AuthType: "bearer",
			Secret:   "not-sent-by-this-request",
		}},
	}}

	rec, hook := postWebhook(t, cfg)

	id := requireCorrelationID(t, rec)

	matched := findLogEntry(hook, "Webhook: matched")
	require.NotNil(t, matched)
	assert.Equal(t, id, matched.Data["correlationId"])

	authFailed := findLogEntry(hook, "Webhook authentication failed")
	require.NotNil(t, authFailed)
	assert.Equal(t, id, authFailed.Data["correlationId"])
}

func TestWebhookCorrelationIDIsUniquePerRequest(t *testing.T) {
	first, _ := postWebhook(t, config.DefaultConfig())
	second, _ := postWebhook(t, config.DefaultConfig())

	assert.NotEqual(t, requireCorrelationID(t, first), requireCorrelationID(t, second))
}
