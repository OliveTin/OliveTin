package otjwt

import (
	"bytes"
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/MicahParks/keyfunc/v3"
	authTypes "github.com/OliveTin/OliveTin/internal/auth/authpublic"
	"github.com/OliveTin/OliveTin/internal/config"
	"github.com/golang-jwt/jwt/v5"
	log "github.com/sirupsen/logrus"
)

func parseJwtToken(cfg *config.Config, jwtString string) (*jwt.Token, error) {
	if cfg.AuthJwtCertsURL != "" {
		return parseJwtTokenWithRemoteKey(cfg, jwtString)
	}

	if cfg.AuthJwtPubKeyPath != "" {
		return parseJwtTokenWithLocalKey(cfg, jwtString)
	}

	if cfg.AuthJwtHmacSecret == "" {
		return nil, errors.New("no JWT authentication method configured")
	}

	return parseJwtTokenWithHMAC(cfg, jwtString)
}

func parserOptionsWithAudience(cfg *config.Config) []jwt.ParserOption {
	if cfg.AuthJwtAud == "" {
		return nil
	}
	return []jwt.ParserOption{jwt.WithAudience(cfg.AuthJwtAud)}
}

func getClaimsFromJwtToken(cfg *config.Config, jwtString string) (jwt.MapClaims, error) {
	token, err := parseJwtToken(cfg, jwtString)

	if err != nil {
		log.Errorf("jwt parse failure: %v", err)
		return nil, errors.New("jwt parse failure")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, errors.New("jwt token isn't valid")
	}
}

func parseJwtTokenWithRemoteKey(cfg *config.Config, jwtToken string) (*jwt.Token, error) {
	err := initJwks(cfg)

	if err != nil {
		log.Errorf("jwt init JWKS failure: %v", err)
		return nil, err
	}

	opts := parserOptionsWithAudience(cfg)
	return jwt.Parse(jwtToken, jwksVerifier.Keyfunc, opts...)
}

var (
	jwksVerifier keyfunc.Keyfunc
	jwksOnce     sync.Once
	jwksInitErr  error

	localKey localKeyCache
)

// Operators rotate the local public key by replacing the file at the same
// path, for example when a Kubernetes or Docker secret is updated, so the file
// is read on every request and parsed again whenever its contents change. A
// key file that cannot be read or parsed rejects every token, rather than
// leaving the previous key trusted.
type localKeyCache struct {
	key      *rsa.PublicKey
	keyBytes []byte
	mu       sync.Mutex
}

func (c *localKeyCache) get(keyPath string) (*rsa.PublicKey, error) {
	keyBytes, err := os.ReadFile(keyPath)

	c.mu.Lock()
	defer c.mu.Unlock()

	if err != nil {
		c.clearLocked()
		return nil, fmt.Errorf("couldn't read public key from file %s", keyPath)
	}

	if c.key != nil && bytes.Equal(c.keyBytes, keyBytes) {
		return c.key, nil
	}

	parsedKey, err := jwt.ParseRSAPublicKeyFromPEM(keyBytes)
	if err != nil {
		c.clearLocked()
		return nil, fmt.Errorf("error parsing public key object (from %s)", keyPath)
	}

	c.keyBytes = keyBytes
	c.key = parsedKey

	return parsedKey, nil
}

func (c *localKeyCache) clearLocked() {
	c.keyBytes = nil
	c.key = nil
}

func initJwks(cfg *config.Config) error {
	jwksOnce.Do(func() {
		if cfg.AuthJwtCertsURL != "" {
			var err error
			jwksVerifier, err = initRemoteJwks(cfg.AuthJwtCertsURL)

			if err != nil {
				log.Errorf("Init JWKS Failure: %v", err)
				jwksInitErr = err
			}
		}
	})
	return jwksInitErr
}

// initRemoteJwks creates a JWKS verifier whose background refresh outlives the
// calling request. keyfunc.NewDefaultCtx stops refreshing when its context is
// cancelled; a WithTimeout/defer-cancel around the initial fetch permanently
// freezes the keyset (GHSA-gp3q-q8qg-j4xf).
func initRemoteJwks(certsURL string) (keyfunc.Keyfunc, error) {
	return keyfunc.NewDefaultCtx(context.Background(), []string{certsURL})
}

