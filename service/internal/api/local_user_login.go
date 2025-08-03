package api

import (
	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/alexedwards/argon2id"
	log "github.com/sirupsen/logrus"
	"runtime"
)

var defaultParams = argon2id.Params{
	Memory:      64 * 1024,
	Iterations:  4,
	Parallelism: uint8(runtime.NumCPU()),
	SaltLength:  16,
	KeyLength:   32,
}

func createHash(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, &defaultParams)

	if err != nil {
		log.Fatal("Error creating hash: ", err)
		return "", err
	}

	return hash, nil
}

func comparePasswordAndHash(password, hash string) bool {
	match, err := argon2id.ComparePasswordAndHash(password, hash)

	if err != nil {
		log.Errorf("Error comparing password and hash: %v", err)
		return false
	}

	return match
}

func checkUserPassword(cfg *config.Config, username, password string) bool {
	for _, user := range cfg.AuthLocalUsers.Users {
		if user.Username == username {
			match := comparePasswordAndHash(password, user.Password)

			if match {
				return true
			} else {
				log.WithFields(log.Fields{
					"username": username,
				}).Warn("Password does not match for user")

				return false
			}
		}
	}

	log.WithFields(log.Fields{
		"username": username,
	}).Warn("Failed to check password for user, as username was not found")

	return false
}
