package httpservers

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
)

var (
	pubKeyBytes []byte = nil
	pubKey      *rsa.PublicKey
)

func readPublicKey() error {
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

func parseJwtTokenWithKey(cookieValue string) (*jwt.Token, error) {
	err := readPublicKey()

	if err != nil {
		return nil, err
	}

	return jwt.Parse(cookieValue, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf(
				"expected token algorithm '%v' but got '%v'",
				jwt.SigningMethodRS256.Name,
				token.Header)
		}
		return pubKey, nil
	})
}

func parseJwtTokenWithoutKey(cookieValue string) (*jwt.Token, error) {
	return jwt.Parse(cookieValue, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(cfg.AuthJwtSecret), nil
	})
}

func parseJwtToken(cookieValue string) (*jwt.Token, error) {
	if cfg.AuthJwtPubKeyPath != "" { // activate this path only if pub key is specified
		return parseJwtTokenWithKey(cookieValue)
	} else {
		return parseJwtTokenWithoutKey(cookieValue)
	}
}

func getClaimsFromJwtToken(cookieValue string) (jwt.MapClaims, error) {
	token, err := parseJwtToken(cookieValue)

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

	claims, err := getClaimsFromJwtToken(cookie.Value)

	log.Debugf("jwt claims data: %+v", claims)

	if err != nil {
		log.Warnf("jwt claim error: %+v", err)
		return "", ""
	}

	username := lookupClaimValueOrDefault(claims, cfg.AuthJwtClaimUsername, "")
	usergroup := lookupClaimValueOrDefault(claims, cfg.AuthJwtClaimUserGroup, "")

	return username, usergroup
}
