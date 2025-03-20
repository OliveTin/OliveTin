package filehelper

import (
	log "github.com/sirupsen/logrus"
	"os"
	"sync"
)

var writeFileMutex sync.Mutex

func WriteFile(filename string, out []byte) {
	writeFileMutex.Lock()

	defer writeFileMutex.Unlock()

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		handle, err := os.Create(filename)
		handle.Close()

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
