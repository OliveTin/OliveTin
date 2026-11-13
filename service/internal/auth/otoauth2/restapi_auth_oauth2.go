package otoauth2

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	authTypes "github.com/OliveTin/OliveTin/internal/auth/authpublic"
	config "github.com/OliveTin/OliveTin/internal/config"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type OAuth2Handler struct {
	cfg                 *config.Config
	pendingStates       map[string]*oauth2PendingState
	sessions            map[string]*oauth2Session
	registeredProviders map[string]*oauth2.Config
	mu                  sync.RWMutex
}

func NewOAuth2Handler(cfg *config.Config) *OAuth2Handler {
	h := &OAuth2Handler{
		cfg: cfg,
	}

	h.pendingStates = make(map[string]*oauth2PendingState)
	h.sessions = make(map[string]*oauth2Session)
	h.registeredProviders = make(map[string]*oauth2.Config)

	for providerName, providerConfig := range cfg.AuthOAuth2Providers {
		completeProviderConfig(providerName, providerConfig)

		newConfig := &oauth2.Config{
			ClientID:     providerConfig.ClientID,
			ClientSecret: providerConfig.ClientSecret,
			Scopes:       providerConfig.Scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  providerConfig.AuthUrl,
				TokenURL: providerConfig.TokenUrl,
			},
			RedirectURL: cfg.AuthOAuth2RedirectURL,
		}

		h.registeredProviders[providerName] = newConfig

		log.Debugf("Dumping newly registered provider: %v = %+v", providerName, providerConfig)
	}

	return h
}

type oauth2PendingState struct {
	createdAt      time.Time
	providerConfig *oauth2.Config
	providerName   string
}

type oauth2Session struct {
	createdAt    time.Time
	Username     string
	Usergroup    string
	providerName string
}

const (
	oauthStateCookieName   = "olivetin-oauth-state"
	oauthSessionCookieName = "olivetin-sid-oauth"
	oauthStateMaxAge       = 900 // seconds; CSRF pending window
	oauthSessionMaxAge     = 900 // seconds; authenticated session lifetime
	oauthStateMaxEntries   = 10000
	oauthSessionMaxEntries = 10000
)

func assignIfEmpty(target *string, value string) {
	if *target == "" {
		*target = value
	}
}

func completeProviderConfig(providerName string, providerConfig *config.OAuth2Provider) {
	dbConfig, ok := oauth2ProviderDatabase[providerName]

	if ok {
		assignIfEmpty(&providerConfig.Name, dbConfig.Name)
		assignIfEmpty(&providerConfig.Title, dbConfig.Title)
		assignIfEmpty(&providerConfig.WhoamiUrl, dbConfig.WhoamiUrl)
		assignIfEmpty(&providerConfig.TokenUrl, dbConfig.TokenUrl)
		assignIfEmpty(&providerConfig.AuthUrl, dbConfig.AuthUrl)
		assignIfEmpty(&providerConfig.Icon, dbConfig.Icon)
		assignIfEmpty(&providerConfig.UsernameField, dbConfig.UsernameField)

		if providerConfig.Scopes == nil {
			providerConfig.Scopes = dbConfig.Scopes
		}
	} else {
		log.Warnf("Provider not found in database: %v", providerName)
	}
}

func (h *OAuth2Handler) getOAuth2Config(providerName string) (*oauth2.Config, error) {
	config, ok := h.registeredProviders[providerName]

	if !ok {
		return nil, fmt.Errorf("provider not found in config: %v", providerName)
	}

	return config, nil
}

func randString(nByte int) (string, error) {
	b := make([]byte, nByte)

	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), nil
}

func (h *OAuth2Handler) cookieSecure(r *http.Request) bool {
	useTLS := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
	return useTLS || h.cfg.Security.ForceSecureCookies
}

func (h *OAuth2Handler) setOAuthCookie(w http.ResponseWriter, r *http.Request, name, value string, maxAge int) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   maxAge,
		Secure:   h.cookieSecure(r),
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, cookie)
}

func (h *OAuth2Handler) clearOAuthCookie(w http.ResponseWriter, r *http.Request, name string) {
	h.setOAuthCookie(w, r, name, "", -1)
}

func (h *OAuth2Handler) deletePendingStateLocked(state string) {
	delete(h.pendingStates, state)
}

func (h *OAuth2Handler) sweepExpiredOAuthEntriesLocked(now time.Time) {
	h.sweepExpiredPendingStatesLocked(now)
	h.sweepExpiredSessionsLocked(now)
}

