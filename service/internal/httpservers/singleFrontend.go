package httpservers

/*
This file implements a very simple, lightweight reverse proxy so that REST and
the webui can be accessed from a single endpoint.

This makes external reverse proxies (treafik, haproxy, etc) easier, CORS goes
away, and several other issues.
*/

import (
	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/websocket"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
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
func StartSingleHTTPFrontend(cfg *config.Config) {
	log.WithFields(log.Fields{
		"address": cfg.ListenAddressSingleHTTPFrontend,
	}).Info("Starting single HTTP frontend")

	apiURL, _ := url.Parse("http://" + cfg.ListenAddressRestActions)
	apiProxy := httputil.NewSingleHostReverseProxy(apiURL)

	webuiURL, _ := url.Parse("http://" + cfg.ListenAddressWebUI)
	webuiProxy := httputil.NewSingleHostReverseProxy(webuiURL)

	mux := http.NewServeMux()
	mux.HandleFunc(baseURLPath("/api/"), func(w http.ResponseWriter, r *http.Request) {
		logDebugRequest(cfg, "api ", r)

		r.URL.Path = strings.TrimPrefix(r.URL.Path, baseURLPath("/"))
		apiProxy.ServeHTTP(w, r)
	})

	mux.HandleFunc(baseURLPath("/websocket"), func(w http.ResponseWriter, r *http.Request) {
		logDebugRequest(cfg, "ws  ", r)

		websocket.HandleWebsocket(w, r)
	})

	mux.HandleFunc(baseURLPath("/oauth/login"), handleOAuthLogin)

	mux.HandleFunc(baseURLPath("/oauth/callback"), handleOAuthCallback)

	mux.HandleFunc(baseURLPath("/"), func(w http.ResponseWriter, r *http.Request) {
		logDebugRequest(cfg, "ui  ", r)

		webuiProxy.ServeHTTP(w, r)
	})

	if cfg.Prometheus.Enabled {
		promURL, _ := url.Parse("http://" + cfg.ListenAddressPrometheus)
		promProxy := httputil.NewSingleHostReverseProxy(promURL)

		mux.HandleFunc(baseURLPath("/metrics"), func(w http.ResponseWriter, r *http.Request) {
			logDebugRequest(cfg, "prom", r)

			promProxy.ServeHTTP(w, r)
		})
	}

	oauth2Init(cfg)

	srv := &http.Server{
		Addr:    cfg.ListenAddressSingleHTTPFrontend,
		Handler: mux,
	}

	log.Fatal(srv.ListenAndServe())
}
