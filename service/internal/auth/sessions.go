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
	sessionStorage      *SessionStorage
	sessionStorageMutex sync.RWMutex
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
		// Initialize empty session storage if file doesn't exist
		if sessionStorage == nil {
			sessionStorage = &SessionStorage{
				Providers: make(map[string]*SessionProvider),
			}
		}
		return
	}

	err = yaml.Unmarshal(data, &sessionStorage)
	if err != nil {
		logrus.WithError(err).Error("Failed to unmarshal sessions.yaml")
		// Initialize empty session storage if unmarshal fails
		if sessionStorage == nil {
			sessionStorage = &SessionStorage{
				Providers: make(map[string]*SessionProvider),
			}
		}
		return
	}

	// Ensure sessionStorage and Providers are properly initialized
	if sessionStorage == nil {
		sessionStorage = &SessionStorage{
			Providers: make(map[string]*SessionProvider),
		}
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
