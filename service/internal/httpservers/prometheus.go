package httpservers

import (
	"net/http"

	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

func StartPrometheus(cfg *config.Config) {
	if !cfg.Prometheus.Enabled {
		return
	}

	if !cfg.Prometheus.DefaultGoMetrics {
		prometheus.Unregister(collectors.NewGoCollector())
	}

	http.Handle("/", promhttp.Handler())
	err := http.ListenAndServe(cfg.ListenAddressPrometheus, nil)

	if err != nil {
		log.WithFields(log.Fields{
			"address": cfg.ListenAddressPrometheus,
			"error":   err,
		}).Warnf("Failed to start Prometheus server")
	}
}
