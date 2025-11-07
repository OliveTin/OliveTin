package config

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
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
	log.WithFields(log.Fields{
		"configPath": configPath,
	}).Info("Appending cfg source")

	loadIncludedConfigsFromDir(k, configPath)

	if !unmarshalRoot(k, cfg) {
		return
	}

	afterLoadFinalize(cfg, configPath)
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

// loadIncludedConfigsFromDir loads configuration files from an include directory and merges them
func loadIncludedConfigsFromDir(k *koanf.Koanf, baseConfigPath string) {
	relativeIncludePath := k.String("include")

	if relativeIncludePath == "" {
		return
	}

	includePath := filepath.Join(filepath.Dir(baseConfigPath), relativeIncludePath)

	log.WithFields(log.Fields{
		"includePath": includePath,
	}).Infof("Loading included configs from dir")

	yamlFiles, ok := listYamlFiles(includePath)
	if !ok || len(yamlFiles) == 0 {
		return
	}

	sort.Strings(yamlFiles)
	for _, filename := range yamlFiles {
		loadAndMergeIncludedFile(k, includePath, filename)
	}

	log.Infof("Finished loading %d included config file(s)", len(yamlFiles))
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

func loadAndMergeIncludedFile(k *koanf.Koanf, includePath, filename string) {
	filePath := filepath.Join(includePath, filename)

	if err := k.Load(file.Provider(filePath), yaml.Parser(), koanf.WithMergeFunc(mergeFunc)); err != nil {
		log.Errorf("Error loading included config file %s: %v", filePath, err)
		return
	}

	log.WithFields(log.Fields{
		"filePath": filePath,
	}).Info("Successfully loaded included config file")
}

func mergeFunc(src map[string]interface{}, dest map[string]interface{}) error {
	// Handle actions merging - koanf provides []interface{} not []*Action
	// Merge src (new) into dest (existing) by appending src's actions to dest's actions
	if srcActions, ok := src["actions"]; ok {
		if destActions, ok := dest["actions"]; ok {
			// Both have actions - append src to dest
			srcSlice, ok1 := srcActions.([]interface{})
			destSlice, ok2 := destActions.([]interface{})
			if ok1 && ok2 {
				dest["actions"] = append(destSlice, srcSlice...)
			} else {
				// Fallback: if types don't match, just use src
				dest["actions"] = srcActions
			}
		} else {
			// dest doesn't have actions, so use src's actions
			dest["actions"] = srcActions
		}
	}
	// If src doesn't have actions, leave dest unchanged

	return nil
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
