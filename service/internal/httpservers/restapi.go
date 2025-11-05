package httpservers

import (
	"context"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"

	apiv1 "github.com/OliveTin/OliveTin/gen/grpc/olivetin/api/v1"

	"github.com/OliveTin/OliveTin/internal/acl"
	config "github.com/OliveTin/OliveTin/internal/config"
	cors "github.com/OliveTin/OliveTin/internal/cors"
)

var (
	cfg *config.Config
)

func parseHttpHeaderForAuth(req *http.Request) (string, string) {
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
func authHttpRequest(req *http.Request) acl.UnauthenticatedUser {
	ret := acl.UnauthenticatedUser{
		Username:  "",
		Usergroup: "",
		Provider:  "unknown",
		Sid:       "",
	}

	if cfg.AuthJwtHeader != "" {
		ret.Username, ret.Usergroup = parseJwtHeader(req)
		ret.Provider = "jwt-header"
	}

	if cfg.AuthJwtCookieName != "" && ret.Username == "" {
		ret.Username, ret.Usergroup = parseJwtCookie(req)
		ret.Provider = "jwt-cookie"
	}

	if cfg.AuthHttpHeaderUsername != "" && ret.Username == "" {
		ret.Username, ret.Usergroup = parseHttpHeaderForAuth(req)
		ret.Provider = "http-header"
	}

	if len(cfg.AuthOAuth2Providers) > 0 && ret.Username == "" {
		ret.Username, ret.Usergroup, ret.Sid = parseOAuth2Cookie(req)
		ret.Provider = "oauth2"
	}

	if cfg.AuthLocalUsers.Enabled && ret.Username == "" {
		ret.Username, ret.Usergroup, ret.Sid = parseLocalUserCookie(req)
		ret.Provider = "local"
	}

	return ret
}

func authHttpRequestToMetadata(ctx context.Context, req *http.Request) metadata.MD {
	authMetadata := authHttpRequest(req)

	md := metadata.New(map[string]string{
		"username":  authMetadata.Username,
		"usergroup": authMetadata.Usergroup,
		"provider":  authMetadata.Provider,
		"sid":       authMetadata.Sid,
	})

	log.Tracef("api request metadata: %+v", md)

	return md
}

func parseJwtHeader(req *http.Request) (string, string) {
	// JWTs in the Authorization header are usually prefixed with "Bearer " which is not part of the JWT token.
	return parseJwt(strings.TrimPrefix(req.Header.Get(cfg.AuthJwtHeader), "Bearer "))
}

func forwardResponseHandler(ctx context.Context, w http.ResponseWriter, msg protoreflect.ProtoMessage) error {
	md, ok := runtime.ServerMetadataFromContext(ctx)

	if !ok {
		log.Warn("Could not get ServerMetadata from context")
		return nil
	}

	forwardResponseHandlerLoginLocalUser(md.HeaderMD, w)
	forwardResponseHandlerLogout(md.HeaderMD, w)

	return nil
}

func forwardResponseHandlerLogout(md metadata.MD, w http.ResponseWriter) {
	if getMetadataKeyOrEmpty(md, "logout-provider") != "" {
		sid := getMetadataKeyOrEmpty(md, "logout-sid")

		delete(registeredStates, sid)
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

		deleteLocalUserSession("local", sid)

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

func SetGlobalRestConfig(config *config.Config) {
	cfg = config
}

func startRestAPIServer(globalConfig *config.Config) error {
	cfg = globalConfig

	loadUserSessions()

	log.WithFields(log.Fields{
		"address": cfg.ListenAddressRestActions,
	}).Info("Starting REST API")

	mux := newMux()

	return http.ListenAndServe(cfg.ListenAddressRestActions, cors.AllowCors(mux))
}

func newMux() *runtime.ServeMux {
	// The MarshalOptions set some important compatibility settings for the webui. See below.
	mux := runtime.NewServeMux(
		runtime.WithMetadata(authHttpRequestToMetadata),
		runtime.WithForwardResponseOption(forwardResponseHandler),
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.HTTPBodyMarshaler{
			Marshaler: &runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{
					UseProtoNames:   false, // eg: canExec for js instead of can_exec from protobuf
					EmitUnpopulated: true,  // Emit empty fields so that javascript does not get "undefined" when accessing fields with empty values.
				},
			},
		}),
	)

	ctx := context.Background()

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(
			insecure.NewCredentials(),
		),
	}

	err := apiv1.RegisterOliveTinApiServiceHandlerFromEndpoint(ctx, mux, cfg.ListenAddressGrpcActions, opts)

	if err != nil {
		log.Panicf("Could not register REST API Handler %v", err)
	}

	return mux
}
