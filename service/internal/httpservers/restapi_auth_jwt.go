package httpservers

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strings"

	//	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/MicahParks/keyfunc/v3"
	"time"
)

var (
	pubKeyBytes []byte = nil
	pubKey      *rsa.PublicKey

	jwksVerifier keyfunc.Keyfunc
)

func initJwks() {
	if jwksVerifier == nil {
		var err error

		if cfg.AuthJwtCertsURL != "" {
			ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)

			jwksVerifier, err = keyfunc.NewDefaultCtx(ctx, []string{
				cfg.AuthJwtCertsURL,
			})

			if err != nil {
				log.Errorf("Init JWKS Failure: %v", err)
			}

			defer cancel()
		}
	}
}

func readLocalPublicKey() error {
	if pubKeyBytes != nil {
		return nil // Already read.
	}

	pubKeyBytes, err := os.ReadFile(cfg.AuthJwtPubKeyPath)
	if err != nil {
		return fmt.Errorf("couldn't read public key from file %s", cfg.AuthJwtPubKeyPath)
	}

	// Since the token is RSA (which we validated at the start of this function), the return type of this function actually has to be rsa.PublicKey!
	pubKey, err = jwt.ParseRSAPublicKeyFromPEM(pubKeyBytes)
	if err != nil {
		return fmt.Errorf("error parsing public key object (from %s)", cfg.AuthJwtPubKeyPath)
	}

	return nil
}

func parseJwtTokenWithRemoteKey(jwtToken string) (*jwt.Token, error) {
	initJwks()

	return jwt.Parse(jwtToken, jwksVerifier.Keyfunc, jwt.WithAudience(cfg.AuthJwtAud))
}

func parseJwtTokenWithLocalKey(jwtString string) (*jwt.Token, error) {
	err := readLocalPublicKey()

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
func parseJwtTokenWithHMAC(jwtString string) (*jwt.Token, error) {
	return jwt.Parse(jwtString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("parseJwt expected token algorithm HMAC but got: %v", token.Header["alg"])
		}

		return []byte(cfg.AuthJwtHmacSecret), nil
	})
}

func parseJwtToken(jwtString string) (*jwt.Token, error) {
	if cfg.AuthJwtCertsURL != "" {
		return parseJwtTokenWithRemoteKey(jwtString)
	}

	if cfg.AuthJwtPubKeyPath != "" {
		return parseJwtTokenWithLocalKey(jwtString)
	}

	return parseJwtTokenWithHMAC(jwtString)
}

func getClaimsFromJwtToken(jwtString string) (jwt.MapClaims, error) {
	token, err := parseJwtToken(jwtString)

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

func lookupClaimValueOrDefault(claims jwt.MapClaims, key string, def string) string {
	if val, ok := claims[key]; ok {
		return fmt.Sprintf("%s", val)
	} else {
		return def
	}
}

func parseJwtCookie(request *http.Request) (string, string) {
	cookie, err := request.Cookie(cfg.AuthJwtCookieName)

	if err != nil {
		log.Debugf("jwt cookie check %v name: %v", err, cfg.AuthJwtCookieName)
		return "", ""
	}

	return parseJwt(cookie.Value)
}

func parseJwt(token string) (string, string) {
	claims, err := getClaimsFromJwtToken(token)

	if err != nil {
		log.Warnf("jwt claim error: %+v", err)
		return "", ""
	}

	if cfg.InsecureAllowDumpJwtClaims {
		log.Debugf("JWT Claims %+v", claims)
	}

	username := lookupClaimValueOrDefault(claims, cfg.AuthJwtClaimUsername, "")
	usergroup := parseGroupClaim(cfg.AuthJwtClaimUserGroup, claims)

	return username, usergroup
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
