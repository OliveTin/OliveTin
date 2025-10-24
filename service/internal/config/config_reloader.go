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

	// If authLocalUsers are not loaded by default unmarshaling, try manual unmarshaling
	// This is a workaround for a koanf issue where nested struct fields are not unmarshaled correctly
	if len(cfg.AuthLocalUsers.Users) == 0 && k.Exists("authLocalUsers") {
		var authLocalUsers AuthLocalUsersConfig
		err := k.Unmarshal("authLocalUsers", &authLocalUsers)
		if err != nil {
			log.Errorf("Error manually unmarshaling authLocalUsers: %v", err)
		} else {
			cfg.AuthLocalUsers = authLocalUsers
		}
	}

	// Manual field assignment for other config fields that might not be unmarshaled correctly
	boolVal(k, "showFooter", &cfg.ShowFooter)
	boolVal(k, "showNavigation", &cfg.ShowNavigation)
	boolVal(k, "checkForUpdates", &cfg.CheckForUpdates)
	stringVal(k, "pageTitle", &cfg.PageTitle)
	stringVal(k, "listenAddressSingleHTTPFrontend", &cfg.ListenAddressSingleHTTPFrontend)
	stringVal(k, "listenAddressWebUI", &cfg.ListenAddressWebUI)
	stringVal(k, "listenAddressRestActions", &cfg.ListenAddressRestActions)
	stringVal(k, "listenAddressPrometheus", &cfg.ListenAddressPrometheus)
	boolVal(k, "useSingleHTTPFrontend", &cfg.UseSingleHTTPFrontend)
	stringVal(k, "logLevel", &cfg.LogLevel)

	// Handle defaultPolicy nested struct
	if k.Exists("defaultPolicy") {
		boolVal(k, "defaultPolicy.showDiagnostics", &cfg.DefaultPolicy.ShowDiagnostics)
		boolVal(k, "defaultPolicy.showLogList", &cfg.DefaultPolicy.ShowLogList)
	}

	// Handle prometheus nested struct
	if k.Exists("prometheus") {
		boolVal(k, "prometheus.enabled", &cfg.Prometheus.Enabled)
		boolVal(k, "prometheus.defaultGoMetrics", &cfg.Prometheus.DefaultGoMetrics)
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

// Helper functions to reduce repetitive if/set chains
func stringVal(k *koanf.Koanf, key string, dest *string) {
	if k.Exists(key) {
		*dest = k.String(key)
	}
}

func boolVal(k *koanf.Koanf, key string, dest *bool) {
	if k.Exists(key) {
		*dest = k.Bool(key)
	}
}

func int64Val(k *koanf.Koanf, key string, dest *int64) {
	if k.Exists(key) {
		*dest = k.Int64(key)
	}
}

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
