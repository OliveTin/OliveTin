package main

import (
	"flag"
	"fmt"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/OliveTin/OliveTin/internal/api"
	"github.com/OliveTin/OliveTin/internal/auth"
	"github.com/OliveTin/OliveTin/internal/entities"
	"github.com/OliveTin/OliveTin/internal/executor"
	"github.com/OliveTin/OliveTin/internal/httpservers"
	"github.com/OliveTin/OliveTin/internal/installationinfo"
	"github.com/OliveTin/OliveTin/internal/oncalendarfile"
	"github.com/OliveTin/OliveTin/internal/oncron"
	"github.com/OliveTin/OliveTin/internal/onfileindir"
	"github.com/OliveTin/OliveTin/internal/onstartup"
	"github.com/OliveTin/OliveTin/internal/servicehost"
	updatecheck "github.com/OliveTin/OliveTin/internal/updatecheck"

	"os"
	"strconv"

	config "github.com/OliveTin/OliveTin/internal/config"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

var (
	cfg     *config.Config
	cli     cliOptions
	version = "dev"
	commit  = "nocommit"
	date    = "nodate"

	syntaxCheckSearchResults []string
	syntaxCheckLoadError     error
)

type cliOptions struct {
	configDir    string
	syntaxCheck  bool
	startupTrace bool
	printVersion bool
}

func init() {
	// Parse flags before initLog so -startuptrace can enable TRACE before config loads.
	cli = parseCliFlags()
	initLog()
	handleEarlyCliFlags()

	initConfig(cli.configDir, !cli.syntaxCheck)
	quietSyntaxCheckLogs()

	initCheckEnvironment()

	initInstallationInfo()

	log.Info("OliveTin initialization complete")
}

func initLog() {
	logFormat := os.Getenv("OLIVETIN_LOG_FORMAT")

	if logFormat == "json" {
		log.SetFormatter(&log.JSONFormatter{})
	} else {
		log.SetFormatter(&log.TextFormatter{
			ForceQuote:       true,
			DisableTimestamp: true,
		})
	}

	if cli.startupTrace {
		log.SetLevel(log.TraceLevel)
		return
	}

	log.SetLevel(log.InfoLevel)
}

func parseCliFlags() cliOptions {
	options := cliOptions{}
	flag.StringVar(&options.configDir, "configdir", ".", "Config directory path")
	flag.BoolVar(&options.printVersion, "version", false, "Prints the version number and exits")
	flag.BoolVar(&options.syntaxCheck, "syntaxcheck", false, "Check configuration for issues and exit")
	flag.BoolVar(&options.startupTrace, "startuptrace", false, "Enable TRACE logging before configuration is loaded")
	flag.Parse()
	return options
}

func handleEarlyCliFlags() {
	// This log message should be the first log message OliveTin prints.
	if cli.printVersion {
		logStartupMessage("OliveTin is just printing the startup message")
		os.Exit(1)
	}

	if cli.syntaxCheck {
		// Suppress routine logs before any further startup output. Re-applied
		// after config load because sanitize may raise the level again.
		log.SetLevel(log.FatalLevel)
		return
	}

	logStartupMessage("OliveTin initializing")

	log.WithFields(log.Fields{
		"value": cli.configDir,
	}).Debugf("Value of -configdir flag")
}

// quietSyntaxCheckLogs keeps -syntaxcheck output limited to the report on
// stdout. Load failures are returned to the syntax-check report instead of
// exiting immediately via Fatalf.
func quietSyntaxCheckLogs() {
	if cli.syntaxCheck {
		log.SetLevel(log.FatalLevel)
	}
}

func failConfigLoad(format string, args ...any) {
	err := fmt.Errorf(format, args...)
	if cli.syntaxCheck {
		if syntaxCheckLoadError == nil {
			syntaxCheckLoadError = err
		}
		return
	}

	log.Fatal(err)
}

func getBasePort() int {
	var err error

	defaultPort := 1337
	basePort := defaultPort

	envPort := os.Getenv("PORT")

	if envPort != "" {
		basePort, err = strconv.Atoi(os.Getenv("PORT"))

		if err != nil {
			log.Errorf("Error converting port to int. %s", err)
			os.Exit(1)
		}
	}

	if defaultPort != basePort {
		log.WithFields(log.Fields{
			"basePort": basePort,
		}).Debug("Base port")
	}

	return basePort
}

func getConfigPath(directory string) string {
	joinedPath := filepath.Join(directory, "config.yaml")

	configPath, err := filepath.Abs(joinedPath)

	if err != nil {
		log.WithError(err).Warnf("Error getting absolute path for %s", joinedPath)
		return joinedPath
	}

	return configPath
}

func configSearchDirectories(configDir string) []string {
	directories := []string{configDir}

	// Only load additional configs if not in integration test mode
	absConfigDir, _ := filepath.Abs(configDir)
	if strings.Contains(absConfigDir, "integration-tests") {
		return directories
	}

	return append(directories,
		servicehost.GetConfigFilePath(),
		"/config", // For containers.
		"/etc/OliveTin/",
	)
}

func configPathExists(configPath string) bool {
	_, err := os.Stat(configPath)
	found := err == nil

	log.WithFields(log.Fields{
		"configPath": configPath,
		"found":      found,
	}).Debug("Checking base config path")

	return found
}

func watchConfigFile(k *koanf.Koanf, f *file.File, configPath string) {
	err := f.Watch(func(evt any, err error) {
		log.Infof("config file changed: %v", evt)

		errLoad := k.Load(f, yaml.Parser())
		if errLoad != nil {
			log.WithFields(log.Fields{
				"error": errLoad,
			}).Fatalf("Error loading config file")
		}

		config.AppendSource(cfg, k, configPath)
	})

	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatalf("Error watching config file")
	}
}

