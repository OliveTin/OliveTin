package main

import (
	"flag"

	log "github.com/sirupsen/logrus"

	"github.com/OliveTin/OliveTin/internal/app"
	"github.com/OliveTin/OliveTin/internal/installationinfo"

	"os"
	"strconv"

	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var (
	cfg     *config.Config
	version = "dev"
	commit  = "nocommit"
	date    = "nodate"
)

func init() {
	initLog()

	initViperConfig(initCliFlags())

	initCheckEnvironment()

	initInstallationInfo()

	log.Info("OliveTin initialization complete")
}

func initLog() {
	log.SetFormatter(&log.TextFormatter{
		ForceQuote:       true,
		DisableTimestamp: true,
	})

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

func initViperConfig(configDir string) {
	viper.AutomaticEnv()
	viper.SetConfigName("config.yaml")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configDir)
	viper.AddConfigPath("../")
	viper.AddConfigPath("/config") // For containers.
	viper.AddConfigPath("/etc/OliveTin/")

	if err := viper.ReadInConfig(); err != nil {
		log.Errorf("Config file error at startup. %s", err)
		os.Exit(1)
	}

	cfg = config.DefaultConfigWithBasePort(getBasePort())

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Infof("Config file changed: %s", viper.ConfigFileUsed())
		config.Reload(cfg)
	})

	config.Reload(cfg)
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
	o := app.CreateOliveTin(cfg)
	o.Start()
}