func (h *OAuth2Handler) sweepExpiredPendingStatesLocked(now time.Time) {
	pendingCutoff := now.Add(-oauthStateMaxAge * time.Second)
	for state, entry := range h.pendingStates {
		if entry.createdAt.Before(pendingCutoff) {
			delete(h.pendingStates, state)
		}
	}
}

func (h *OAuth2Handler) sweepExpiredSessionsLocked(now time.Time) {
	sessionCutoff := now.Add(-oauthSessionMaxAge * time.Second)
	for sid, entry := range h.sessions {
		if entry.createdAt.Before(sessionCutoff) {
			delete(h.sessions, sid)
		}
	}
}

func (h *OAuth2Handler) HandleOAuthLogin(w http.ResponseWriter, r *http.Request) {
	state, err := randString(16)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	providerName := r.URL.Query().Get("provider")
	provider, err := h.getOAuth2Config(providerName)

	if err != nil {
		log.Errorf("Failed to get provider config: %v %v", providerName, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.mu.Lock()
	h.sweepExpiredOAuthEntriesLocked(time.Now())
	if len(h.pendingStates) >= oauthStateMaxEntries {
		h.mu.Unlock()
		http.Error(w, "OAuth login temporarily unavailable", http.StatusServiceUnavailable)
		return
	}
	h.pendingStates[state] = &oauth2PendingState{
		providerConfig: provider,
		providerName:   providerName,
		createdAt:      time.Now(),
	}
	h.mu.Unlock()

	// CSRF double-submit cookie only — never reuse this value as the session ID.
	h.setOAuthCookie(w, r, oauthStateCookieName, state, oauthStateMaxAge)

	log.Infof("OAuth2 login redirect for provider %v", providerName)

	http.Redirect(w, r, provider.AuthCodeURL(state), http.StatusFound)
}

func (h *OAuth2Handler) validateStateMatch(queryState, cookieState string) bool {
	return queryState != "" && queryState == cookieState
}

func pendingStateExpired(pending *oauth2PendingState, now time.Time) bool {
	return pending.createdAt.Before(now.Add(-oauthStateMaxAge * time.Second))
}

func sessionExpired(session *oauth2Session, now time.Time) bool {
	return session.createdAt.Before(now.Add(-oauthSessionMaxAge * time.Second))
}

func (h *OAuth2Handler) consumePendingOAuthState(w http.ResponseWriter, r *http.Request) (*oauth2PendingState, string, bool) {
	cookie, err := r.Cookie(oauthStateCookieName)
	if err != nil {
		log.Errorf("Failed to get OAuth state cookie: %v", err)
		http.Error(w, "State not found", http.StatusBadRequest)
		return nil, "", false
	}

	state := cookie.Value
	queryState := r.URL.Query().Get("state")

	if !h.validateStateMatch(queryState, state) {
		// Keep the cookie's pending state: an unrelated callback (different or
		// missing query state) must not invalidate an in-flight login.
		log.Errorf("OAuth2 state mismatch between query and cookie")
		http.Error(w, "State mismatch", http.StatusBadRequest)
		return nil, state, false
	}

	return h.loadValidPendingOAuthState(w, state)
}

func (h *OAuth2Handler) loadValidPendingOAuthState(w http.ResponseWriter, state string) (*oauth2PendingState, string, bool) {
	h.mu.RLock()
	pending, ok := h.pendingStates[state]
	expired := ok && pendingStateExpired(pending, time.Now())
	h.mu.RUnlock()

	if !ok {
		log.Errorf("OAuth2 pending state not found on server")
		http.Error(w, "State not found in server", http.StatusBadRequest)
		return nil, state, false
	}

	if expired {
		h.mu.Lock()
		h.deletePendingStateLocked(state)
		h.mu.Unlock()
		log.Errorf("OAuth2 pending state expired")
		http.Error(w, "State expired", http.StatusBadRequest)
		return nil, state, false
	}

	return pending, state, true
}

type HttpClientSettings struct {
	Transport *http.Transport
	Timeout   time.Duration
}

func getOAuth2HttpClient(providerConfig *config.OAuth2Provider) *HttpClientSettings {
	config := &HttpClientSettings{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: providerConfig.InsecureSkipVerify},
		},
		Timeout: time.Duration(min(3, providerConfig.CallbackTimeout)) * time.Second,
	}

	if providerConfig.CertBundlePath != "" {
		config.Transport.TLSClientConfig.RootCAs = getOAuthCertBundle(providerConfig)
	}

	return config
}

func getOAuthCertBundle(providerConfig *config.OAuth2Provider) *x509.CertPool {
	caCert, err := os.ReadFile(providerConfig.CertBundlePath)

	if err != nil {
		log.Errorf("OAuth2 Cert Bundle - failed to read file: %v", err)

		return nil
	}

	caCertPool := x509.NewCertPool()

	if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
		log.Errorf("OAuth2 Cert Bundle - failed to append certificates from PEM")
	}

	return caCertPool
}

