package config

import (
	log "github.com/sirupsen/logrus"
	"strings"
)

type ActionButton struct {
	Title   string
	Icon    string
	Shell   string
	CSS     map[string]string `mapstructure:"omitempty"`
	Timeout int
}

type Entity struct {
	Title         string
	Icon          string
	ActionButtons []ActionButton `mapstructure:"actions"`
	CSS           map[string]string
}

type Config struct {
	ListenAddressWebUI       string
	ListenAddressRestActions string
	ListenAddressGrpcActions string
	LogLevel                 string
	ActionButtons            []ActionButton `mapstructure:"actions"`
	Entities                 []Entity       `mapstructure:"omitempty"`
}

func DefaultConfig() *Config {
	config := Config{}
	config.ListenAddressWebUI = "0.0.0.0:1337"
	config.ListenAddressRestActions = "0.0.0.0:1338"
	config.ListenAddressGrpcActions = "0.0.0.0:1339"
	config.LogLevel = "INFO"

	return &config
}

func (cfg *Config) GetLogLevel() log.Level {
	switch strings.ToUpper(cfg.LogLevel) {
	case "INFO":
		return log.InfoLevel
	case "WARN":
		return log.WarnLevel
	case "DEBUG":
		return log.DebugLevel
	default:
		return log.InfoLevel
	}
}
