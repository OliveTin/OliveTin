package auth

import (
	"os"
	"sync"
	"time"

	"github.com/OliveTin/OliveTin/internal/config"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// Session management for user authentication
type UserSession struct {
	Username string
	Expiry   int64
}

type SessionProvider struct {
	Sessions map[string]*UserSession
}

type SessionStorage struct {
	Providers map[string]*SessionProvider
}

var (
	sessionStorage       *SessionStorage
	sessionStorageMutex  sync.RWMutex
	oauth2SessionRevoker func(sid string)
)

func init() {
	sessionStorage = &SessionStorage{
		Providers: make(map[string]*SessionProvider),
	}
}

// RegisterUserSession registers a user session
func RegisterUserSession(cfg *config.Config, provider string, sid string, username string) {
	sessionStorageMutex.Lock()
	defer sessionStorageMutex.Unlock()

	if sessionStorage.Providers[provider] == nil {
		sessionStorage.Providers[provider] = &SessionProvider{
			Sessions: make(map[string]*UserSession),
		}
	}

	if sessionStorage.Providers == nil {
		sessionStorage.Providers = make(map[string]*SessionProvider)
	}

	sessionStorage.Providers[provider].Sessions[sid] = &UserSession{
		Username: username,
		Expiry:   time.Now().Unix() + 31556952, // 1 year
	}

	saveUserSessions(cfg)
}

// RegisterOAuth2SessionRevoker registers a callback to revoke OAuth2 sessions on logout.
// OAuth2 uses its own session storage; the API calls this when provider is oauth2.
func RegisterOAuth2SessionRevoker(fn func(sid string)) {
	oauth2SessionRevoker = fn
}

// RevokeSessionForProvider invalidates the session for the given provider and SID (e.g. on logout).
// Local auth uses shared SessionStorage; OAuth2 uses a separate storage and revoker.
func RevokeSessionForProvider(cfg *config.Config, provider string, sid string) {
	if sid == "" {
		return
	}
	if provider == "oauth2" && oauth2SessionRevoker != nil {
		oauth2SessionRevoker(sid)
		return
	}
	RevokeUserSession(cfg, provider, sid)
}

// RevokeUserSession removes a session from storage so it can no longer be used (e.g. on logout).
func RevokeUserSession(cfg *config.Config, provider string, sid string) {
	sessionStorageMutex.Lock()
	defer sessionStorageMutex.Unlock()

	if sessionStorage.Providers[provider] != nil {
		delete(sessionStorage.Providers[provider].Sessions, sid)
		if cfg != nil {
			saveUserSessions(cfg)
		}
	}
}

// GetUserSession retrieves a user session
func GetUserSession(provider string, sid string) *UserSession {
	sessionStorageMutex.Lock()
	defer sessionStorageMutex.Unlock()

	if sessionStorage.Providers[provider] == nil {
		return nil
	}

	session := sessionStorage.Providers[provider].Sessions[sid]
	if session == nil {
		return nil
	}

	if session.Expiry < time.Now().Unix() {
		delete(sessionStorage.Providers[provider].Sessions, sid)
		return nil
	}

	return session
}

// LoadUserSessions loads sessions from disk
func LoadUserSessions(cfg *config.Config) {
	sessionStorageMutex.Lock()
	defer sessionStorageMutex.Unlock()

	data, err := os.ReadFile(cfg.GetDir() + "/sessions.yaml")
	if err != nil {
		logrus.WithError(err).Warn("Failed to read sessions.yaml file")
		ensureEmptySessionStorage()
		return
	}

	if err := yaml.Unmarshal(data, &sessionStorage); err != nil {
		logrus.WithError(err).Error("Failed to unmarshal sessions.yaml")
		ensureEmptySessionStorage()
		return
	}

	ensureEmptySessionStorage()
}

func ensureEmptySessionStorage() {
	if sessionStorage == nil {
		sessionStorage = &SessionStorage{Providers: make(map[string]*SessionProvider)}
	}
	if sessionStorage.Providers == nil {
		sessionStorage.Providers = make(map[string]*SessionProvider)
	}
}

func saveUserSessions(cfg *config.Config) {
	out, err := yaml.Marshal(sessionStorage)
	if err != nil {
		logrus.WithError(err).Error("Failed to marshal session storage")
		return
	}

	err = os.WriteFile(cfg.GetDir()+"/sessions.yaml", out, 0600)
	if err != nil {
		logrus.WithError(err).Error("Failed to write sessions.yaml file")
		return
	}
}
