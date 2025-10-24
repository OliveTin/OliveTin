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

// StartSingleHTTPFrontend will create a reverse proxy that proxies the API
// and webui internally.
func StartSingleHTTPFrontend(cfg *config.Config, ex *executor.Executor) {
	log.WithFields(log.Fields{
		"address": cfg.ListenAddressSingleHTTPFrontend,
	}).Info("Starting single HTTP frontend")

	mux := http.NewServeMux()

	apiPath, apiHandler := api.GetNewHandler(ex)

	log.Infof("API path is %s", apiPath)

	mux.Handle("/api/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fn := path.Base(r.URL.Path)

		// Translate /api/foo/bar to /api/bar - this preserves compatibility
		// with OliveTin 2k.

		r.URL.Path = apiPath + fn

		log.Debugf("SingleFrontend HTTP API Req URL after rewrite: %v", r.URL.Path)

		apiHandler.ServeHTTP(w, r)
	}))

	oauth2handler := NewOAuth2Handler(cfg)

	mux.HandleFunc("/oauth/login", oauth2handler.handleOAuthLogin)
	mux.HandleFunc("/oauth/callback", oauth2handler.handleOAuthCallback)

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
	w.Write([]byte("OK. Single HTTP Frontend is ready.\n"))
}
