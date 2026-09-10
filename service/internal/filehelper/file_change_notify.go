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

	fileWatchMu    sync.Mutex
	fileWatchStops = map[string]chan struct{}{}
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
	callback        func(filename string)
	event           *fsnotify.Event
	meta            WatchMeta
	filename        string
	filedir         string
	interestedEvent fsnotify.Op
}

func WatchDirectoryCreate(fullpath string, callback func(filename string), meta WatchMeta) {
	watchPath(&watchContext{
		filedir:         fullpath,
		filename:        "",
		callback:        callback,
		interestedEvent: fsnotify.Create,
		meta:            meta,
	}, nil)
}

func WatchDirectoryWrite(fullpath string, callback func(filename string), meta WatchMeta) {
	watchPath(&watchContext{
		filedir:         fullpath,
		filename:        "",
		callback:        callback,
		interestedEvent: fsnotify.Write,
		meta:            meta,
	}, nil)
}

func WatchFileWrite(fullpath string, callback func(filename string), meta WatchMeta) error {
	filename := filepath.Base(fullpath)
	filedir := filepath.Dir(fullpath)
	watchKey := filepath.Join(filedir, filename)

	ctx := &watchContext{
		filedir:         filedir,
		filename:        filename,
		callback:        callback,
		interestedEvent: fsnotify.Write,
		meta:            meta,
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		reportWatcherFailure(ctx, err)
		return err
	}

	if err := watcher.Add(filedir); err != nil {
		reportWatcherFailure(ctx, err)
		closeWatcher(watcher)
		return err
	}

	done := registerFileWatch(watchKey)
	go runWatcher(ctx, watcher, done)
	return nil
}

// StopFileWatch stops a file write watcher started by WatchFileWrite.
func StopFileWatch(fullpath string) {
	watchKey, err := filepath.Abs(fullpath)
	if err != nil {
		watchKey = fullpath
	}

	fileWatchMu.Lock()
	done, ok := fileWatchStops[watchKey]
	if ok {
		delete(fileWatchStops, watchKey)
	}
	fileWatchMu.Unlock()

	if ok {
		close(done)
	}
}

func registerFileWatch(watchKey string) chan struct{} {
	absKey, err := filepath.Abs(watchKey)
	if err == nil {
		watchKey = absKey
	}

	done := make(chan struct{})
	fileWatchMu.Lock()
	if previous, ok := fileWatchStops[watchKey]; ok {
		close(previous)
	}
	fileWatchStops[watchKey] = done
	fileWatchMu.Unlock()
	return done
}

func watchPath(ctx *watchContext, done <-chan struct{}) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		reportWatcherFailure(ctx, err)
		return
	}

	if err := watcher.Add(ctx.filedir); err != nil {
		reportWatcherFailure(ctx, err)
		closeWatcher(watcher)
		return
	}

	runWatcher(ctx, watcher, done)
}

func runWatcher(ctx *watchContext, watcher *fsnotify.Watcher, done <-chan struct{}) {
	defer closeWatcher(watcher)

	for processEvent(ctx, watcher, done) {
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
func processEvent(ctx *watchContext, watcher *fsnotify.Watcher, done <-chan struct{}) bool {
	if done == nil {
		return waitForWatcherEvent(ctx, watcher)
	}
	return waitForWatcherEventOrCancel(ctx, watcher, done)
}

func waitForWatcherEvent(ctx *watchContext, watcher *fsnotify.Watcher) bool {
	select {
	case event, ok := <-watcher.Events:
		return handleWatcherEvent(ctx, event, ok)
	case err, ok := <-watcher.Errors:
		return handleWatcherError(ctx, err, ok)
	}
}

func waitForWatcherEventOrCancel(ctx *watchContext, watcher *fsnotify.Watcher, done <-chan struct{}) bool {
	select {
	case <-done:
		return false
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
