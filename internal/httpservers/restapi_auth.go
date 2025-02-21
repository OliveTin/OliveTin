package httpservers

import (
	"github.com/OliveTin/OliveTin/internal/config"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
)

type UserSession struct {
	Username string
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
	delete(sessionStorage.Providers[provider].Sessions, sid)

	saveUserSessions()
}

func registerUserSession(provider string, sid string, username string) {
	sessionStorage.Providers[provider].Sessions[sid] = &UserSession{
		Username: username,
	}

	saveUserSessions()
}

func saveUserSessions() {
	configDir := cfg.GetDir()
	filename := configDir + "/sessions.db.yaml"

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		_, err := os.Create(filename)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Errorf("Failed to create %v", filename)
			return
		}
	}

	out, _ := yaml.Marshal(sessionStorage)
	err := os.WriteFile(filename, out, 0644)

	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Errorf("Failed to write session to %v", filename)
		return
	}
}

func loadUserSessions() {
	registerSessionProviders()

	configDir := cfg.GetDir()
	filename := filepath.Join(configDir, "sessions.db.yaml")

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
			"provider": "local",
		}).Warnf("Stale session")
		return nil
	}

	user := cfg.FindUserByUsername(session.Username)

	if user == nil {
		log.WithFields(log.Fields{
			"sid":      sid,
			"provider": "local",
		}).Warnf("User not found")
		return nil
	}

	return user
}
