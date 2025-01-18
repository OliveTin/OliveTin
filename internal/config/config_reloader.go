package config

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/viper"
)

var (
	metricConfigActionCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "olivetin_config_action_count",
		Help: "The number of actions in the config file",
	})

	metricConfigReloadedCount = promauto.NewCounter(prometheus.CounterOpts{
		Name: "olivetin_config_reloaded_count",
		Help: "The number of times the config has been reloaded",
	})

	listeners []func()
)

func AddListener(l func()) {
	listeners = append(listeners, l)
}

func Reload(cfg *Config) {
	if err := viper.UnmarshalExact(&cfg); err != nil {
		log.Errorf("Config unmarshal error %+v", err)
		os.Exit(1)
	}

	if cfg.AuthRequireGuestsToLogin {
		log.Infof("AuthRequireGuestsToLogin is enabled. All defaultPermissions will be set to false")

		cfg.DefaultPermissions.View = false
		cfg.DefaultPermissions.Exec = false
		cfg.DefaultPermissions.Logs = false
	}

	if cfg.LogHistoryPageSize < 10 {
		log.Warnf("LogsHistoryLimit is too low, setting it to 10")
		cfg.LogHistoryPageSize = 10
	} else if cfg.LogHistoryPageSize > 100 {
		log.Warnf("LogsHistoryLimit is high, you can do this, but expect browser lag.")
	}

	metricConfigReloadedCount.Inc()
	metricConfigActionCount.Set(float64(len(cfg.Actions)))

	cfg.SetDir(filepath.Dir(viper.ConfigFileUsed()))
	cfg.Sanitize()

	for _, l := range listeners {
		l()
	}
}
