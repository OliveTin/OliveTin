package otoauth2

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	authTypes "github.com/OliveTin/OliveTin/internal/auth/authpublic"
	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestSweepExpiredOAuthEntriesLocked(t *testing.T) {
	h := &OAuth2Handler{
		pendingStates: make(map[string]*oauth2PendingState),
		sessions:      make(map[string]*oauth2Session),
	}

	h.pendingStates["fresh"] = &oauth2PendingState{
		providerName: "test",
		createdAt:    time.Now(),
	}
	h.pendingStates["stale"] = &oauth2PendingState{
		providerName: "test",
		createdAt:    time.Now().Add(-2 * oauthStateMaxAge * time.Second),
	}
	h.sessions["fresh-session"] = &oauth2Session{
		Username:  "alice",
		createdAt: time.Now(),
	}
	h.sessions["stale-session"] = &oauth2Session{
		Username:  "bob",
		createdAt: time.Now().Add(-2 * oauthSessionMaxAge * time.Second),
	}

	h.sweepExpiredOAuthEntriesLocked(time.Now())

	_, freshPending := h.pendingStates["fresh"]
	_, stalePending := h.pendingStates["stale"]
	_, freshSession := h.sessions["fresh-session"]
	_, staleSession := h.sessions["stale-session"]
	assert.True(t, freshPending)
	assert.False(t, stalePending)
	assert.True(t, freshSession)
	assert.False(t, staleSession)
}

func TestGetGroupFieldString(t *testing.T) {
	data := map[string]any{"olivetin_group": "admins"}

	assert.Equal(t, "admins", getGroupField(data, "olivetin_group", ""))
}

func TestGetGroupFieldMissing(t *testing.T) {
	data := map[string]any{}

	assert.Empty(t, getGroupField(data, "olivetin_group", ""))
}

func TestGetGroupFieldEmptyFieldName(t *testing.T) {
	data := map[string]any{"olivetin_group": "admins"}

	assert.Empty(t, getGroupField(data, "", ""))
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

	assert.Empty(t, getGroupField(data, "groups", ""))
}

func TestGetGroupFieldNotStringOrArray(t *testing.T) {
	data := map[string]any{"groups": map[string]any{"nested": "value"}}

	assert.Empty(t, getGroupField(data, "groups", ""))
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
	h.pendingStates = make(map[string]*oauth2PendingState, oauthStateMaxEntries)
	for i := range oauthStateMaxEntries {
		h.pendingStates[strconv.Itoa(i)] = &oauth2PendingState{
			providerConfig: &oauth2.Config{},
			providerName:   "test",
			createdAt:      time.Now(),
		}
	}

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/oauth/login?provider=test", nil)
	rec := httptest.NewRecorder()

	h.HandleOAuthLogin(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.Len(t, h.pendingStates, oauthStateMaxEntries)
}

func TestHandleOAuthLoginUsesSeparateStateCookie(t *testing.T) {
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
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/oauth/login?provider=test", nil)
	rec := httptest.NewRecorder()

	h.HandleOAuthLogin(rec, req)

	assert.Equal(t, http.StatusFound, rec.Code)

	cookies := rec.Result().Cookies()
	var stateCookie, sessionCookie *http.Cookie
	for _, cookie := range cookies {
		switch cookie.Name {
		case oauthStateCookieName:
			stateCookie = cookie
		case oauthSessionCookieName:
			sessionCookie = cookie
		}
	}

	requireCookie := func(cookie *http.Cookie, name string) string {
		t.Helper()
		assert.NotNil(t, cookie, "expected %s cookie", name)
		assert.NotEmpty(t, cookie.Value)
		return cookie.Value
	}

	stateValue := requireCookie(stateCookie, oauthStateCookieName)
	assert.Nil(t, sessionCookie, "login must not set the session cookie yet")

	location, err := url.Parse(rec.Header().Get("Location"))
	require.NoError(t, err)
	assert.Equal(t, stateValue, location.Query().Get("state"))
	assert.Contains(t, h.pendingStates, stateValue)
	assert.Empty(t, h.sessions)
}

func TestIssueOAuthSessionRotatesAwayFromCSRFState(t *testing.T) {
	h := &OAuth2Handler{
		pendingStates: make(map[string]*oauth2PendingState),
		sessions:      make(map[string]*oauth2Session),
	}

	state := "csrf-state-value"
	h.pendingStates[state] = &oauth2PendingState{
		providerName: "test",
		createdAt:    time.Now(),
	}

	sessionID, err := h.issueOAuthSession(h.pendingStates[state], state, "alice", "admins")
	require.NoError(t, err)
	assert.NotEmpty(t, sessionID)
	assert.NotEqual(t, state, sessionID)
	assert.NotContains(t, h.pendingStates, state)
	assert.Contains(t, h.sessions, sessionID)

	user, found := h.lookupOAuth2UserBySession(sessionID)
	assert.True(t, found)
	assert.Equal(t, "alice", user.Username)
	assert.Equal(t, sessionID, user.SID)

	_, foundAsState := h.lookupOAuth2UserBySession(state)
	assert.False(t, foundAsState, "CSRF state must not authenticate as a session")
}

func TestCheckUserFromOAuth2CookieUsesSessionCookieOnly(t *testing.T) {
	cfg := config.DefaultConfig()
	h := NewOAuth2Handler(cfg)
	h.sessions["session-id"] = &oauth2Session{
		Username:  "alice",
		Usergroup: "admins",
		createdAt: time.Now(),
	}
	h.pendingStates["csrf-state"] = &oauth2PendingState{
		providerName: "test",
		createdAt:    time.Now(),
	}

	reqWithSession := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	reqWithSession.AddCookie(&http.Cookie{Name: oauthSessionCookieName, Value: "session-id"})
	user := h.CheckUserFromOAuth2Cookie(&authTypes.AuthCheckingContext{Request: reqWithSession})
	assert.NotNil(t, user)
	assert.Equal(t, "alice", user.Username)

	reqWithState := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	reqWithState.AddCookie(&http.Cookie{Name: oauthSessionCookieName, Value: "csrf-state"})
	assert.Nil(t, h.CheckUserFromOAuth2Cookie(&authTypes.AuthCheckingContext{Request: reqWithState}))

	reqLegacyStateCookie := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	reqLegacyStateCookie.AddCookie(&http.Cookie{Name: oauthStateCookieName, Value: "csrf-state"})
	assert.Nil(t, h.CheckUserFromOAuth2Cookie(&authTypes.AuthCheckingContext{Request: reqLegacyStateCookie}))
}
