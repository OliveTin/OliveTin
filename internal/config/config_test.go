package config

import (
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetLog(t *testing.T) {
	c := Config{}
	c.LogLevel = ""

	assert.Equal(t, c.GetLogLevel(), log.InfoLevel, "Info log level should be default")

	c.LogLevel = "INFO"
	assert.Equal(t, c.GetLogLevel(), log.InfoLevel, "set info log level")

	c.LogLevel = "WARN"
	assert.Equal(t, c.GetLogLevel(), log.WarnLevel, "set warn log level")

	c.LogLevel = "DEBUG"
	assert.Equal(t, c.GetLogLevel(), log.DebugLevel, "set debug log level")
}

func TestCreateDefaultConfig(t *testing.T) {
	c := DefaultConfig()

	assert.NotNil(t, c, "Create a default config")
}