func readLocalPublicKey(cfg *config.Config) (*rsa.PublicKey, error) {
	if cfg.AuthJwtPubKeyPath == "" {
		return nil, errors.New("no JWT public key path configured")
	}

	return localKey.get(cfg.AuthJwtPubKeyPath)
}

func parseJwtTokenWithLocalKey(cfg *config.Config, jwtString string) (*jwt.Token, error) {
	publicKey, err := readLocalPublicKey(cfg)
	if err != nil {
		return nil, err
	}

	keyFunc := func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("parseJwt expected token algorithm RSA but got: %v", token.Header["alg"])
		}

		return publicKey, nil
	}

	opts := parserOptionsWithAudience(cfg)
	return jwt.Parse(jwtString, keyFunc, opts...)
}

// Hash-based Message Authentication Code
func parseJwtTokenWithHMAC(cfg *config.Config, jwtString string) (*jwt.Token, error) {
	keyFunc := func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("parseJwt expected token algorithm HMAC but got: %v", token.Header["alg"])
		}

		return []byte(cfg.AuthJwtHmacSecret), nil
	}

	opts := parserOptionsWithAudience(cfg)
	return jwt.Parse(jwtString, keyFunc, opts...)
}

func lookupClaimValueOrDefault(claims jwt.MapClaims, key string, def string) string {
	if val, ok := claims[key]; ok {
		return fmt.Sprintf("%s", val)
	} else {
		return def
	}
}

func CheckUserFromJwtCookie(context *authTypes.AuthCheckingContext) *authTypes.AuthenticatedUser {
	cookie, err := context.Request.Cookie(context.Config.AuthJwtCookieName)

	if err != nil {
		log.Debugf("jwt cookie check %v name: %v", err, context.Config.AuthJwtCookieName)
		return nil
	}

	return parseJwt(context.Config, cookie.Value)
}

func CheckUserFromJwtHeader(context *authTypes.AuthCheckingContext) *authTypes.AuthenticatedUser {
	header := context.Request.Header.Get(context.Config.AuthJwtHeader)
	if header == "" {
		return nil
	}

	token := strings.TrimPrefix(header, "Bearer ")
	token = strings.TrimSpace(token)

	return parseJwt(context.Config, token)
}

func parseJwt(cfg *config.Config, token string) *authTypes.AuthenticatedUser {
	claims, err := getClaimsFromJwtToken(cfg, token)

	if err != nil {
		log.Warnf("jwt claim error: %+v", err)
		return nil
	}

	if cfg.InsecureAllowDumpJwtClaims {
		log.Debugf("JWT Claims %+v", claims)
	}

	user := &authTypes.AuthenticatedUser{
		Username:      lookupClaimValueOrDefault(claims, cfg.AuthJwtClaimUsername, ""),
		UsergroupLine: parseGroupClaim(cfg.AuthJwtClaimUserGroup, claims, cfg.AuthHttpHeaderUserGroupSep),
		Provider:      "jwt",
	}

	return user
}

func parseGroupClaim(groupClaim string, claims jwt.MapClaims, sep string) string {
	val, ok := claims[groupClaim]
	if !ok {
		return ""
	}

	array, isArray := val.([]any)
	if isArray {
		return joinJWTGroupArray(array, groupClaim, sep)
	}

	return fmt.Sprintf("%s", val)
}

func joinJWTGroupArray(arrayVal []any, groupClaim string, sep string) string {
	if sep == "" {
		sep = " "
	}

	groups := make([]string, 0, len(arrayVal))
	for _, element := range arrayVal {
		groupName, isString := element.(string)
		if !isString {
			log.Warnf("Skipping non-string group entry in JWT claim %v: %v", groupClaim, element)
			continue
		}
		groups = append(groups, groupName)
	}

	return strings.Join(groups, sep)
}
