package httpservers

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/reflect/protoreflect"
	"net/http"
	"strings"

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
func parseRequestMetadata(cfg *config.Config, ctx context.Context, req *http.Request) metadata.MD {
	username := ""
	usergroup := ""
	provider := "unknown"
	sid := ""

	if cfg.AuthJwtHeader != "" {
		username, usergroup = parseJwtHeader(cfg, req)
		provider = "jwt-header"
	}

	if cfg.AuthJwtCookieName != "" {
		username, usergroup = parseJwtCookie(cfg, req)
		provider = "jwt-cookie"
	}

	if cfg.AuthHttpHeaderUsername != "" && username == "" {
		username, usergroup = parseHttpHeaderForAuth(cfg, req)
		provider = "http-header"
	}

//	if len(cfg.AuthOAuth2Providers) > 0 && username == "" {
//		username, usergroup, sid = parseOAuth2Cookie(req)
//		provider = "oauth2"
//	}

	if cfg.AuthLocalUsers.Enabled && username == "" {
		username, usergroup, sid = parseLocalUserCookie(cfg, req)
		provider = "local"
	}

	md := metadata.New(map[string]string{
		"username":  username,
		"usergroup": usergroup,
		"provider":  provider,
		"sid":       sid,
	})

	log.Tracef("api request metadata: %+v", md)

	return md
}

func parseJwtHeader(cfg *config.Config, req *http.Request) (string, string) {
	// JWTs in the Authorization header are usually prefixed with "Bearer " which is not part of the JWT token.
	return parseJwt(cfg, strings.TrimPrefix(req.Header.Get(cfg.AuthJwtHeader), "Bearer "))
}

func (h *OAuth2Handler) forwardResponseHandler(cfg *config.Config, ctx context.Context, w http.ResponseWriter, msg protoreflect.ProtoMessage) error {
	md, ok := runtime.ServerMetadataFromContext(ctx)

	if !ok {
		log.Warn("Could not get ServerMetadata from context")
		return nil
	}

	forwardResponseHandlerLoginLocalUser(cfg, md.HeaderMD, w)
	h.forwardResponseHandlerLogout(cfg, md.HeaderMD, w)

	return nil
}

func (h *OAuth2Handler) forwardResponseHandlerLogout(cfg *config.Config, md metadata.MD, w http.ResponseWriter) {
	if getMetadataKeyOrEmpty(md, "logout-provider") != "" {
		sid := getMetadataKeyOrEmpty(md, "logout-sid")

		delete(h.registeredStates, sid)
		http.SetCookie(
			w,
			&http.Cookie{
				Name:     "olivetin-sid-oauth",
				MaxAge:   31556952, // 1 year
				Value:    "",
				HttpOnly: true,
				Path:     "/",
			},
		)

		deleteLocalUserSession(cfg, "local", sid)

		http.SetCookie(
			w,
			&http.Cookie{
				Name:     "olivetin-sid-local",
				MaxAge:   31556952, // 1 year
				Value:    "",
				HttpOnly: true,
				Path:     "/",
			},
		)

		w.Header().Set("Content-Type", "text/html")
		// We cannot send a HTTP redirect here, because we don't have access to req.
		w.Write([]byte("<script>window.location.href = '/';</script>"))
	}
}

func getMetadataKeyOrEmpty(md metadata.MD, key string) string {
	mdValues := md.Get(key)

	if len(mdValues) > 0 {
		return mdValues[0]
	}

	return ""
}

