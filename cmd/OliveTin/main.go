package main

import (
	"flag"

	log "github.com/sirupsen/logrus"

	"github.com/OliveTin/OliveTin/internal/executor"
	grpcapi "github.com/OliveTin/OliveTin/internal/grpcapi"
	"github.com/OliveTin/OliveTin/internal/installationinfo"
	"github.com/OliveTin/OliveTin/internal/oncron"
	"github.com/OliveTin/OliveTin/internal/onfileindir"
	"github.com/OliveTin/OliveTin/internal/onstartup"
	updatecheck "github.com/OliveTin/OliveTin/internal/updatecheck"
	"github.com/OliveTin/OliveTin/internal/websocket"

	"github.com/OliveTin/OliveTin/internal/httpservers"

	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"os"
	"path"
)

var (
	cfg     *config.Config
	version = "dev"
	commit  = "nocommit"
	date    = "nodate"
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		ForceQuote:       true,
		DisableTimestamp: true,
	})

	log.WithFields(log.Fields{
		"version": version,
		"commit":  commit,
		"date":    date,
	}).Info("OliveTin initializing")

	log.SetLevel(log.DebugLevel) // Default to debug, to catch cfg issues

	var configDir string
	flag.StringVar(&configDir, "configdir", ".", "Config directory path")
	flag.Parse()

	log.WithFields(log.Fields{
		"value": configDir,
	}).Debugf("Value of -configdir flag")

	viper.AutomaticEnv()
	viper.SetConfigName("config.yaml")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configDir)
	viper.AddConfigPath("/config") // For containers.
	viper.AddConfigPath("/etc/OliveTin/")

	if err := viper.ReadInConfig(); err != nil {
		log.Errorf("Config file error at startup. %s", err)
		os.Exit(1)
	}

	cfg = config.DefaultConfig()

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		if e.Op == fsnotify.Write {
			log.Info("Config file changed:", e.String())

			reloadConfig()
		}
	})

	reloadConfig()

	warnIfPuidGuid()

	installationinfo.Config = cfg
	installationinfo.Build.Version = version
	installationinfo.Build.Commit = commit
	installationinfo.Build.Date = date

	log.Info("Init complete")
}

func warnIfPuidGuid() {
	if os.Getenv("PUID") != "" || os.Getenv("PGID") != "" {
		log.Warnf("PUID or PGID seem to be set to something, but they are ignored by OliveTin. Please check https://docs.olivetin.app/no-puid-pgid.html")
	}
}

func reloadConfig() {
	if err := viper.UnmarshalExact(&cfg); err != nil {
		log.Errorf("Config unmarshal error %+v", err)
		os.Exit(1)
	}

	cfg.Sanitize()
}

func main() {
	configDir := path.Dir(viper.ConfigFileUsed())

	log.WithFields(log.Fields{
		"configDir": configDir,
	}).Infof("OliveTin started")

	log.Debugf("Config: %+v", cfg)

	executor := executor.DefaultExecutor()
	executor.AddListener(websocket.ExecutionListener)

	go onstartup.Execute(cfg, executor)
	go oncron.Schedule(cfg, executor)
	go onfileindir.WatchFilesInDirectory(cfg, executor)

	go updatecheck.StartUpdateChecker(version, commit, cfg, configDir)

	go grpcapi.Start(cfg, executor)

	httpservers.StartServers(cfg)
}
