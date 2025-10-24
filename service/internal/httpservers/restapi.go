package httpservers

import (
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"

	//	apiv1 "github.com/OliveTin/OliveTin/gen/olivetin/api/v1"

	config "github.com/OliveTin/OliveTin/internal/config"
)

func parseHttpHeaderForAuth(cfg *config.Config, req *http.Request) (string, string) {
	username, ok := req.Header[cfg.AuthHttpHeaderUsername]

	if !ok {
		log.Warnf("Config has AuthHttpHeaderUsername set to %v, but it was not found", cfg.AuthHttpHeaderUsername)

		return "", ""
	}

	if cfg.AuthHttpHeaderUserGroup != "" {
		usergroup, ok := req.Header[cfg.AuthHttpHeaderUserGroup]

		if ok {
			log.Debugf("HTTP Header Auth found a username and usergroup")

			return username[0], usergroup[0]
		} else {
			log.Warnf("Config has AuthHttpHeaderUserGroup set to %v, but it was not found", cfg.AuthHttpHeaderUserGroup)
		}
	}

	log.Debugf("HTTP Header Auth found a username, but usergroup is not being used")

	return username[0], ""
}

//gocyclo:ignore
func parseJwtHeader(cfg *config.Config, req *http.Request) (string, string) {
	// JWTs in the Authorization header are usually prefixed with "Bearer " which is not part of the JWT token.
	return parseJwt(cfg, strings.TrimPrefix(req.Header.Get(cfg.AuthJwtHeader), "Bearer "))
}
