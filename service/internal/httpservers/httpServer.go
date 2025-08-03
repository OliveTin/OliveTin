package httpservers

import (
	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/executor"
)

// StartServers will start 3 HTTP servers. The WebUI, the Rest API, and a proxy
// for both of them.
func StartServers(cfg *config.Config, ex *executor.Executor) {
	if cfg.Prometheus.Enabled {
		go StartPrometheus(cfg)
	}

	StartSingleHTTPFrontend(cfg, ex)
}
