package httpservers

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"net/http"

	gw "github.com/OliveTin/OliveTin/gen/grpc"

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
func parseRequestMetadata(ctx context.Context, req *http.Request) metadata.MD {
	username := ""
	usergroup := ""
	provider := "unknown"
	sid := ""

	if cfg.AuthJwtCookieName != "" {
		username, usergroup = parseJwtCookie(req)
		provider = "jwt-cookie"
	}

	if cfg.AuthHttpHeaderUsername != "" && username == "" {
		username, usergroup = parseHttpHeaderForAuth(req)
		provider = "http-header"
	}

	if len(cfg.AuthOAuth2Providers) > 0 && username == "" {
		username, usergroup, sid = parseOAuth2Cookie(req)
		provider = "oauth2"
	}

	if cfg.AuthLocalUsers.Enabled && username == "" {
		username, usergroup, sid = parseLocalUserCookie(req)
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
		runtime.WithMetadata(parseRequestMetadata),
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

	opts := []grpc.DialOption{grpc.WithInsecure()}

	err := gw.RegisterOliveTinApiServiceHandlerFromEndpoint(ctx, mux, cfg.ListenAddressGrpcActions, opts)

	if err != nil {
		log.Panicf("Could not register REST API Handler %v", err)
	}

	return mux
}
