package main

import (
	log "github.com/sirupsen/logrus"

	grpcapi "github.com/jamesread/OliveTin/internal/grpcapi"
	restapi "github.com/jamesread/OliveTin/internal/restapi"
	webuiServer "github.com/jamesread/OliveTin/internal/webuiServer"

	config "github.com/jamesread/OliveTin/internal/config"
	"github.com/spf13/viper"
	"os"
)

var (
	cfg *config.Config
)

func init() {
	log.Info("OliveTin initializing")
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

	log.SetLevel(cfg.GetLogLevel())

	viper.WatchConfig()
}

func main() {
	log.WithFields(log.Fields{
		"listenAddressGrpcActions": cfg.ListenAddressGrpcActions,
		"listenAddressRestActions": cfg.ListenAddressRestActions,
		"listenAddressWebUI":       cfg.ListenAddressWebUI,
	}).Info("OliveTin started")

	log.Debugf("%+v", cfg)

	go grpcapi.Start(cfg.ListenAddressGrpcActions, cfg)
	go restapi.Start(cfg.ListenAddressRestActions, cfg.ListenAddressGrpcActions, cfg)

	webuiServer.Start(cfg.ListenAddressWebUI, cfg.ListenAddressRestActions)
}
