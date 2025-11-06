package config

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
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

	if !unmarshalRoot(k, cfg) {
		return
	}

	afterLoadFinalize(cfg, configPath)
}

func AppendSourceWithIncludes(cfg *Config, k *koanf.Koanf, configPath string) {
	AppendSource(cfg, k, configPath)

	if cfg.Include != "" {
		LoadIncludedConfigs(cfg, k, configPath)
	}
}

func unmarshalRoot(k *koanf.Koanf, cfg *Config) bool {
	err := k.UnmarshalWithConf("", cfg, koanf.UnmarshalConf{
		Tag: "koanf",
	})

	if err != nil {
		log.Errorf("Error unmarshalling config: %v", err)
		return false
	}
	return true
}

func afterLoadFinalize(cfg *Config, configPath string) {
	metricConfigReloadedCount.Inc()
	metricConfigActionCount.Set(float64(len(cfg.Actions)))

	cfg.SetDir(filepath.Dir(configPath))
	cfg.Sanitize()

	for _, l := range listeners {
		l()
	}
}

// LoadIncludedConfigs loads configuration files from an include directory and merges them
func LoadIncludedConfigs(cfg *Config, k *koanf.Koanf, baseConfigPath string) {
	if cfg.Include == "" {
		return
	}

	includePath := filepath.Join(filepath.Dir(baseConfigPath), cfg.Include)
	log.Infof("Loading included configs from: %s", includePath)

	yamlFiles, ok := listYamlFiles(includePath)
	if !ok || len(yamlFiles) == 0 {
		return
	}

	sort.Strings(yamlFiles)
	for _, filename := range yamlFiles {
		loadAndMergeIncludedFile(k, cfg, includePath, filename)
	}

	log.Infof("Finished loading %d included config file(s)", len(yamlFiles))
	cfg.Sanitize()
}

func listYamlFiles(includePath string) ([]string, bool) {
	dirInfo, err := os.Stat(includePath)
	if err != nil {
		log.Warnf("Include directory not found: %s", includePath)
		return nil, false
	}
	if !dirInfo.IsDir() {
		log.Warnf("Include path is not a directory: %s", includePath)
		return nil, false
	}
	entries, err := os.ReadDir(includePath)
	if err != nil {
		log.Errorf("Error reading include directory: %v", err)
		return nil, false
	}
	var yamlFiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".yml") || strings.HasSuffix(name, ".yaml") {
			yamlFiles = append(yamlFiles, name)
		}
	}
	if len(yamlFiles) == 0 {
		log.Infof("No YAML files found in include directory: %s", includePath)
	}
	return yamlFiles, true
}

func loadAndMergeIncludedFile(k *koanf.Koanf, cfg *Config, includePath, filename string) {
	filePath := filepath.Join(includePath, filename)
	log.Infof("Loading included config file: %s", filePath)

	if err := k.Load(file.Provider(filePath), yaml.Parser()); err != nil {
		log.Errorf("Error loading included config file %s: %v", filePath, err)
		return
	}

	if err := k.Unmarshal(".", cfg); err != nil {
		log.Errorf("Error unmarshalling included config file %s: %v", filePath, err)
		return
	}

	log.Infof("Successfully loaded %s", filename)
}

/**
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
*/
