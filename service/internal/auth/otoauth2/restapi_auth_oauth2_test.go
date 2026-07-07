package otoauth2

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

func TestSweepExpiredOAuthStatesLocked(t *testing.T) {
	h := &OAuth2Handler{
		registeredStates: make(map[string]*oauth2State),
	}

	h.registeredStates["fresh"] = &oauth2State{
		providerName: "test",
		createdAt:    time.Now(),
	}
	h.registeredStates["stale"] = &oauth2State{
		providerName: "test",
		createdAt:    time.Now().Add(-2 * oauthStateMaxAge * time.Second),
	}

	h.sweepExpiredOAuthStatesLocked(time.Now())

	_, freshFound := h.registeredStates["fresh"]
	_, staleFound := h.registeredStates["stale"]
	assert.True(t, freshFound)
	assert.False(t, staleFound)
}

func TestHandleOAuthLoginRejectsWhenStateMapFull(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.AuthOAuth2Providers = map[string]*config.OAuth2Provider{
		"test": {
			Name:         "test",
			ClientID:     "id",
			ClientSecret: "secret",
			AuthUrl:      "https://example.com/auth",
			TokenUrl:     "https://example.com/token",
		},
	}

	h := NewOAuth2Handler(cfg)
	h.registeredStates = make(map[string]*oauth2State, oauthStateMaxEntries)
	for i := 0; i < oauthStateMaxEntries; i++ {
		h.registeredStates[strconv.Itoa(i)] = &oauth2State{
			providerConfig: &oauth2.Config{},
			providerName:   "test",
			createdAt:      time.Now(),
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/oauth/login?provider=test", nil)
	rec := httptest.NewRecorder()

	h.HandleOAuthLogin(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.Equal(t, oauthStateMaxEntries, len(h.registeredStates))
}
