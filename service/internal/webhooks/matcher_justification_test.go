package webhooks

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	config "github.com/OliveTin/OliveTin/internal/config"
)

func TestExtractJustificationFromWebhookBody(t *testing.T) {
	body := []byte(`{"message":"deploy production","repo":"my-app"}`)
	req, err := http.NewRequest(http.MethodPost, "/webhooks/deploy", nil)
	require.NoError(t, err)

	matcher := NewWebhookMatcher(config.WebhookConfig{
		Justification: "$.message",
	}, req, body)

	value, err := matcher.ExtractJustification()
	require.NoError(t, err)
	assert.Equal(t, "deploy production", value)
}

func TestExtractJustificationEmptyWhenNotConfigured(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, "/webhooks/deploy", nil)
	require.NoError(t, err)

	matcher := NewWebhookMatcher(config.WebhookConfig{}, req, []byte(`{}`))

	value, err := matcher.ExtractJustification()
	require.NoError(t, err)
	assert.Empty(t, value)
}
