package httpservers

import (
	"net/http"
	"time"

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

	mux := http.NewServeMux()
	mux.Handle("/", promhttp.Handler())

	srv := &http.Server{
		Addr:              cfg.ListenAddressPrometheus,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	err := srv.ListenAndServe()

	if err != nil {
		log.WithFields(log.Fields{
			"address": cfg.ListenAddressPrometheus,
			"error":   err,
		}).Warnf("Failed to start Prometheus server")
	}
}
