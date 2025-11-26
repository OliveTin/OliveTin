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
	"time"

	authTypes "github.com/OliveTin/OliveTin/internal/auth/authpublic"
	config "github.com/OliveTin/OliveTin/internal/config"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type OAuth2Handler struct {
	cfg                 *config.Config
	registeredStates    map[string]*oauth2State
	registeredProviders map[string]*oauth2.Config
}

func NewOAuth2Handler(cfg *config.Config) *OAuth2Handler {
	h := &OAuth2Handler{
		cfg: cfg,
	}

	h.registeredStates = make(map[string]*oauth2State)
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

type oauth2State struct {
	providerConfig *oauth2.Config
	providerName   string
	Username       string
	Usergroup      string
}

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

func (h *OAuth2Handler) setOAuthCallbackCookie(w http.ResponseWriter, r *http.Request, name, value string) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   900, // 15 minutes
		Secure:   r.TLS != nil,
		HttpOnly: true,
		Path:     "/",
	}

	http.SetCookie(w, cookie)
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

	h.registeredStates[state] = &oauth2State{
		providerConfig: provider,
		providerName:   providerName,
		Username:       "",
	}

	h.setOAuthCallbackCookie(w, r, "olivetin-sid-oauth", state)

	log.Infof("OAuth2 state: %v mapped to provider %v (found: %v), now redirecting", state, providerName, provider != nil)

	http.Redirect(w, r, provider.AuthCodeURL(state), http.StatusFound)
}

func (h *OAuth2Handler) validateStateMatch(queryState, cookieState string) bool {
	return queryState == cookieState
}

func (h *OAuth2Handler) checkOAuthCallbackCookie(w http.ResponseWriter, r *http.Request) (*oauth2State, string, bool) {
	cookie, err := r.Cookie("olivetin-sid-oauth")
	if err != nil {
		log.Errorf("Failed to get state cookie: %v", err)
		http.Error(w, "State not found", http.StatusBadRequest)
		return nil, "", false
	}

	state := cookie.Value

	if !h.validateStateMatch(r.URL.Query().Get("state"), state) {
		log.Errorf("State mismatch: %v != %v", r.URL.Query().Get("state"), state)
		http.Error(w, "State mismatch", http.StatusBadRequest)
		return nil, state, false
	}

	registeredState, ok := h.registeredStates[state]
	if !ok {
		log.Errorf("State not found in server: %v", state)
		http.Error(w, "State not found in server", http.StatusBadRequest)
		return nil, state, false
	}

	return registeredState, state, true
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

func (h *OAuth2Handler) HandleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	log.Infof("OAuth2 Callback received")

	registeredState, state, ok := h.checkOAuthCallbackCookie(w, r)

	if !ok {
		return
	}

	code := r.FormValue("code")

	log.WithFields(log.Fields{
		"state":      state,
		"token-code": code,
	}).Debug("OAuth2 Token Code")

	providerConfig := h.cfg.AuthOAuth2Providers[registeredState.providerName]

	clientSettings := getOAuth2HttpClient(providerConfig)

	exchangeClient := &http.Client{
		Transport: clientSettings.Transport,
		Timeout:   clientSettings.Timeout,
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, oauth2.HTTPClient, exchangeClient)

	tok, err := registeredState.providerConfig.Exchange(ctx, code)

	if err != nil {
		log.Errorf("Failed to exchange code: %v", err)
		http.Error(w, "Failed to exchange code", http.StatusBadRequest)
		return
	}

	userInfoClient := &http.Client{
		Transport: &oauth2.Transport{
			Source: registeredState.providerConfig.TokenSource(ctx, tok),
			Base:   clientSettings.Transport,
		},
		Timeout: clientSettings.Timeout,
	}

	userinfo := getUserInfo(h.cfg, userInfoClient, h.cfg.AuthOAuth2Providers[registeredState.providerName])

	h.registeredStates[state].Username = userinfo.Username
	h.registeredStates[state].Usergroup = userinfo.Usergroup

	http.Redirect(w, r, "/", http.StatusFound)
}

type UserInfo struct {
	Username  string
	Usergroup string
}

//gocyclo:ignore
func getUserInfo(cfg *config.Config, client *http.Client, provider *config.OAuth2Provider) *UserInfo {
	ret := &UserInfo{}

	res, err := client.Get(provider.WhoamiUrl)

	if err != nil {
		log.Errorf("Failed to get user data: %v", err)
		return ret
	}

	if res.StatusCode != http.StatusOK {
		log.Errorf("Failed to get user data: %v", res.StatusCode)
		return ret
	}

	defer res.Body.Close()

	contents, err := io.ReadAll(res.Body)

	if err != nil {
		log.Errorf("Failed to read user data: %v", err)
		return ret
	}

	var userData map[string]any

	if cfg.InsecureAllowDumpOAuth2UserData {
		log.Debugf("OAuth2 User Data: %v+", string(contents))
	}

	err = json.Unmarshal([]byte(contents), &userData)

	if err != nil {
		log.Errorf("Failed to unmarshal user data: %v", err)

		return ret
	}

	ret.Username = getDataField(userData, provider.UsernameField)
	ret.Usergroup = getDataField(userData, provider.UserGroupField)

	return ret
}

func getDataField(data map[string]any, field string) string {
	if field == "" {
		return ""
	}

	val, ok := data[field]

	if !ok {
		log.Errorf("Failed to get field from user data: %v / %v", data, field)

		return ""
	}

	stringVal, ok := val.(string)

	if !ok {
		log.Errorf("Field %v is not a string: %v", field, val)
		return ""
	}

	return stringVal
}

func (h *OAuth2Handler) CheckUserFromOAuth2Cookie(context *authTypes.AuthCheckingContext) *authTypes.AuthenticatedUser {
	cookie, err := context.Request.Cookie("olivetin-sid-oauth")

	user := &authTypes.AuthenticatedUser{}

	if err != nil {
		log.Warnf("Failed to read OAuth2 cookie: %v", err)
		return nil
	}

	if cookie.Value == "" {
		return nil
	}

	serverState, found := h.registeredStates[cookie.Value]

	if !found {
		log.WithFields(log.Fields{
			"sid":      cookie.Value,
			"provider": "oauth2",
		}).Warnf("Stale session")

		return nil
	}

	user.Username = serverState.Username
	user.UsergroupLine = serverState.Usergroup
	user.Provider = "oauth2"
	user.SID = cookie.Value

	return user
}