func (h *OAuth2Handler) exchangeOAuthCode(ctx context.Context, providerConfig *oauth2.Config, code string, clientSettings *HttpClientSettings) (*oauth2.Token, error) {
	exchangeClient := &http.Client{
		Transport: clientSettings.Transport,
		Timeout:   clientSettings.Timeout,
	}

	ctx = context.WithValue(ctx, oauth2.HTTPClient, exchangeClient)

	return providerConfig.Exchange(ctx, code)
}

func (h *OAuth2Handler) createUserInfoClient(ctx context.Context, providerConfig *oauth2.Config, tok *oauth2.Token, clientSettings *HttpClientSettings) *http.Client {
	return &http.Client{
		Transport: &oauth2.Transport{
			Source: providerConfig.TokenSource(ctx, tok),
			Base:   clientSettings.Transport,
		},
		Timeout: clientSettings.Timeout,
	}
}

func (h *OAuth2Handler) computeUsergroup(userinfo *UserInfo, providerConfig *config.OAuth2Provider) string {
	usergroup := userinfo.Usergroup
	if providerConfig != nil && providerConfig.AddToUsergroup != "" {
		if usergroup != "" {
			sep := h.cfg.AuthHttpHeaderUserGroupSep
			if sep == "" {
				sep = " "
			}
			usergroup = usergroup + sep + providerConfig.AddToUsergroup
		} else {
			usergroup = providerConfig.AddToUsergroup
		}
	}
	return usergroup
}

func (h *OAuth2Handler) takeValidPendingStateLocked(state string, now time.Time) (*oauth2PendingState, error) {
	pendingEntry, ok := h.pendingStates[state]
	if !ok {
		return nil, fmt.Errorf("pending OAuth state no longer exists")
	}
	if pendingStateExpired(pendingEntry, now) {
		h.deletePendingStateLocked(state)
		return nil, fmt.Errorf("pending OAuth state expired")
	}

	h.deletePendingStateLocked(state)
	return pendingEntry, nil
}

func (h *OAuth2Handler) issueOAuthSession(state, username, usergroup string) (string, error) {
	sessionID, err := randString(16)
	if err != nil {
		return "", err
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	h.sweepExpiredOAuthEntriesLocked(time.Now())
	if len(h.sessions) >= oauthSessionMaxEntries {
		return "", fmt.Errorf("OAuth session capacity exceeded")
	}

	pendingEntry, err := h.takeValidPendingStateLocked(state, time.Now())
	if err != nil {
		return "", err
	}

	h.sessions[sessionID] = &oauth2Session{
		createdAt:    time.Now(),
		Username:     username,
		Usergroup:    usergroup,
		providerName: pendingEntry.providerName,
	}

	return sessionID, nil
}

func (h *OAuth2Handler) HandleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	log.Infof("OAuth2 Callback received")

	pending, state, ok := h.consumePendingOAuthState(w, r)
	if !ok {
		return
	}

	code := r.FormValue("code")
	log.Debug("OAuth2 token exchange starting")

	providerConfig := h.cfg.AuthOAuth2Providers[pending.providerName]
	clientSettings := getOAuth2HttpClient(providerConfig)

	ctx := context.Background()
	tok, err := h.exchangeOAuthCode(ctx, pending.providerConfig, code, clientSettings)
	if err != nil {
		log.Errorf("Failed to exchange code: %v", err)
		h.mu.Lock()
		h.deletePendingStateLocked(state)
		h.mu.Unlock()
		http.Error(w, "Failed to exchange code", http.StatusBadRequest)
		return
	}

	userInfoClient := h.createUserInfoClient(ctx, pending.providerConfig, tok, clientSettings)
	userinfo := getUserInfo(h.cfg, userInfoClient, providerConfig)
	if userinfo.Username == "" {
		h.mu.Lock()
		h.deletePendingStateLocked(state)
		h.mu.Unlock()
		log.Errorf("OAuth2 user info returned empty username")
		http.Error(w, "Failed to get user info", http.StatusBadRequest)
		return
	}

	usergroup := h.computeUsergroup(userinfo, providerConfig)

	sessionID, err := h.issueOAuthSession(state, userinfo.Username, usergroup)
	if err != nil {
		log.Errorf("Failed to issue OAuth session: %v", err)
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	h.clearOAuthCookie(w, r, oauthStateCookieName)
	h.setOAuthCookie(w, r, oauthSessionCookieName, sessionID, oauthSessionMaxAge)

	http.Redirect(w, r, "/", http.StatusFound)
}

type UserInfo struct {
	Username  string
	Usergroup string
}

//gocyclo:ignore
func getUserInfo(cfg *config.Config, client *http.Client, provider *config.OAuth2Provider) *UserInfo {
	ret := &UserInfo{}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, provider.WhoamiUrl, nil)

	if err != nil {
		log.Error("Could not construct user data request", err)
		return ret
	}

	res, err := client.Do(req)

	if err != nil {
		log.Errorf("Failed to get user data: %v", err)
		return ret
	}

	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		log.Errorf("Failed to get user data: %v", res.StatusCode)
		return ret
	}

	contents, err := io.ReadAll(res.Body)

	if err != nil {
		log.Errorf("Failed to read user data: %v", err)
		return ret
	}

	var userData map[string]any

	if cfg.InsecureAllowDumpOAuth2UserData {
		log.Debugf("OAuth2 User Data: %v+", string(contents))
	}

	err = json.Unmarshal(contents, &userData)

	if err != nil {
		log.Errorf("Failed to unmarshal user data: %v", err)

		return ret
	}

	ret.Username = getDataField(userData, provider.UsernameField)
	ret.Usergroup = getGroupField(userData, provider.UserGroupField, cfg.AuthHttpHeaderUserGroupSep)

	return ret
}

