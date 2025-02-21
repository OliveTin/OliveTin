package filehelper

import (
	log "github.com/sirupsen/logrus"
	"os"
)

func WriteFile(filename string, out []byte) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		_, err := os.Create(filename)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Errorf("Failed to create %v", filename)
			return
		}
	}

	err := os.WriteFile(filename, out, 0600)

	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Errorf("Failed to write session to %v", filename)
		return
	}
}
