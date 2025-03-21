package httpservers

import (
	config "github.com/OliveTin/OliveTin/internal/config"
)

// StartServers will start 3 HTTP servers. The WebUI, the Rest API, and a proxy
// for both of them.
func StartServers(cfg *config.Config) {
	go startWebUIServer(cfg)

	if cfg.UseSingleHTTPFrontend {
		go StartSingleHTTPFrontend(cfg)
	}

	if cfg.Prometheus.Enabled {
		go StartPrometheus(cfg)
	}

	startRestAPIServer(cfg)
}