func lookupRawField(data map[string]any, field string) (any, bool) {
	if field == "" {
		return nil, false
	}

	val, ok := data[field]

	if !ok {
		log.Errorf("Failed to get field from user data: %v / %v", data, field)

		return nil, false
	}

	return val, true
}

func getDataField(data map[string]any, field string) string {
	val, ok := lookupRawField(data, field)

	if !ok {
		return ""
	}

	stringVal, ok := val.(string)

	if !ok {
		log.Errorf("Field %v is not a string: %v", field, val)
		return ""
	}

	return stringVal
}

// getGroupField reads a userinfo field that may be either a single string
// group name, or a JSON array of group names (e.g. an OIDC "groups" claim).
// Array elements that aren't strings are skipped and logged, rather than
// discarding the whole claim.
func getGroupField(data map[string]any, field string, sep string) string {
	val, found := lookupRawField(data, field)

	if !found {
		return ""
	}

	if stringVal, isString := val.(string); isString {
		return stringVal
	}

	arrayVal, isArray := val.([]any)

	if !isArray {
		log.Errorf("Field %v is not a string or array: %v", field, val)
		return ""
	}

	return joinGroupArray(arrayVal, field, sep)
}

// joinGroupArray joins the string elements of a group claim array using sep,
// defaulting to a space. Non-string elements are skipped and logged, rather
// than discarding the whole claim.
func joinGroupArray(arrayVal []any, field string, sep string) string {
	groups := make([]string, 0, len(arrayVal))

	for _, v := range arrayVal {
		strVal, isString := v.(string)

		if !isString {
			log.Warnf("Skipping non-string group entry in field %v: %v", field, v)
			continue
		}

		groups = append(groups, strVal)
	}

	if sep == "" {
		sep = " "
	}

	return strings.Join(groups, sep)
}

func (h *OAuth2Handler) lookupOAuth2UserBySession(sessionID string) (*authTypes.AuthenticatedUser, bool) {
	h.mu.RLock()
	session, found := h.sessions[sessionID]
	if !found || session.Username == "" {
		h.mu.RUnlock()
		return nil, false
	}
	if sessionExpired(session, time.Now()) {
		h.mu.RUnlock()
		h.deleteSessionIfExpired(sessionID)
		return nil, false
	}

	user := &authTypes.AuthenticatedUser{
		Username:      session.Username,
		UsergroupLine: session.Usergroup,
		Provider:      "oauth2",
		SID:           sessionID,
	}
	h.mu.RUnlock()
	return user, true
}

func (h *OAuth2Handler) deleteSessionIfExpired(sessionID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	session, found := h.sessions[sessionID]
	if found && sessionExpired(session, time.Now()) {
		delete(h.sessions, sessionID)
	}
}

func (h *OAuth2Handler) RevokeSession(sid string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.sessions, sid)
}

func (h *OAuth2Handler) CheckUserFromOAuth2Cookie(context *authTypes.AuthCheckingContext) *authTypes.AuthenticatedUser {
	cookie, err := context.Request.Cookie(oauthSessionCookieName)
	if err != nil || cookie.Value == "" {
		return nil
	}

	user, found := h.lookupOAuth2UserBySession(cookie.Value)
	if !found {
		log.WithFields(log.Fields{
			"provider": "oauth2",
		}).Warnf("Stale OAuth2 session")
		return nil
	}
	return user
}
