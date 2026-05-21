package auth

import (
	"crypto/subtle"
	"strings"

	types "github.com/OliveTin/OliveTin/internal/auth/authpublic"
	"github.com/OliveTin/OliveTin/internal/config"
	log "github.com/sirupsen/logrus"
)

const localBearerScheme = "Bearer"

func constantTimeEqualString(a, b string) bool {
	if len(a) != len(b) {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

func bearerTokenFromAuthorizationHeader(authz string) (string, bool) {
	idx := strings.IndexByte(authz, ' ')
	if idx <= 0 {
		return "", false
	}

	if !strings.EqualFold(authz[:idx], localBearerScheme) {
		return "", false
	}

	token := strings.TrimSpace(authz[idx+1:])
	if token == "" {
		return "", false
	}

	return token, true
}

func localUserHasAPIKey(user *config.LocalUser) bool {
	return user != nil && user.ApiKey != ""
}

func findLocalUserByAPIKey(cfg *config.Config, token string) *config.LocalUser {
	for _, user := range cfg.AuthLocalUsers.Users {
		if !localUserHasAPIKey(user) {
			continue
		}

		if constantTimeEqualString(token, user.ApiKey) {
			return user
		}
	}

	return nil
}

func localBearerAuthorizationHasEmptyCredential(authz string) bool {
	idx := strings.IndexByte(authz, ' ')
	return idx > 0 &&
		strings.EqualFold(authz[:idx], localBearerScheme) &&
		strings.TrimSpace(authz[idx+1:]) == ""
}

func logLocalBearerAPIKeyParseFailure(authz string) {
	if strings.TrimSpace(authz) == "" {
		return
	}

	if localBearerAuthorizationHasEmptyCredential(authz) {
		log.Debugf("Local bearer API key: rejected (empty credential after Bearer prefix)")
		return
	}

	log.Tracef("Local bearer API key: skipped (Authorization is not a Bearer token)")
}

func checkUserFromLocalBearerApiKey(context *types.AuthCheckingContext) *types.AuthenticatedUser {
	if !context.Config.AuthLocalUsers.Enabled {
		log.Tracef("Local bearer API key: skipped (authLocalUsers disabled)")
		return nil
	}

	authz := context.Request.Header.Get("Authorization")
	token, ok := bearerTokenFromAuthorizationHeader(authz)
	if !ok {
		logLocalBearerAPIKeyParseFailure(authz)
		return nil
	}

	log.Debugf("Local bearer API key: checking configured local user API keys")

	user := findLocalUserByAPIKey(context.Config, token)
	if user == nil {
		log.Debugf("Local bearer API key: rejected (no matching local user)")
		return nil
	}

	log.WithFields(log.Fields{
		"username":  user.Username,
		"usergroup": user.Usergroup,
	}).Debugf("Local bearer API key: authenticated")

	return &types.AuthenticatedUser{
		Username:      user.Username,
		UsergroupLine: user.Usergroup,
		Provider:      "local",
	}
}
