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

// AppendSourceWithIncludes loads base config and any included configs
func AppendSourceWithIncludes(cfg *Config, k *koanf.Koanf, configPath string) {
	// Load base config first
	AppendSource(cfg, k, configPath)

	// Load included configs if specified
	if cfg.Include != "" {
		LoadIncludedConfigs(cfg, k, configPath)
	}
}

func AppendSource(cfg *Config, k *koanf.Koanf, configPath string) {
    log.Infof("Appending cfg source: %s", configPath)

    if !unmarshalRoot(k, cfg) {
        return
    }

    loadCollectionsFallbacks(k, cfg)

    applyConfigOverrides(k, cfg)

    afterLoadFinalize(cfg, configPath)
}

func unmarshalRoot(k *koanf.Koanf, cfg *Config) bool {
    if err := k.Unmarshal(".", cfg); err != nil {
        log.Errorf("Error unmarshalling config: %v", err)
        return false
    }
    return true
}

func loadCollectionsFallbacks(k *koanf.Koanf, cfg *Config) {
    maybeUnmarshalActions(k, cfg)
    maybeUnmarshalDashboards(k, cfg)
    maybeUnmarshalEntities(k, cfg)
    maybeUnmarshalAuthLocalUsers(k, cfg)
}

func maybeUnmarshalActions(k *koanf.Koanf, cfg *Config) {
    if len(cfg.Actions) != 0 || !k.Exists("actions") {
        return
    }
    var actions []*Action
    if err := k.Unmarshal("actions", &actions); err == nil {
        cfg.Actions = actions
        log.Debugf("Manually loaded %d actions", len(actions))
    }
}

func maybeUnmarshalDashboards(k *koanf.Koanf, cfg *Config) {
    if len(cfg.Dashboards) != 0 || !k.Exists("dashboards") {
        return
    }
    var dashboards []*DashboardComponent
    if err := k.Unmarshal("dashboards", &dashboards); err == nil {
        cfg.Dashboards = dashboards
        log.Debugf("Manually loaded %d dashboards", len(dashboards))
    }
}

func maybeUnmarshalEntities(k *koanf.Koanf, cfg *Config) {
    if len(cfg.Entities) != 0 || !k.Exists("entities") {
        return
    }
    var entities []*EntityFile
    if err := k.Unmarshal("entities", &entities); err == nil {
        cfg.Entities = entities
        log.Debugf("Manually loaded %d entities", len(entities))
    }
}

