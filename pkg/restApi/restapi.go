package restApi;

import (
	"google.golang.org/grpc"
	log "github.com/sirupsen/logrus"
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"net/http"

	gw "github.com/jamesread/OliveTin/gen/grpc"

	cors "github.com/jamesread/OliveTin/pkg/cors"

	config "github.com/jamesread/OliveTin/pkg/config"
)

var (
	cfg *config.Config;
)


func Start(listenAddressRest string, listenAddressGrpc string, globalConfig *config.Config) (error) {
	cfg = globalConfig

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// The JSONPb.EmitDefaults is necssary, so "empty" fields are returned in JSON.
	mux := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{OrigName: true, EmitDefaults: true}))
	opts := []grpc.DialOption{grpc.WithInsecure()}

	err := gw.RegisterOliveTinApiHandlerFromEndpoint(ctx, mux, listenAddressGrpc, opts)

	if err != nil {
		log.Fatalf("gw error %v", err)
	}

	return http.ListenAndServe(listenAddressRest, cors.AllowCors(mux))
}
