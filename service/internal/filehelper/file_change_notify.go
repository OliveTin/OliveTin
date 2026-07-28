package filehelper

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/OliveTin/OliveTin/internal/configissues"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
)

var (
	debounceWriteLog map[string]*FsNotifyLogEntry

	debounceWriteLogMutex = sync.Mutex{}
)

func init() {
	debounceWriteLog = make(map[string]*FsNotifyLogEntry)
}

type FsNotifyLogEntry struct {
	callbackWrapper  *time.Timer
	callbackComplete bool
}

const (
	debounceDelay = 300 * time.Millisecond
)

// WatchMeta attaches action context to watcher setup failures so Diagnostics
// can apply the same action view ACL as other configuration issues.
type WatchMeta struct {
	ActionID    string
	ActionTitle string
	ConfigFile  string
}

type watchContext struct {
	filename        string
	filedir         string
	callback        func(filename string)
	interestedEvent fsnotify.Op
	event           *fsnotify.Event
	meta            WatchMeta
}

func WatchDirectoryCreate(fullpath string, callback func(filename string), meta WatchMeta) {
	watchPath(&watchContext{
		filedir:         fullpath,
		filename:        "",
		callback:        callback,
		interestedEvent: fsnotify.Create,
		meta:            meta,
	})
}

func WatchDirectoryWrite(fullpath string, callback func(filename string), meta WatchMeta) {
	watchPath(&watchContext{
		filedir:         fullpath,
		filename:        "",
		callback:        callback,
		interestedEvent: fsnotify.Write,
		meta:            meta,
	})
}

func WatchFileWrite(fullpath string, callback func(filename string), meta WatchMeta) {
	filename := filepath.Base(fullpath)
	filedir := filepath.Dir(fullpath)

	watchPath(&watchContext{
		filedir:         filedir,
		filename:        filename,
		callback:        callback,
		interestedEvent: fsnotify.Write,
		meta:            meta,
	})
}

func watchPath(ctx *watchContext) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		reportWatcherFailure(ctx, err)
		return
	}

	defer closeWatcher(watcher)

	if err := watcher.Add(ctx.filedir); err != nil {
		reportWatcherFailure(ctx, err)
		return
	}

	for processEvent(ctx, watcher) {
	}
}

func closeWatcher(watcher *fsnotify.Watcher) {
	if err := watcher.Close(); err != nil {
		log.Errorf("Failed to close file watcher: %v", err)
	}
}

func watchTarget(ctx *watchContext) string {
	if ctx.filename == "" {
		return ctx.filedir
	}
	return filepath.Join(ctx.filedir, ctx.filename)
}

func reportWatcherFailure(ctx *watchContext, err error) {
	path := watchTarget(ctx)
	message := fmt.Sprintf("Could not create watcher for %q: %v", path, err)
	configissues.Report(configissues.Issue{
		Severity:    configissues.SeverityError,
		Code:        configissues.CodeWatcherPath,
		Message:     message,
		ActionID:    ctx.meta.ActionID,
		ActionTitle: ctx.meta.ActionTitle,
		Source:      path,
		ConfigFile:  ctx.meta.ConfigFile,
	})
}

// processEvent waits for one watcher event. It returns false when the watcher
// channels are closed so the caller can stop looping.
func processEvent(ctx *watchContext, watcher *fsnotify.Watcher) bool {
	select {
	case event, ok := <-watcher.Events:
		return handleWatcherEvent(ctx, event, ok)
	case err, ok := <-watcher.Errors:
		return handleWatcherError(ctx, err, ok)
	}
}

func handleWatcherEvent(ctx *watchContext, event fsnotify.Event, ok bool) bool {
	if !ok {
		return false
	}
	ctx.event = &event
	consumeEvent(ctx)
	return true
}

func handleWatcherError(ctx *watchContext, err error, ok bool) bool {
	if !ok {
		return false
	}
	if err != nil {
		log.WithFields(log.Fields{
			"path": watchTarget(ctx),
		}).Errorf("Error in fsnotify: %v", err)
	}
	return true
}

func consumeEvent(ctx *watchContext) {
	if ctx.filename != "" && filepath.Base(ctx.event.Name) != ctx.filename {
		log.Tracef("fsnotify irreleventa event different file %+v", ctx.event)
		return
	}

	consumeRelevantEvents(ctx)
}

func consumeRelevantEvents(ctx *watchContext) {
	if ctx.event.Has(ctx.interestedEvent) {
		log.Debugf("fsnotify event relevant: %v", ctx.event)

		processDebounce(ctx)
	} else {
		log.Debugf("fsnotify event irrelevant: %v", ctx.event)
	}
}

func processDebounce(ctx *watchContext) {
	debounceWriteLogMutex.Lock()

	logEntry, found := debounceWriteLog[ctx.filename]

	if !found {
		logEntry = &FsNotifyLogEntry{
			callbackComplete: false,
			callbackWrapper:  nil,
		}

		debounceWriteLog[ctx.filename] = logEntry
	}

	log.Debugf("fsnotify event %+v", logEntry)

	if logEntry.callbackComplete || logEntry.callbackWrapper == nil {
		log.Debugf("fsnotify event callback queued within debounce delay: %v", ctx.filename)

		logEntry.callbackComplete = false
		logEntry.callbackWrapper = time.AfterFunc(debounceDelay, func() {
			log.Debugf("fsnotify event callback being fired: %v", ctx.filename)

			ctx.callback(ctx.event.Name)

			logEntry.callbackComplete = true
		})
	} else {
		log.Debugf("fsnotify event suppressed because it's within the debounce delay: %v", ctx.filename)
	}

	debounceWriteLogMutex.Unlock()
}
