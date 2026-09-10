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

func TestGetGroupFieldString(t *testing.T) {
	data := map[string]any{"olivetin_group": "admins"}

	assert.Equal(t, "admins", getGroupField(data, "olivetin_group", ""))
}

func TestGetGroupFieldMissing(t *testing.T) {
	data := map[string]any{}

	assert.Equal(t, "", getGroupField(data, "olivetin_group", ""))
}

func TestGetGroupFieldEmptyFieldName(t *testing.T) {
	data := map[string]any{"olivetin_group": "admins"}

	assert.Equal(t, "", getGroupField(data, "", ""))
}

func TestGetGroupFieldArrayDefaultSeparator(t *testing.T) {
	data := map[string]any{"groups": []any{"admins", "ops"}}

	assert.Equal(t, "admins ops", getGroupField(data, "groups", ""))
}

func TestGetGroupFieldArrayCustomSeparator(t *testing.T) {
	data := map[string]any{"groups": []any{"admins", "ops"}}

	assert.Equal(t, "admins,ops", getGroupField(data, "groups", ","))
}

func TestGetGroupFieldArraySkipsNonStringElements(t *testing.T) {
	data := map[string]any{"groups": []any{"admins", float64(5), "ops"}}

	assert.Equal(t, "admins ops", getGroupField(data, "groups", ""))
}

func TestGetGroupFieldArrayAllNonStringElements(t *testing.T) {
	data := map[string]any{"groups": []any{float64(1), true}}

	assert.Equal(t, "", getGroupField(data, "groups", ""))
}

func TestGetGroupFieldNotStringOrArray(t *testing.T) {
	data := map[string]any{"groups": map[string]any{"nested": "value"}}

	assert.Equal(t, "", getGroupField(data, "groups", ""))
}

func TestGetUserInfoWithArrayGroupsClaim(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"preferred_username":"john","groups":["admins","ops"]}`))
	}))
	defer srv.Close()

	cfg := config.DefaultConfig()
	cfg.AuthHttpHeaderUserGroupSep = ","

	provider := &config.OAuth2Provider{
		WhoamiUrl:      srv.URL,
		UsernameField:  "preferred_username",
		UserGroupField: "groups",
	}

	userinfo := getUserInfo(cfg, srv.Client(), provider)

	assert.Equal(t, "john", userinfo.Username)
	assert.Equal(t, "admins,ops", userinfo.Usergroup)
}

func TestComputeUsergroupUsesConfiguredSeparatorWithAddToUsergroup(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.AuthHttpHeaderUserGroupSep = ","

	h := &OAuth2Handler{cfg: cfg}

	userinfo := &UserInfo{Usergroup: "admins,ops"}
	providerConfig := &config.OAuth2Provider{AddToUsergroup: "github"}

	assert.Equal(t, "admins,ops,github", h.computeUsergroup(userinfo, providerConfig))
}

func TestComputeUsergroupDefaultSeparatorWithAddToUsergroup(t *testing.T) {
	cfg := config.DefaultConfig()

	h := &OAuth2Handler{cfg: cfg}

	userinfo := &UserInfo{Usergroup: "admins"}
	providerConfig := &config.OAuth2Provider{AddToUsergroup: "github"}

	assert.Equal(t, "admins github", h.computeUsergroup(userinfo, providerConfig))
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

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/oauth/login?provider=test", nil)
	rec := httptest.NewRecorder()

	h.HandleOAuthLogin(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.Equal(t, oauthStateMaxEntries, len(h.registeredStates))
}
