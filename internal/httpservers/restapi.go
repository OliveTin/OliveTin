package httpservers

import (
	"context"
	"errors"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"net/http"

	"github.com/golang-jwt/jwt/v4"

	gw "github.com/OliveTin/OliveTin/gen/grpc"

	config "github.com/OliveTin/OliveTin/internal/config"
	cors "github.com/OliveTin/OliveTin/internal/cors"
)

var (
	cfg *config.Config
)

func parseToken(cookieValue string) (*jwt.Token, error) {
	return jwt.Parse(cookieValue, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(cfg.AuthJwtSecret), nil
	})
}

func getClaimsFromJwtToken(cookieValue string) (jwt.MapClaims, error) {
	token, err := parseToken(cookieValue)

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
		runtime.WithMetadata(func(ctx context.Context, request *http.Request) metadata.MD {
			cookie, err := request.Cookie(cfg.AuthJwtCookieName)

			if err != nil {
				log.Debugf("jwt cookie check %v name: %v", err, cfg.AuthJwtCookieName)
				return nil
			}

			claims, err := getClaimsFromJwtToken(cookie.Value)

			log.Debugf("jwt claims data: %+v", claims)

			if err != nil {
				log.Warnf("jwt claim error: %+v", err)
				return nil
			}

			username := lookupClaimValueOrDefault(claims, "name", "none")
			usergroup := lookupClaimValueOrDefault(claims, "group", "none")

			md := metadata.Pairs(
				"username", username,
				"usergroup", usergroup,
			)

			log.Debugf("jwt usable claims: %+v", md)

			return md
		}),
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.HTTPBodyMarshaler{
			Marshaler: &runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{
					UseProtoNames:   true,
					EmitUnpopulated: true,
				},
			},
		}),
	)
	opts := []grpc.DialOption{grpc.WithInsecure()}

	err := gw.RegisterOliveTinApiHandlerFromEndpoint(ctx, mux, cfg.ListenAddressGrpcActions, opts)

	if err != nil {
		log.Errorf("Could not register REST API Handler %v", err)

		return err
	}

	return http.ListenAndServe(cfg.ListenAddressRestActions, cors.AllowCors(mux))
}
