//go:build !windows
// +build !windows

package servicehost

import (
	log "github.com/sirupsen/logrus"
)

func Start(mode string) {
	log.Debugf("servicehost nonwin")
}

func GetConfigFilePath() string {
	log.Debugf("servicehost nonwin")
	return "../"
}
