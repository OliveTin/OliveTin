package httpservers

/*
This file implements a very simple, lightweight reverse proxy so that REST and
the webui can be accessed from a single endpoint.

This makes external reverse proxies (treafik, haproxy, etc) easier, CORS goes
away, and several other issues.
*/

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"

	"github.com/OliveTin/OliveTin/internal/api"
	"github.com/OliveTin/OliveTin/internal/auth"
	"github.com/OliveTin/OliveTin/internal/auth/otoauth2"
	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/executor"
	log "github.com/sirupsen/logrus"
)

func logDebugRequest(cfg *config.Config, source string, r *http.Request) {
	if cfg.LogDebugOptions.SingleFrontendRequests {
		log.Debugf("SingleFrontend HTTP Req URL %v: %q", source, r.URL)

		if cfg.LogDebugOptions.SingleFrontendRequestHeaders {
			for name, values := range r.Header {
				log.Debugf("SingleFrontend HTTP Req Hdr: %v = %v", name, values)
			}
		}
	}
}

func StartFrontendMux(cfg *config.Config, ex *executor.Executor) {
	log.WithFields(log.Fields{
		"address": cfg.ListenAddressSingleHTTPFrontend,
	}).Info("Starting single HTTP frontend")

	go StartPrometheus(cfg)

	mux := http.NewServeMux()

	apiPath, apiHandler := api.GetNewHandler(ex)

	log.Infof("API path is %s", apiPath)

	mux.Handle("/api/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fn := path.Base(r.URL.Path)

		// Translate /api/foo/bar to /api/bar - this preserves compatibility
		// with OliveTin 2k.

		r.URL.Path = apiPath + fn

		log.WithFields(log.Fields{
			"path": r.URL.Path,
		}).Tracef("SingleFrontend HTTP API Req URL after rewrite")

		logDebugRequest(cfg, "api", r)

		apiHandler.ServeHTTP(w, r)
	}))

	oauth2handler := otoauth2.NewOAuth2Handler(cfg)
	auth.AddAuthChainFunction(oauth2handler.CheckUserFromOAuth2Cookie)

	mux.HandleFunc("/oauth/login", oauth2handler.HandleOAuthLogin)
	mux.HandleFunc("/oauth/callback", oauth2handler.HandleOAuthCallback)

	mux.HandleFunc("/readyz", handleReadyz)

	webuiServer := NewWebUIServer(cfg)

	mux.HandleFunc("/theme.css", webuiServer.generateThemeCss)
	mux.Handle("/custom-webui/", webuiServer.handleCustomWebui())
	mux.HandleFunc("/", webuiServer.handleWebui)

	if cfg.Prometheus.Enabled {
		promURL, _ := url.Parse("http://" + cfg.ListenAddressPrometheus)
		promProxy := httputil.NewSingleHostReverseProxy(promURL)

		mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
			logDebugRequest(cfg, "prom", r)

			promProxy.ServeHTTP(w, r)
		})
	}

	srv := &http.Server{
		Addr:    cfg.ListenAddressSingleHTTPFrontend,
		Handler: mux,
	}

	log.Fatal(srv.ListenAndServe())
}

func handleReadyz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("OK. Single HTTP Frontend is ready.\n"))

	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Warnf("Failed to write readyz response")
	}
}
