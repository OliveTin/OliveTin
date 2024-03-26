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
)

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

	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		log.Debugf("api req: %q", r.URL)
		apiProxy.ServeHTTP(w, r)
	})

	mux.HandleFunc("/websocket", func(w http.ResponseWriter, r *http.Request) {
		log.Debugf("ws  req: %q", r.URL)
		websocket.HandleWebsocket(w, r)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Debugf("ui  req: %q", r.URL)
		webuiProxy.ServeHTTP(w, r)
	})

	if cfg.Prometheus.Enabled {
		promURL, _ := url.Parse("http://" + cfg.ListenAddressPrometheus)
		promProxy := httputil.NewSingleHostReverseProxy(promURL)

		mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
			log.Debugf("prom req: %q", r.URL)
			promProxy.ServeHTTP(w, r)
		})
	}

	srv := &http.Server{
		Addr:    cfg.ListenAddressSingleHTTPFrontend,
		Handler: mux,
	}

	log.Fatal(srv.ListenAndServe())
}
