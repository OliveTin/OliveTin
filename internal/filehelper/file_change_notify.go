package filehelper

import (
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"path/filepath"
)

func WatchFile(fullpath string, callback func()) {
	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		log.Errorf("Could not watch for files being created: %v", err)
		return
	}

	defer watcher.Close()

	done := make(chan bool)

	filename := filepath.Base(fullpath)
	filedir := filepath.Dir(fullpath)

	go func() {
		for {
			processEvent(filename, watcher, fsnotify.Write, callback)
		}
	}()

	err = watcher.Add(filedir)

	if err != nil {
		log.Errorf("Could not create watcher: %v", err)
	}

	<-done
}

func processEvent(filename string, watcher *fsnotify.Watcher, eventType fsnotify.Op, callback func()) {
	select {
	case event, ok := <-watcher.Events:
		if !consumeEvent(ok, filename, &event, callback) {
			return
		}

		break
	case err := <-watcher.Errors:
		log.Errorf("Error in fsnotify: %v", err)
		return
	}
}

func consumeEvent(ok bool, filename string, event *fsnotify.Event, callback func()) bool {
	if !ok {
		return false
	}

	if filepath.Base(event.Name) != filename {
		log.Tracef("fsnotify irreleventa event different file %+v", event)
		return true
	}

	consumeWriteEvents(event, callback)

	return true
}

func consumeWriteEvents(event *fsnotify.Event, callback func()) {
	if event.Has(fsnotify.Write) {
		log.Debugf("fsnotify write event: %v", event)
		callback()
	} else {
		log.Debugf("fsnotify irrelevant event on file %v", event)
	}
}
