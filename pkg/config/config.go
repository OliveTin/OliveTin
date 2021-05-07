package config

import (
	log "github.com/sirupsen/logrus"
	"strings"
)

type ActionButton struct {
	Title string
	Icon string
	Css map[string]string `mapstructure:omitempty`
}

type Entity struct {
	Title string
	Icon string
	ActionButtons []ActionButton `mapstructure:"actions"`
	Css map[string]string
}

type Config struct {
	ListenPortRestActions int
	ListenPortWebUi int
	LogLevel string
	ActionButtons []ActionButton `mapstructure:"actions"`
	Entities []Entity `mapstructure:omitempty`
}

func DefaultConfig() *Config {
	config := Config{};
	config.ListenPortRestActions = 1337;
	config.ListenPortWebUi = 1339;
	config.LogLevel = "INFO"

	return &config
}

func (cfg *Config) GetLogLevel() (log.Level) {
	switch (strings.ToUpper(cfg.LogLevel)) {
	case "INFO": return log.InfoLevel;
	case "WARN": return log.WarnLevel;
	case "DEBUG": return log.DebugLevel;
	default: return log.DebugLevel; 
	}
}
