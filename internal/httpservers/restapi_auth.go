package httpservers

import (
	"github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/filehelper"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var sessionStorageMutex sync.Mutex

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
	sessionStorage *SessionStorage
)

func registerSessionProviders() {
	sessionStorage = &SessionStorage{
		Providers: make(map[string]*SessionProvider),
	}

	registerSessionProvider("local")
	registerSessionProvider("oauth2")
}

func registerSessionProvider(provider string) {
	sessionStorage.Providers[provider] = &SessionProvider{
		Sessions: make(map[string]*UserSession),
	}
}

func deleteLocalUserSession(provider string, sid string) {
	sessionStorageMutex.Lock()

	deleteLocalUserSessionBatch(provider, sid)

	sessionStorageMutex.Unlock()

	saveUserSessions()
}

func deleteLocalUserSessionBatch(provider string, sid string) {
	log.WithFields(log.Fields{
		"sid":      sid,
		"provider": provider,
	}).Debug("Deleting user session")

	if _, ok := sessionStorage.Providers[provider]; !ok {
		return
	}

	delete(sessionStorage.Providers[provider].Sessions, sid)
}

func registerUserSession(provider string, sid string, username string) {
	sessionStorageMutex.Lock()
	sessionStorage.Providers[provider].Sessions[sid] = &UserSession{
		Username: username,
		Expiry:   time.Now().Unix() + 31556952, // 1 year
	}
	sessionStorageMutex.Unlock()

	saveUserSessions()
}

func saveUserSessions() {
	sessionStorageMutex.Lock()
	defer sessionStorageMutex.Unlock()

	filename := filepath.Join(cfg.GetDir(), "sessions.db.yaml")

	out, err := yaml.Marshal(sessionStorage)

	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Errorf("Failed to marshal session data to %v", filename)
		return
	}

	filehelper.WriteFile(filename, out)
}

func loadUserSessions() {
	registerSessionProviders()

	filename := filepath.Join(cfg.GetDir(), "sessions.db.yaml")

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return
	}

	data, err := os.ReadFile(filename)

	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Errorf("Failed to read %v", filename)
		return
	}

	err = yaml.Unmarshal(data, &sessionStorage)

	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to unmarshal sessions.local.db")
		return
	}

	deleteExpiredSessions()
}

func deleteExpiredSessions() {
	sessionStorageMutex.Lock()

	for provider, sessions := range sessionStorage.Providers {
		for sid, session := range sessions.Sessions {
			if session.Expiry < time.Now().Unix() {
				deleteLocalUserSessionBatch(provider, sid)
			}
		}
	}

	sessionStorageMutex.Unlock()

	saveUserSessions()
}

func getUserFromSession(providerName string, sid string) *config.LocalUser {
	provider, ok := sessionStorage.Providers[providerName]

	if !ok {
		log.WithFields(log.Fields{
			"provider": providerName,
		}).Warnf("Provider not found")
		return nil
	}

	session, ok := provider.Sessions[sid]

	if !ok {
		log.WithFields(log.Fields{
			"sid":      sid,
			"provider": providerName,
		}).Warnf("Stale session")
		return nil
	}

	user := cfg.FindUserByUsername(session.Username)

	if user == nil {
		log.WithFields(log.Fields{
			"sid":      sid,
			"provider": providerName,
		}).Warnf("User not found")
		return nil
	}

	return user
}