func loadConfigFromPath(k *koanf.Koanf, configPath string, watch bool) bool {
	log.WithFields(log.Fields{
		"configPath": configPath,
	}).Info("Loading config from path")

	f := file.Provider(configPath)

	if err := k.Load(f, yaml.Parser()); err != nil {
		failConfigLoad("error loading config from %s: %v", configPath, err)
		return false
	}

	if watch {
		watchConfigFile(k, f, configPath)
	}

	return true
}

func findAndLoadBaseConfig(k *koanf.Koanf, directories []string, watch bool) string {
	for _, directory := range directories {
		configPath := getConfigPath(directory)
		found := configPathExists(configPath)
		printConfigSearchResult(configPath, found)

		if !found {
			continue
		}

		if !loadConfigFromPath(k, configPath, watch) {
			return ""
		}

		return configPath
	}

	return ""
}

func printConfigSearchResult(configPath string, found bool) {
	if !cli.syntaxCheck {
		return
	}

	if found {
		syntaxCheckSearchResults = append(syntaxCheckSearchResults, "Found config file: "+configPath)
		return
	}

	syntaxCheckSearchResults = append(syntaxCheckSearchResults, "Config file not found: "+configPath)
}

func initConfig(configDir string, watch bool) {
	k := koanf.New(".")
	err := k.Load(env.Provider(".", ".", nil), nil)
	if err != nil {
		failConfigLoad("Error loading environment variables: %v", err)
		return
	}

	baseConfigPath := findAndLoadBaseConfig(k, configSearchDirectories(configDir), watch)
	cfg = config.DefaultConfigWithBasePort(getBasePort())

	if baseConfigPath == "" {
		if syntaxCheckLoadError == nil {
			failConfigLoad("No base config file found")
		}
		return
	}

	config.AppendSource(cfg, k, baseConfigPath)
}

func initInstallationInfo() {
	installationinfo.Config = cfg
	installationinfo.Build.Version = version
	installationinfo.Build.Commit = commit
	installationinfo.Build.Date = date
}

func logStartupMessage(message string) {
	log.WithFields(log.Fields{
		"version": version,
		"commit":  commit,
		"date":    date,
	}).Info(message)
}

func initCheckEnvironment() {
	warnIfPuidGuid()
}

func warnIfPuidGuid() {
	if os.Getenv("PUID") != "" || os.Getenv("PGID") != "" {
		log.Warnf("PUID or PGID seem to be set to something, but they are ignored by OliveTin. Please check https://docs.olivetin.app/no-puid-pgid.html")
	}
}

func main() {
	if cli.syntaxCheck {
		runSyntaxCheck()
		return
	}

	servicehost.Start(cfg.ServiceHostMode, cfg.ServiceLogs.Directory)

	log.WithFields(log.Fields{
		"configDir": cfg.GetDir(),
	}).Infof("OliveTin started")

	log.Debugf("Config: %+v", cfg)

	executor := executor.DefaultExecutor(cfg)
	executor.RebuildActionMap()
	config.AddListener(executor.RebuildActionMap)
	config.AddListener(func() {
		entities.SyncEntityFileWatchers(cfg)
	})

	executor.LoadLogsFromDisk()

	api.RegisterExecutorListener(executor)
	entities.AddListener(executor.RebuildActionMap)

	go onstartup.Execute(cfg, executor)
	go oncron.Schedule(cfg, executor)
	go onfileindir.WatchFilesInDirectory(cfg, executor)
	go oncalendarfile.Schedule(cfg, executor)

	go entities.SyncEntityFileWatchers(cfg)

	go updatecheck.StartUpdateChecker(cfg)

	// Load persistent sessions from disk
	auth.LoadUserSessions(cfg)

	httpservers.StartFrontendMux(cfg, executor)
}
