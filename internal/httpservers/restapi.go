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

	if cfg.AuthJwtCookieName != "" {
		username, usergroup = parseJwtCookie(req)
	}

	if cfg.AuthHttpHeaderUsername != "" {
		username, usergroup = parseHttpHeaderForAuth(req)
	}

	if len(cfg.AuthOAuth2Providers) > 0 {
		username, usergroup = parseOAuth2Cookie(req)
	}

	if cfg.AuthLocalUsers.Enabled {
		username, usergroup = parseLocalUserCookie(req)
	}

	md := metadata.New(map[string]string{
		"username":  username,
		"usergroup": usergroup,
	})

	log.Tracef("api request metadata: %+v", md)

	return md
}

func forwardResponseHandler(ctx context.Context, w http.ResponseWriter, msg protoreflect.ProtoMessage) error {
	forwardResponseHandlerLoginLocalUser(ctx, w, msg)

	return nil
}

func SetGlobalRestConfig(config *config.Config) {
	cfg = config
}

func startRestAPIServer(globalConfig *config.Config) error {
	cfg = globalConfig

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