func maybeUnmarshalAuthLocalUsers(k *koanf.Koanf, cfg *Config) {
    if len(cfg.AuthLocalUsers.Users) != 0 || !k.Exists("authLocalUsers") {
        return
    }
    var authLocalUsers AuthLocalUsersConfig
    if err := k.Unmarshal("authLocalUsers", &authLocalUsers); err == nil {
        cfg.AuthLocalUsers = authLocalUsers
        log.Debugf("Manually loaded local auth config")
    }
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

func applyConfigOverrides(k *koanf.Koanf, cfg *Config) {
	// Override fields that should be read from config
	// mapstructure tags should make most of this unnecessary, but keep for safety
	boolVal(k, "showFooter", &cfg.ShowFooter)
	boolVal(k, "showNavigation", &cfg.ShowNavigation)
	boolVal(k, "checkForUpdates", &cfg.CheckForUpdates)
	boolVal(k, "useSingleHTTPFrontend", &cfg.UseSingleHTTPFrontend)
	stringVal(k, "logLevel", &cfg.LogLevel)
	stringVal(k, "pageTitle", &cfg.PageTitle)
	boolVal(k, "authRequireGuestsToLogin", &cfg.AuthRequireGuestsToLogin)
	stringVal(k, "include", &cfg.Include)

	// Handle nested defaultPolicy struct
	if k.Exists("defaultPolicy") {
		boolVal(k, "defaultPolicy.showDiagnostics", &cfg.DefaultPolicy.ShowDiagnostics)
		boolVal(k, "defaultPolicy.showLogList", &cfg.DefaultPolicy.ShowLogList)
	}

	// Handle nested prometheus struct
	if k.Exists("prometheus") {
		boolVal(k, "prometheus.enabled", &cfg.Prometheus.Enabled)
		boolVal(k, "prometheus.defaultGoMetrics", &cfg.Prometheus.DefaultGoMetrics)
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
        loadAndMergeIncludedFile(cfg, includePath, filename)
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

func loadAndMergeIncludedFile(cfg *Config, includePath, filename string) {
    filePath := filepath.Join(includePath, filename)
    log.Infof("Loading included config file: %s", filePath)

    includeK := koanf.New(".")
    if err := includeK.Load(file.Provider(filePath), yaml.Parser()); err != nil {
        log.Errorf("Error loading included config file %s: %v", filePath, err)
        return
    }

    tempCfg := &Config{}
    if err := includeK.Unmarshal(".", tempCfg); err != nil {
        log.Errorf("Error unmarshalling included config file %s: %v", filePath, err)
        return
    }
    // Fallbacks similar to AppendSource
    if len(tempCfg.Actions) == 0 && includeK.Exists("actions") {
        var actions []*Action
        if err := includeK.Unmarshal("actions", &actions); err == nil {
            tempCfg.Actions = actions
            log.Debugf("Manually loaded %d actions from %s", len(actions), filename)
        }
    }

    mergeConfig(cfg, tempCfg)
    log.Infof("Successfully loaded and merged %s", filename)
}

func mergeConfig(base *Config, overlay *Config) {
    mergeSlices(base, overlay)
    overrideSimple(base, overlay)
    overrideNested(base, overlay)
    overrideStrings(base, overlay)
}

func mergeSlices(base *Config, overlay *Config) {
    if len(overlay.Actions) > 0 {
        base.Actions = append(base.Actions, overlay.Actions...)
    }
    if len(overlay.Dashboards) > 0 {
        base.Dashboards = append(base.Dashboards, overlay.Dashboards...)
        log.Debugf("Merged %d dashboards from include", len(overlay.Dashboards))
    }
    if len(overlay.Entities) > 0 {
        base.Entities = append(base.Entities, overlay.Entities...)
        log.Debugf("Merged %d entities from include", len(overlay.Entities))
    }
    if len(overlay.AccessControlLists) > 0 {
        base.AccessControlLists = append(base.AccessControlLists, overlay.AccessControlLists...)
        log.Debugf("Merged %d access control lists from include", len(overlay.AccessControlLists))
    }
    if len(overlay.AuthLocalUsers.Users) > 0 {
        base.AuthLocalUsers.Users = append(base.AuthLocalUsers.Users, overlay.AuthLocalUsers.Users...)
        log.Debugf("Merged %d local users from include", len(overlay.AuthLocalUsers.Users))
    }
    if len(overlay.StyleMods) > 0 {
        base.StyleMods = append(base.StyleMods, overlay.StyleMods...)
    }
    if len(overlay.AdditionalNavigationLinks) > 0 {
        base.AdditionalNavigationLinks = append(base.AdditionalNavigationLinks, overlay.AdditionalNavigationLinks...)
    }
}

func overrideSimple(base *Config, overlay *Config) {
    if overlay.LogLevel != "" {
        base.LogLevel = overlay.LogLevel
    }
    if overlay.PageTitle != "" {
        base.PageTitle = overlay.PageTitle
    }
    if overlay.ShowFooter != base.ShowFooter {
        base.ShowFooter = overlay.ShowFooter
    }
    if overlay.ShowNavigation != base.ShowNavigation {
        base.ShowNavigation = overlay.ShowNavigation
    }
    if overlay.CheckForUpdates != base.CheckForUpdates {
        base.CheckForUpdates = overlay.CheckForUpdates
    }
    if overlay.UseSingleHTTPFrontend != base.UseSingleHTTPFrontend {
        base.UseSingleHTTPFrontend = overlay.UseSingleHTTPFrontend
    }
    if overlay.AuthRequireGuestsToLogin != base.AuthRequireGuestsToLogin {
        base.AuthRequireGuestsToLogin = overlay.AuthRequireGuestsToLogin
    }
    if overlay.AuthLocalUsers.Enabled {
        base.AuthLocalUsers.Enabled = overlay.AuthLocalUsers.Enabled
    }
}

func overrideNested(base *Config, overlay *Config) {
    if overlay.DefaultPolicy.ShowDiagnostics != base.DefaultPolicy.ShowDiagnostics {
        base.DefaultPolicy.ShowDiagnostics = overlay.DefaultPolicy.ShowDiagnostics
    }
    if overlay.DefaultPolicy.ShowLogList != base.DefaultPolicy.ShowLogList {
        base.DefaultPolicy.ShowLogList = overlay.DefaultPolicy.ShowLogList
    }
    if overlay.Prometheus.Enabled != base.Prometheus.Enabled {
        base.Prometheus.Enabled = overlay.Prometheus.Enabled
    }
    if overlay.Prometheus.DefaultGoMetrics != base.Prometheus.DefaultGoMetrics {
        base.Prometheus.DefaultGoMetrics = overlay.Prometheus.DefaultGoMetrics
    }
}

func overrideStrings(base *Config, overlay *Config) {
    overrideString(&base.BannerMessage, overlay.BannerMessage)
    overrideString(&base.BannerCSS, overlay.BannerCSS)
    overrideString(&base.LogLevel, overlay.LogLevel)
    overrideString(&base.PageTitle, overlay.PageTitle)
    overrideString(&base.SectionNavigationStyle, overlay.SectionNavigationStyle)
    overrideString(&base.DefaultPopupOnStart, overlay.DefaultPopupOnStart)
}

func overrideString(base *string, overlay string) {
	if overlay != "" {
		*base = overlay
	}
}

func getActionTitles(actions []*Action) []string {
	titles := make([]string, len(actions))
	for i, action := range actions {
		titles[i] = action.Title
	}
	return titles
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
