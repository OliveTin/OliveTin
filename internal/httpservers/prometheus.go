package httpservers

import (
	"net/http"

	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func StartPrometheus(cfg *config.Config) {
	if !cfg.Prometheus.DefaultGoMetrics {
		prometheus.Unregister(collectors.NewGoCollector())
	}

	http.Handle("/", promhttp.Handler())
	http.ListenAndServe(cfg.ListenAddressPrometheus, nil)
}
