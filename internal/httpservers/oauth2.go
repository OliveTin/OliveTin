package httpservers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	config "github.com/OliveTin/OliveTin/internal/config"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"io"
	"net/http"
	"time"
)

var (
	registeredStates    = make(map[string]*oauth2State)
	registeredProviders = make(map[string]*oauth2.Config)
)

type oauth2State struct {
	provider  *oauth2.Config
	Username  string
	Usergroup string
}

func assignIfEmpty(target *string, value string) {
	if *target == "" {
		*target = value
	}
}

func completeProviderConfig(providerName string, providerConfig *config.OAuth2Provider) {
	dbConfig, ok := oauth2ProviderDatabase[providerName]

	if ok {
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
		providerConfig, ok := cfg.AuthOAuth2Providers[providerName]

		if !ok {
			return nil, fmt.Errorf("Provider not found in config: %v", providerName)
		}

		completeProviderConfig(providerName, providerConfig)

		config = &oauth2.Config{
			ClientID:     providerConfig.ClientID,
			ClientSecret: providerConfig.ClientSecret,
			Scopes:       providerConfig.Scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  providerConfig.AuthUrl,
				TokenURL: providerConfig.TokenUrl,
			},
			RedirectURL: "http://localhost:1337/oauth/callback",
		}

		registeredProviders[providerName] = config

		log.Debugf("Dumping newly registered provider: %v = %+v", providerName, providerConfig)
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

func setOauthCallbackCookie(w http.ResponseWriter, r *http.Request, name, value string) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   int(time.Hour.Seconds()),
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

	registeredStates[state] = &oauth2State{
		provider: provider,
	}

	if err != nil {
		log.Errorf("Failed to get provider config: %v %v", providerName, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	setOauthCallbackCookie(w, r, "oauth2state", state)

	log.Infof("OAuth2 state: %v mapped to provider %v (found: %v), now redirecting", state, providerName, provider != nil)

	http.Redirect(w, r, provider.AuthCodeURL(state), http.StatusFound)
}

func checkOAuthCallbackCookie(w http.ResponseWriter, r *http.Request) (*oauth2State, bool) {
	state, err := r.Cookie("oauth2state")

	if err != nil {
		log.Errorf("Failed to get state cookie: %v", err)

		http.Error(w, "State not found", http.StatusBadRequest)
		return nil, false
	}

	if r.URL.Query().Get("state") != state.Value {
		log.Errorf("State mismatch: %v != %v", r.URL.Query().Get("state"), state.Value)

		http.Error(w, "State mismatch", http.StatusBadRequest)
		return nil, false
	}

	registeredState, ok := registeredStates[state.Value]

	if !ok {
		log.Errorf("State not found in server: %v", state.Value)

		http.Error(w, "State not found in server", http.StatusBadRequest)
	}

	return registeredState, true
}

func handleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	log.Infof("OAuth2 Callback received")

	registeredState, ok := checkOAuthCallbackCookie(w, r)

	if !ok {
		return
	}

	code := r.FormValue("code")

	log.Debugf("OAuth2 Token Code: %v", code)

	httpClient := &http.Client{Timeout: 2 * time.Second}
	ctx := context.Background()
	ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)

	tok, err := registeredState.provider.Exchange(ctx, code)

	if err != nil {
		log.Errorf("Failed to exchange code: %v", err)
		http.Error(w, "Failed to exchange code", http.StatusBadRequest)
		return
	}

	client := registeredState.provider.Client(ctx, tok)

	registeredState.Username = getUsername(client)

	loginMessage := fmt.Sprintf("Logged in as %v", registeredState.Username)

	log.Infof(loginMessage)

	w.Write([]byte(loginMessage))
}

func getUsername(client *http.Client) string {
	provider := cfg.AuthOAuth2Providers["github"]

	res, err := client.Get(provider.WhoamiUrl)

	if res.StatusCode != http.StatusOK {
		log.Errorf("Failed to get user data: %v", res.StatusCode)
		return ""
	}

	defer res.Body.Close()

	contents, err := io.ReadAll(res.Body)

	var userData map[string]interface{}

	err = json.Unmarshal([]byte(contents), &userData)

	if err != nil {
		log.Errorf("Failed to unmarshal user data: %v", err)

		return ""
	}

	username, ok := userData[provider.UsernameField]

	if !ok {
		log.Errorf("Failed to get username from user data: %v", userData)

		return ""
	}

	return username.(string)
}

func parseOAuth2Cookie(r *http.Request) (string, string) {
	cookie, err := r.Cookie("oauth2state")

	if err != nil {
		log.Warnf("Failed to read OAuth2 cookie: %v", err)
		return "", ""
	}

	serverState, found := registeredStates[cookie.Value]

	if !found {
		log.Warnf("Failed to find OAuth2 state: %v", cookie.Value)
		return "", ""
	}

	return serverState.Username, serverState.Usergroup
}
