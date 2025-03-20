package httpservers

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	config "github.com/OliveTin/OliveTin/internal/config"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"io"
	"net/http"
	"os"
	"time"
)

var (
	registeredStates    = make(map[string]*oauth2State)
	registeredProviders = make(map[string]*oauth2.Config)
)

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

func oauth2Init(cfg *config.Config) {
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

		registeredProviders[providerName] = newConfig

		log.Debugf("Dumping newly registered provider: %v = %+v", providerName, providerConfig)
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

func getOAuth2Config(cfg *config.Config, providerName string) (*oauth2.Config, error) {
	config, ok := registeredProviders[providerName]

	if !ok {
		return nil, fmt.Errorf("Provider not found in config: %v", providerName)
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

func setOAuthCallbackCookie(w http.ResponseWriter, r *http.Request, name, value string) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   31556952, // 1 year
		Secure:   r.TLS != nil,
		HttpOnly: true,
		Path:     "/",
	}

	http.SetCookie(w, cookie)
}

func handleOAuthLogin(w http.ResponseWriter, r *http.Request) {
	state, err := randString(16)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	providerName := r.URL.Query().Get("provider")
	provider, err := getOAuth2Config(cfg, providerName)

	if err != nil {
		log.Errorf("Failed to get provider config: %v %v", providerName, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	registeredStates[state] = &oauth2State{
		providerConfig: provider,
		providerName:   providerName,
		Username:       "",
	}

	setOAuthCallbackCookie(w, r, "olivetin-sid-oauth", state)

	log.Infof("OAuth2 state: %v mapped to provider %v (found: %v), now redirecting", state, providerName, provider != nil)

	http.Redirect(w, r, provider.AuthCodeURL(state), http.StatusFound)
}

func checkOAuthCallbackCookie(w http.ResponseWriter, r *http.Request) (*oauth2State, string, bool) {
	cookie, err := r.Cookie("olivetin-sid-oauth")
	state := cookie.Value

	if err != nil {
		log.Errorf("Failed to get state cookie: %v", err)

		http.Error(w, "State not found", http.StatusBadRequest)
		return nil, state, false
	}

	if r.URL.Query().Get("state") != state {
		log.Errorf("State mismatch: %v != %v", r.URL.Query().Get("state"), state)

		http.Error(w, "State mismatch", http.StatusBadRequest)
		return nil, state, false
	}

	registeredState, ok := registeredStates[state]

	if !ok {
		log.Errorf("State not found in server: %v", state)

		http.Error(w, "State not found in server", http.StatusBadRequest)
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
		log.Errorf("OAuth2 Cert Bundle - failed to append certificates: %v", err)
	}

	return caCertPool
}

func handleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	log.Infof("OAuth2 Callback received")

	registeredState, state, ok := checkOAuthCallbackCookie(w, r)

	if !ok {
		return
	}

	code := r.FormValue("code")

	log.WithFields(log.Fields{
		"state":      state,
		"token-code": code,
	}).Debug("OAuth2 Token Code")

	providerConfig := cfg.AuthOAuth2Providers[registeredState.providerName]

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

	userinfo := getUserInfo(userInfoClient, cfg.AuthOAuth2Providers[registeredState.providerName])

	registeredStates[state].Username = userinfo.Username
	registeredStates[state].Usergroup = userinfo.Usergroup

	for k, v := range registeredStates {
		log.Debugf("states: %+v %+v", k, v)
	}

	loginMessage := fmt.Sprintf("OAuth2 login complete for %v", registeredStates[state].Username)

	log.WithFields(log.Fields{
		"state": state,
	}).Infof(loginMessage)

	http.Redirect(w, r, "/", http.StatusFound)
	w.Write([]byte(loginMessage))
}

type UserInfo struct {
	Username  string
	Usergroup string
}

func getUserInfo(client *http.Client, provider *config.OAuth2Provider) *UserInfo {
	ret := &UserInfo{}

	res, err := client.Get(provider.WhoamiUrl)

	if res.StatusCode != http.StatusOK {
		log.Errorf("Failed to get user data: %v", res.StatusCode)
		return ret
	}

	defer res.Body.Close()

	contents, err := io.ReadAll(res.Body)

	var userData map[string]interface{}

	err = json.Unmarshal([]byte(contents), &userData)

	if err != nil {
		log.Errorf("Failed to unmarshal user data: %v", err)

		return ret
	}

	ret.Username = getDataField(userData, provider.UsernameField)
	ret.Usergroup = getDataField(userData, provider.UserGroupField)

	return ret
}

func getDataField(data map[string]interface{}, field string) string {
	if field == "" {
		return ""
	}

	val, ok := data[field]

	if !ok {
		log.Errorf("Failed to get field from user data: %v / %v", data, field)

		return ""
	}

	return val.(string)
}

func parseOAuth2Cookie(r *http.Request) (string, string, string) {
	cookie, err := r.Cookie("olivetin-sid-oauth")

	if err != nil {
		log.Warnf("Failed to read OAuth2 cookie: %v", err)
		return "", "", ""
	}

	if cookie.Value == "" {
		return "", "", ""
	}

	serverState, found := registeredStates[cookie.Value]

	if !found {
		log.WithFields(log.Fields{
			"sid":      cookie.Value,
			"provider": "oauth2",
		}).Warnf("Stale session")

		return "", "", cookie.Value
	}

	log.Debugf("Found OAuth2 state: %+v", serverState)

	return serverState.Username, serverState.Usergroup, cookie.Value
}
