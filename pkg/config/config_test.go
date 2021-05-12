package config

import (
	"testing"
	log "github.com/sirupsen/logrus"
)

func TestGetLog(t *testing.T) {
	c := Config{}
	c.LogLevel = ""

	if c.GetLogLevel() != log.InfoLevel {
		t.Errorf("Info expected")
	}
}
