package otjwt

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

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

	return jwt.Parse(jwtToken, jwksVerifier.Keyfunc, jwt.WithAudience(cfg.AuthJwtAud))
}

var (
	pubKeyBytes   []byte = nil
	pubKey        *rsa.PublicKey
	loadedKeyPath string

	jwksVerifier keyfunc.Keyfunc
	jwksOnce     sync.Once
	jwksInitErr  error

	localKeyMutex   sync.RWMutex
	localKeyInitErr error
)

func initJwks(cfg *config.Config) error {
	jwksOnce.Do(func() {
		if cfg.AuthJwtCertsURL != "" {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			var err error
			jwksVerifier, err = keyfunc.NewDefaultCtx(ctx, []string{
				cfg.AuthJwtCertsURL,
			})

			if err != nil {
				log.Errorf("Init JWKS Failure: %v", err)
				jwksInitErr = err
			}
		}
	})
	return jwksInitErr
}

func loadPublicKeyFromFile(keyPath string) error {
	keyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("couldn't read public key from file %s", keyPath)
	}

	parsedKey, err := jwt.ParseRSAPublicKeyFromPEM(keyBytes)
	if err != nil {
		return fmt.Errorf("error parsing public key object (from %s)", keyPath)
	}

	pubKeyBytes = keyBytes
	pubKey = parsedKey
	loadedKeyPath = keyPath
	localKeyInitErr = nil
	return nil
}

func isKeyLoadedForPath(keyPath string) bool {
	return pubKeyBytes != nil && loadedKeyPath == keyPath
}

func readLocalPublicKeyWithLock(keyPath string) error {
	localKeyMutex.RLock()
	alreadyLoaded := isKeyLoadedForPath(keyPath)
	localKeyMutex.RUnlock()

	if alreadyLoaded {
		return nil
	}

	localKeyMutex.Lock()
	defer localKeyMutex.Unlock()

	if isKeyLoadedForPath(keyPath) {
		return nil
	}

	localKeyInitErr = loadPublicKeyFromFile(keyPath)
	return localKeyInitErr
}

func readLocalPublicKey(cfg *config.Config) error {
	if cfg.AuthJwtPubKeyPath == "" {
		return errors.New("no JWT public key path configured")
	}
	return readLocalPublicKeyWithLock(cfg.AuthJwtPubKeyPath)
}

func parseJwtTokenWithLocalKey(cfg *config.Config, jwtString string) (*jwt.Token, error) {
	err := readLocalPublicKey(cfg)

	if err != nil {
		return nil, err
	}

	return jwt.Parse(jwtString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("parseJwt expected token algorithm RSA but got: %v", token.Header["alg"])
		}

		return pubKey, nil
	})
}

// Hash-based Message Authentication Code
func parseJwtTokenWithHMAC(cfg *config.Config, jwtString string) (*jwt.Token, error) {
	return jwt.Parse(jwtString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("parseJwt expected token algorithm HMAC but got: %v", token.Header["alg"])
		}

		return []byte(cfg.AuthJwtHmacSecret), nil
	})
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
		UsergroupLine: parseGroupClaim(cfg.AuthJwtClaimUserGroup, claims),
	}

	return user
}

func parseGroupClaim(groupClaim string, claims jwt.MapClaims) string {
	usergroup := ""
	if val, ok := claims[groupClaim]; ok {
		if array, ok := val.([]interface{}); ok {
			groups := make([]string, len(array))
			for i, v := range array {
				groups[i] = fmt.Sprintf("%s", v)
			}
			usergroup = strings.Join(groups, " ")
		} else {
			usergroup = fmt.Sprintf("%s", val)
		}
	}
	return usergroup
}
