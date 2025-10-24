package config

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"

	"github.com/knadh/koanf/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"
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

func AppendSource(cfg *Config, k *koanf.Koanf, configPath string) {
	log.Infof("Appending cfg source: %s", configPath)

	// Try default unmarshaling first
	err := k.Unmarshal(".", cfg)
	if err != nil {
		log.Errorf("Error unmarshalling config: %v", err)
		return
	}

	// If actions are not loaded by default unmarshaling, try manual unmarshaling
	// This is a workaround for a koanf issue where []*Action fields are not unmarshaled correctly
	if len(cfg.Actions) == 0 && k.Exists("actions") {
		var actions []*Action
		err := k.Unmarshal("actions", &actions)
		if err != nil {
			log.Errorf("Error manually unmarshaling actions: %v", err)
		} else {
			cfg.Actions = actions
		}
	}

	// If dashboards are not loaded by default unmarshaling, try manual unmarshaling
	// This is a workaround for a koanf issue where []*DashboardComponent fields are not unmarshaled correctly
	if len(cfg.Dashboards) == 0 && k.Exists("dashboards") {
		var dashboards []*DashboardComponent
		err := k.Unmarshal("dashboards", &dashboards)
		if err != nil {
			log.Errorf("Error manually unmarshaling dashboards: %v", err)
		} else {
			cfg.Dashboards = dashboards
		}
	}

	// If entities are not loaded by default unmarshaling, try manual unmarshaling
	// This is a workaround for a koanf issue where []*EntityFile fields are not unmarshaled correctly
	if len(cfg.Entities) == 0 && k.Exists("entities") {
		var entities []*EntityFile
		err := k.Unmarshal("entities", &entities)
		if err != nil {
			log.Errorf("Error manually unmarshaling entities: %v", err)
		} else {
			cfg.Entities = entities
		}
	}

	// Manual field assignment for other config fields that might not be unmarshaled correctly
	if k.Exists("showFooter") {
		cfg.ShowFooter = k.Bool("showFooter")
	}
	if k.Exists("showNavigation") {
		cfg.ShowNavigation = k.Bool("showNavigation")
	}
	if k.Exists("checkForUpdates") {
		cfg.CheckForUpdates = k.Bool("checkForUpdates")
	}
	if k.Exists("pageTitle") {
		cfg.PageTitle = k.String("pageTitle")
	}
	if k.Exists("listenAddressSingleHTTPFrontend") {
		cfg.ListenAddressSingleHTTPFrontend = k.String("listenAddressSingleHTTPFrontend")
	}
	if k.Exists("listenAddressWebUI") {
		cfg.ListenAddressWebUI = k.String("listenAddressWebUI")
	}
	if k.Exists("listenAddressRestActions") {
		cfg.ListenAddressRestActions = k.String("listenAddressRestActions")
	}
	if k.Exists("listenAddressGrpcActions") {
		cfg.ListenAddressGrpcActions = k.String("listenAddressGrpcActions")
	}
	if k.Exists("listenAddressPrometheus") {
		cfg.ListenAddressPrometheus = k.String("listenAddressPrometheus")
	}
	if k.Exists("useSingleHTTPFrontend") {
		cfg.UseSingleHTTPFrontend = k.Bool("useSingleHTTPFrontend")
	}
	if k.Exists("logLevel") {
		cfg.LogLevel = k.String("logLevel")
	}

	// Handle defaultPolicy nested struct
	if k.Exists("defaultPolicy") {
		if k.Exists("defaultPolicy.showDiagnostics") {
			cfg.DefaultPolicy.ShowDiagnostics = k.Bool("defaultPolicy.showDiagnostics")
		}
		if k.Exists("defaultPolicy.showLogList") {
			cfg.DefaultPolicy.ShowLogList = k.Bool("defaultPolicy.showLogList")
		}
	}

	// Handle prometheus nested struct
	if k.Exists("prometheus") {
		if k.Exists("prometheus.enabled") {
			cfg.Prometheus.Enabled = k.Bool("prometheus.enabled")
		}
		if k.Exists("prometheus.defaultGoMetrics") {
			cfg.Prometheus.DefaultGoMetrics = k.Bool("prometheus.defaultGoMetrics")
		}
	}

	metricConfigReloadedCount.Inc()
	metricConfigActionCount.Set(float64(len(cfg.Actions)))

	cfg.SetDir(filepath.Dir(configPath))
	cfg.Sanitize()

	for _, l := range listeners {
		l()
	}
}

var envRegex = regexp.MustCompile(`\${{ *?(\S+) *?}}`)

func envDecodeHookFunc(from reflect.Type, to reflect.Type, data any) (any, error) {
	log.Debugf("envDecodeHookFunc called: from=%v, to=%v, data=%v", from, to, data)
	if from.Kind() != reflect.String {
		return data, nil
	}
	input := data.(string)
	log.Debugf("Processing string input: %q", input)
	output := envRegex.ReplaceAllStringFunc(input, func(match string) string {
		submatches := envRegex.FindStringSubmatch(match)
		key := submatches[1]
		val, set := os.LookupEnv(key)
		log.Debugf("Environment variable %q: set=%v, value=%q", key, set, val)
		if !set {
			log.Warnf("Config file references unset environment variable: \"%s\"", key)
		}
		return val
	})
	log.Debugf("Environment variable interpolation result: %q -> %q", input, output)
	return output, nil
}
