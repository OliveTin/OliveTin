package restApi;

import (
	"google.golang.org/grpc"
	log "github.com/sirupsen/logrus"
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"net/http"
	"strings"

	gw "github.com/jamesread/OliveTin/gen/grpc"

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

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}

	err := gw.RegisterOliveTinApiHandlerFromEndpoint(ctx, mux, listenAddressGrpc, opts)

	if err != nil {
		log.Fatalf("gw error %v", err)
	}

	return http.ListenAndServe(listenAddressRest, allowCors(mux))
}

func allowCors(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
				preflightHandler(w, r)
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}

func preflightHandler(w http.ResponseWriter, r *http.Request) {
	headers := []string{"Content-Type", "Accept"}
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))
	methods := []string{"GET", "HEAD", "POST", "PUT", "DELETE"}
	w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
	log.Infof("preflight request for %s", r.URL.Path)
	return
}
