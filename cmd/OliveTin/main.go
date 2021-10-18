package main

import (
	log "github.com/sirupsen/logrus"

	grpcapi "github.com/jamesread/OliveTin/internal/grpcapi"
	updatecheck "github.com/jamesread/OliveTin/internal/updatecheck"

	"github.com/jamesread/OliveTin/internal/httpservers"

	"github.com/fsnotify/fsnotify"
	config "github.com/jamesread/OliveTin/internal/config"
	"github.com/spf13/viper"
	"os"
)

var (
	cfg     *config.Config
	version = "dev"
	commit  = "nocommit"
	date    = "nodate"
)

func init() {
	log.WithFields(log.Fields{
		"version": version,
		"commit":  commit,
		"date":    date,
	}).Info("OliveTin initializing")

	log.SetLevel(log.DebugLevel) // Default to debug, to catch cfg issues

	viper.AutomaticEnv()
	viper.SetConfigName("config.yaml")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/config") // For containers.
	viper.AddConfigPath("/etc/OliveTin/")

	if err := viper.ReadInConfig(); err != nil {
		log.Errorf("Config file error at startup. %s", err)
		os.Exit(1)
	}

	cfg = config.DefaultConfig()

	reloadConfig()

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		if e.Op == fsnotify.Write {
			log.Info("Config file changed:", e.String())

			reloadConfig()
		}
	})
}

func reloadConfig() {
	if err := viper.UnmarshalExact(&cfg); err != nil {
		log.Errorf("Config unmarshal error %+v", err)
		os.Exit(1)
	}

	config.Sanitize(cfg);
}

func main() {
	log.Info("OliveTin started")

	log.Debugf("Config: %+v", cfg)

	go updatecheck.StartUpdateChecker(version, commit, cfg)

	go grpcapi.Start(cfg)

	httpservers.StartServers(cfg)
}
