package config

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"

	"github.com/mitchellh/mapstructure"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
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
	if err := viper.UnmarshalExact(&cfg, configureDecoder); err != nil {
		log.Errorf("Config unmarshal error %+v", err)
		os.Exit(1)
	}

	metricConfigReloadedCount.Inc()
	metricConfigActionCount.Set(float64(len(cfg.Actions)))

	cfg.SetDir(filepath.Dir(viper.ConfigFileUsed()))
	cfg.Sanitize()

	for _, l := range listeners {
		l()
	}
}

func configureDecoder(config *mapstructure.DecoderConfig) {
	config.DecodeHook = mapstructure.ComposeDecodeHookFunc(envDecodeHookFunc, config.DecodeHook)

}

var envRegex = regexp.MustCompile(`\${{ *?(\S+) *?}}`)

func envDecodeHookFunc(from reflect.Value, to reflect.Value) (any, error) {
	if from.Kind() != reflect.String {
		return from.Interface(), nil
	}
	input := from.Interface().(string)
	output := envRegex.ReplaceAllStringFunc(input, func(match string) string {
		submatches := envRegex.FindStringSubmatch(match)
		key := submatches[1]
		val, set := os.LookupEnv(key)
		if !set {
			log.Warnf("Config file references unset environment variable: \"%s\"", key)
		}
		return val
	})
	return output, nil
}
