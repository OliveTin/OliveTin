package httpservers

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
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

func parseRequestMetadata(ctx context.Context, req *http.Request) metadata.MD {
	username := ""
	usergroup := ""

	if cfg.AuthJwtCookieName != "" {
		username, usergroup = parseJwtCookie(req)
	}

	if cfg.AuthHttpHeaderUsername != "" {
		username, usergroup = parseHttpHeaderForAuth(req)
	}

	md := metadata.Pairs(
		"username", username,
		"usergroup", usergroup,
	)

	log.Debugf("api request metadata: %+v", md)

	return md
}

func SetGlobalRestConfig(config *config.Config) {
	cfg = config
}

func startRestAPIServer(globalConfig *config.Config) error {
	cfg = globalConfig

	log.WithFields(log.Fields{
		"address": cfg.ListenAddressGrpcActions,
	}).Info("Starting REST API")

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// The JSONPb.EmitDefaults is necssary, so "empty" fields are returned in JSON.
	mux := runtime.NewServeMux(
		runtime.WithMetadata(parseRequestMetadata),
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.HTTPBodyMarshaler{
			Marshaler: &runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{
					UseProtoNames:   false, // eg: canExec for js instead of can_exec from protobuf
					EmitUnpopulated: true,
				},
			},
		}),
	)
	opts := []grpc.DialOption{grpc.WithInsecure()}

	err := gw.RegisterOliveTinApiServiceHandlerFromEndpoint(ctx, mux, cfg.ListenAddressGrpcActions, opts)

	if err != nil {
		log.Errorf("Could not register REST API Handler %v", err)

		return err
	}

	return http.ListenAndServe(cfg.ListenAddressRestActions, cors.AllowCors(mux))
}
