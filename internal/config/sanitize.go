package config

import (
	log "github.com/sirupsen/logrus"
)

func Sanitize(cfg *Config) {
	sanitizeLogLevel(cfg);

	for _, action := range cfg.ActionButtons {
		sanitizeAction(action)
	}
}

func sanitizeLogLevel(cfg *Config) {
	if logLevel, err := log.ParseLevel(cfg.LogLevel); err == nil {
		log.Info("lvl", logLevel)
		log.SetLevel(logLevel)
	}
}

func sanitizeAction(action ActionButton) {
	for _, argument := range action.Arguments {
		sanitizeActionArgument(argument)
	}
}

func sanitizeActionArgument(arg ActionArgument) {
	log.Info("Sanitize AA")
	arg.Label = "foo"
	arg.Name = "blat"
}
