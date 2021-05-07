package main

import (
	log "github.com/sirupsen/logrus"

	restApi "github.com/jamesread/OliveTin/pkg/restApi"
	grpcApi "github.com/jamesread/OliveTin/pkg/grpcApi"

	config "github.com/jamesread/OliveTin/pkg/config"
	"github.com/spf13/viper"
	"os"
)

var (
	cfg *config.Config;
)

func init() {
	log.Info("OliveTin initializing");
	log.SetLevel(log.DebugLevel) // Default to debug, to catch cfg issues

	viper.AutomaticEnv();
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/olivetin/")

	if err := viper.ReadInConfig(); err != nil {
		log.Panicf("Config file error %s", err);
	};

	cfg = config.DefaultConfig();

	if err := viper.UnmarshalExact(cfg); err != nil {
		log.Errorf("Config unmarshal error %+v", err)
		os.Exit(1);
	}

	log.SetLevel(cfg.GetLogLevel())

	viper.WatchConfig();
}

func main() {
	log.WithFields(log.Fields {
		"listenPortRestActions": cfg.ListenPortRestActions,
		"listenPortWebUi": cfg.ListenPortWebUi,
	}).Info("OliveTin started");

	log.Debugf("%+v", cfg)

	go grpcApi.Start()

	restApi.Start();
}
