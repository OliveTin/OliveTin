package main

import (
	"flag"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

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
	version = "dev"
	commit  = "nocommit"
	date    = "nodate"
)

func init() {
	initLog()

	initConfig(initCliFlags())

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

	// Use debug this early on to catch details about startup errors. The
	// default config will raise the log level later, if not set.
	log.SetLevel(log.DebugLevel) // Default to debug, to catch cfg issue
}

func initCliFlags() string {
	var configDir string
	flag.StringVar(&configDir, "configdir", ".", "Config directory path")

	var printVersion bool
	flag.BoolVar(&printVersion, "version", false, "Prints the version number and exits")
	flag.Parse()

	// This log message should be the first log message OliveTin prints.
	if printVersion {
		logStartupMessage("OliveTin is just printing the startup message")
		os.Exit(1)
	} else {
		logStartupMessage("OliveTin initializing")
	}

	log.WithFields(log.Fields{
		"value": configDir,
	}).Debugf("Value of -configdir flag")

	return configDir
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

func initConfig(configDir string) {
	k := koanf.New(".")
	k.Load(env.Provider(".", ".", nil), nil)

	directories := []string{
		configDir,
	}

	// Only load additional configs if not in integration test mode
	absConfigDir, _ := filepath.Abs(configDir)
	if !strings.Contains(absConfigDir, "integration-tests") {
		directories = append(directories,
			servicehost.GetConfigFilePath(),
			"/config", // For containers.
			"/etc/OliveTin/",
		)
	}

	var baseConfigPath string

	for _, directory := range directories {
		configPath := getConfigPath(directory)

		found := true
		if _, err := os.Stat(configPath); err != nil {
			found = false
		}

		log.WithFields(log.Fields{
			"configPath": configPath,
			"found":      found,
		}).Debug("Checking base config path")

		if !found {
			continue
		}

		if baseConfigPath == "" {
			baseConfigPath = configPath
		}

		log.WithFields(log.Fields{
			"configPath": configPath,
		}).Info("Loading config from path")

		f := file.Provider(configPath)

		if err := k.Load(f, yaml.Parser()); err != nil {
			log.Fatalf("error loading config from %s: %v", configPath, err)
			os.Exit(1)
		}

		f.Watch(func(evt interface{}, err error) {
			log.Infof("config file changed: %v", evt)

			k.Load(f, yaml.Parser())
			config.AppendSource(cfg, k, configPath)
		})

		break
	}

	cfg = config.DefaultConfigWithBasePort(getBasePort())

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
	servicehost.Start(cfg.ServiceHostMode)

	log.WithFields(log.Fields{
		"configDir": cfg.GetDir(),
	}).Infof("OliveTin started")

	log.Debugf("Config: %+v", cfg)

	executor := executor.DefaultExecutor(cfg)
	executor.RebuildActionMap()
	config.AddListener(executor.RebuildActionMap)

	go onstartup.Execute(cfg, executor)
	go oncron.Schedule(cfg, executor)
	go onfileindir.WatchFilesInDirectory(cfg, executor)
	go oncalendarfile.Schedule(cfg, executor)

	entities.AddListener(executor.RebuildActionMap)
	go entities.SetupEntityFileWatchers(cfg)

	go updatecheck.StartUpdateChecker(cfg)

	// Load persistent sessions from disk
	auth.LoadUserSessions(cfg)

	httpservers.StartServers(cfg, executor)
}
