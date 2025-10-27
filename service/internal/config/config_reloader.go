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

	// Unmarshal the entire config with mapstructure tags
	err := k.Unmarshal(".", cfg)
	if err != nil {
		log.Errorf("Error unmarshalling config: %v", err)
		return
	}

	// Fallback for complex nested structures that might not unmarshal correctly
	// Only attempt manual unmarshaling if the automatic approach didn't populate the fields
	if len(cfg.Actions) == 0 && k.Exists("actions") {
		var actions []*Action
		if err := k.Unmarshal("actions", &actions); err == nil {
			cfg.Actions = actions
			log.Debugf("Manually loaded %d actions", len(actions))
		}
	}

	if len(cfg.Dashboards) == 0 && k.Exists("dashboards") {
		var dashboards []*DashboardComponent
		if err := k.Unmarshal("dashboards", &dashboards); err == nil {
			cfg.Dashboards = dashboards
			log.Debugf("Manually loaded %d dashboards", len(dashboards))
		}
	}

	if len(cfg.Entities) == 0 && k.Exists("entities") {
		var entities []*EntityFile
		if err := k.Unmarshal("entities", &entities); err == nil {
			cfg.Entities = entities
			log.Debugf("Manually loaded %d entities", len(entities))
		}
	}

	if len(cfg.AuthLocalUsers.Users) == 0 && k.Exists("authLocalUsers") {
		var authLocalUsers AuthLocalUsersConfig
		if err := k.Unmarshal("authLocalUsers", &authLocalUsers); err == nil {
			cfg.AuthLocalUsers = authLocalUsers
			log.Debugf("Manually loaded local auth config")
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
