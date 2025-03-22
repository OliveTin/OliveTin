package filehelper

import (
	log "github.com/sirupsen/logrus"
	"os"
)

func Touch(filename string, description string) {
	_, err := os.Stat(filename)

	if os.IsNotExist(err) {
		_, err := os.Create(filename)

		if err != nil {
			log.Warnf("Could not create %v: %v", description, filename)
		} else {
			log.Infof("Created %v: %v", description, filename)
		}
	}
}
