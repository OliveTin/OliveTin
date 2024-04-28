package filehelper

import (
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"path/filepath"
)

type watchContext struct {
	filename        string
	filedir         string
	callback        func(filename string)
	interestedEvent fsnotify.Op
}

func WatchDirectoryCreate(fullpath string, callback func(filename string)) {
	watchPath(&watchContext{
		filedir:         fullpath,
		filename:        "",
		callback:        callback,
		interestedEvent: fsnotify.Create,
	})
}

func WatchDirectoryWrite(fullpath string, callback func(filename string)) {
	watchPath(&watchContext{
		filedir:         fullpath,
		filename:        "",
		callback:        callback,
		interestedEvent: fsnotify.Write,
	})
}

func WatchFileWrite(fullpath string, callback func(filename string)) {
	filename := filepath.Base(fullpath)
	filedir := filepath.Dir(fullpath)

	watchPath(&watchContext{
		filedir:         filedir,
		filename:        filename,
		callback:        callback,
		interestedEvent: fsnotify.Write,
	})
}

func watchPath(ctx *watchContext) {
	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		log.Errorf("Could not watch for files being created: %v", err)
		return
	}

	defer watcher.Close()

	done := make(chan bool)

	go func() {
		for {
			processEvent(ctx, watcher)
		}
	}()

	err = watcher.Add(ctx.filedir)

	if err != nil {
		log.Errorf("Could not create watcher: %v", err)
	}

	<-done
}

func processEvent(ctx *watchContext, watcher *fsnotify.Watcher) {
	select {
	case event, ok := <-watcher.Events:
		if !consumeEvent(ok, ctx, &event) {
			return
		}

		break
	case err := <-watcher.Errors:
		log.Errorf("Error in fsnotify: %v", err)
		return
	}
}

func consumeEvent(ok bool, ctx *watchContext, event *fsnotify.Event) bool {
	if !ok {
		return false
	}

	if ctx.filename != "" && filepath.Base(event.Name) != ctx.filename {
		log.Tracef("fsnotify irreleventa event different file %+v", event)
		return true
	}

	consumeRelevantEvents(ctx, event)

	return true
}

func consumeRelevantEvents(ctx *watchContext, event *fsnotify.Event) {
	if event.Has(ctx.interestedEvent) {
		log.Debugf("fsnotify write event: %v", event)
		ctx.callback(event.Name)
	} else {
		log.Debugf("fsnotify irrelevant event on file %v", event)
	}
}
