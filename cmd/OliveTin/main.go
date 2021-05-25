package main

import (
	log "github.com/sirupsen/logrus"

	grpcapi "github.com/jamesread/OliveTin/internal/grpcapi"

	"github.com/jamesread/OliveTin/internal/httpservers"

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

	if err := viper.UnmarshalExact(cfg); err != nil {
		log.Errorf("Config unmarshal error %+v", err)
		os.Exit(1)
	}

	if logLevel, err := log.ParseLevel(cfg.LogLevel); err == nil {
		log.SetLevel(logLevel)
	}

	viper.WatchConfig()
}

func main() {
	log.Info("OliveTin started")

	log.Debugf("%+v", cfg)

	go grpcapi.Start(cfg)

	httpservers.StartServers(cfg)
}
