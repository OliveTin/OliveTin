package api

import (
	"crypto/rand"
	"errors"
	"runtime"
	"sync"

	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/alexedwards/argon2id"
	log "github.com/sirupsen/logrus"
)

var ErrArgon2Busy = errors.New("too many concurrent password operations")

const argon2MaxConcurrent = 10

var argon2Sem = make(chan struct{}, argon2MaxConcurrent)

var defaultParams = argon2id.Params{
	Memory:      64 * 1024,
	Iterations:  4,
	Parallelism: uint8(runtime.NumCPU()),
	SaltLength:  16,
	KeyLength:   32,
}

func CreateHash(password string) (string, error) {
	select {
	case argon2Sem <- struct{}{}:
		defer func() { <-argon2Sem }()
	default:
		return "", ErrArgon2Busy
	}
	hash, err := argon2id.CreateHash(password, &defaultParams)

	if err != nil {
		log.Warnf("Error creating hash: %v", err)
		return "", err
	}

	return hash, nil
}

func createHash(password string) (string, error) {
	return CreateHash(password)
}

func comparePasswordAndHash(password, hash string) (bool, error) {
	select {
	case argon2Sem <- struct{}{}:
		defer func() { <-argon2Sem }()
	default:
		return false, ErrArgon2Busy
	}
	match, err := argon2id.ComparePasswordAndHash(password, hash)

	if err != nil {
		log.Errorf("Error comparing password and hash: %v", err)
		return false, nil
	}

	return match, nil
}

// A login for an unknown username, or for a user with no password, still runs
// one Argon2id comparison against a dummy hash, through the same concurrency
// limit. Otherwise the response time, and whether the limit is reported as
// busy, would show which usernames exist.
// The dummy password is random, so it is not known to anyone, and the result of
// the comparison is never used.
var (
	dummyPassword         string
	dummyPasswordHash     string
	dummyPasswordHashErr  error
	dummyPasswordHashOnce sync.Once
)

func getDummyPasswordHash() (string, error) {
	dummyPasswordHashOnce.Do(func() {
		dummyPassword = rand.Text()
		dummyPasswordHash, dummyPasswordHashErr = argon2id.CreateHash(dummyPassword, &defaultParams)
	})

	return dummyPasswordHash, dummyPasswordHashErr
}

func compareAgainstDummyHash(password string) error {
	hash, err := getDummyPasswordHash()
	if err != nil {
		return err
	}

	_, err = comparePasswordAndHash(password, hash)

	return err
}

func checkUserPassword(cfg *config.Config, username, password string) (bool, error) {
	user := cfg.FindUserByUsername(username)
	if user == nil || user.Password == "" {
		log.WithFields(log.Fields{"username": username}).Warn("Failed to check password for user, as the username was not found or has no password")
		return false, compareAgainstDummyHash(password)
	}

	return comparePasswordAndLogResult(password, user.Password, username)
}

func comparePasswordAndLogResult(password, hash, username string) (bool, error) {
	match, err := comparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}
	if !match {
		log.WithFields(log.Fields{"username": username}).Warn("Password does not match for user")
		return false, nil
	}
	return true, nil
}
