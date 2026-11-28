package filehelper

import (
	"fmt"
	"os"
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
	callback func(filename string)
	event    *fsnotify.Event
	watcher  *fsnotify.Watcher
	meta     WatchMeta
	filename string
	filedir  string

	// The fields below are only used by file watches, and only touched by the
	// watcher's own goroutine once it has started.

	// watchedPath is the file that was asked for, which may be a symlink.
	watchedPath string
	// resolvedPath is where watchedPath last pointed, or empty if it has not
	// resolved yet.
	resolvedPath string
	// resolvedDir is the extra directory watched when resolvedPath is in a
	// different directory to watchedPath, so in-place edits of a symlink's
	// target are seen.
	resolvedDir string

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
		watchedPath:     watchKey,
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

	ctx.watcher = watcher
	refreshResolvedPath(ctx)

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
		return !isStopped(done) && handleWatcherEvent(ctx, event, ok)
	case err, ok := <-watcher.Errors:
		return handleWatcherError(ctx, err, ok)
	}
}

// isStopped stops a watch that has been asked to stop even when an event was
// already waiting, because select picks between ready channels at random.
func isStopped(done <-chan struct{}) bool {
	select {
	case <-done:
		return true
	default:
		return false
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
	if ctx.watchedPath != "" {
		consumeFileEvent(ctx)
		return
	}

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

// consumeFileEvent decides whether an event means the watched file's content
// may have changed. Writing the file in place gives a Write event, but tools
// that replace a file give other events: editors and mv rename a new file over
// it (Create), and Kubernetes updates a mounted ConfigMap by swapping a
// directory symlink, so the file itself gets no event at all and only the
// target of its symlink changes.
func consumeFileEvent(ctx *watchContext) {
	if isWatchedFileChange(ctx) {
		consumeWatchedFileChange(ctx)
		return
	}

	if isDirectoryChange(ctx.event) && refreshResolvedPath(ctx) {
		log.Debugf("fsnotify symlink target changed for %v: %v", ctx.watchedPath, ctx.event)
		processDebounce(ctx)
		return
	}

	log.Tracef("fsnotify event irrelevant: %v", ctx.event)
}

// consumeWatchedFileChange reloads after a write or create at the watched
// path. A create can leave a dangling symlink, so it only reloads once the
// file can be read. A write always reloads, as it always has.
func consumeWatchedFileChange(ctx *watchContext) {
	_, resolved := resolveWatchedPath(ctx)
	if !resolved && !ctx.event.Has(fsnotify.Write) {
		log.Debugf("fsnotify event waiting for %v to resolve: %v", ctx.watchedPath, ctx.event)
		return
	}

	log.Debugf("fsnotify event relevant: %v", ctx.event)
	processDebounce(ctx)
}

func isWatchedFileChange(ctx *watchContext) bool {
	if !ctx.event.Has(fsnotify.Write) && !ctx.event.Has(fsnotify.Create) {
		return false
	}

	name := filepath.Clean(ctx.event.Name)
	return name == ctx.watchedPath || (ctx.resolvedPath != "" && name == ctx.resolvedPath)
}

func isDirectoryChange(event *fsnotify.Event) bool {
	return event.Has(fsnotify.Create) || event.Has(fsnotify.Rename) || event.Has(fsnotify.Remove)
}

// refreshResolvedPath checks where the watched file points now, and reports
// whether that changed. While the file is missing or a symlink is dangling,
// for example part way through a replacement, the last target is kept and no
// change is reported, so the file is not reloaded until it can be read.
func refreshResolvedPath(ctx *watchContext) bool {
	changed, _ := resolveWatchedPath(ctx)
	return changed
}

// resolveWatchedPath updates where the watched file points. It reports
// whether the target changed, and separately whether the file resolved at
// all.
func resolveWatchedPath(ctx *watchContext) (changed bool, resolved bool) {
	target, err := filepath.EvalSymlinks(ctx.watchedPath)
	if err != nil {
		log.Debugf("fsnotify could not resolve %v: %v", ctx.watchedPath, err)
		watchDanglingTargetDir(ctx)
		return false, false
	}

	// The target's directory is watched again even when the target has not
	// changed, because a dangling symlink in between may have moved the extra
	// watch to the directory it pointed into.
	watchResolvedDir(ctx, filepath.Dir(target))

	if target == ctx.resolvedPath {
		return false, true
	}

	ctx.resolvedPath = target
	return true, true
}

// watchDanglingTargetDir watches the directory a dangling symlink points into.
// Directory watches are not recursive, so when that directory is not already
// watched, creating the missing target there would give no event, and the file
// would not be reloaded once it can be read.
func watchDanglingTargetDir(ctx *watchContext) {
	target, err := os.Readlink(ctx.watchedPath)
	if err != nil {
		return
	}

	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(ctx.watchedPath), target)
	}

	dir := filepath.Dir(target)
	if _, err := os.Stat(dir); err != nil {
		return
	}

	watchResolvedDir(ctx, dir)
}

func watchResolvedDir(ctx *watchContext, dir string) {
	if dir == ctx.resolvedDir {
		return
	}

	unwatchResolvedDir(ctx)

	// The directory may be the watched directory reached through a symlink.
	// inotify gives both paths the same watch, so adding it again would rename
	// that watch, and removing it later would stop the main watch.
	if sameDirectory(dir, ctx.filedir) {
		return
	}

	if err := ctx.watcher.Add(dir); err != nil {
		log.WithFields(log.Fields{
			"path": ctx.watchedPath,
			"dir":  dir,
		}).Warnf("Could not watch the directory a symlink points to, so edits there will not be seen: %v", err)
		return
	}

	ctx.resolvedDir = dir
}

func unwatchResolvedDir(ctx *watchContext) {
	if ctx.resolvedDir == "" {
		return
	}

	// A ConfigMap's old directory has already been deleted by now, and the
	// watch went with it, so an error here is expected.
	if err := ctx.watcher.Remove(ctx.resolvedDir); err != nil {
		log.Debugf("fsnotify could not stop watching %v: %v", ctx.resolvedDir, err)
	}

	ctx.resolvedDir = ""
}

func sameDirectory(left, right string) bool {
	leftInfo, err := os.Stat(left)
	if err != nil {
		return false
	}

	rightInfo, err := os.Stat(right)
	if err != nil {
		return false
	}

	return os.SameFile(leftInfo, rightInfo)
}

// debounceKey keeps file watches apart by their whole path, so two files with
// the same name in different directories do not suppress each other.
// Directory watches keep their existing key.
func debounceKey(ctx *watchContext) string {
	if ctx.watchedPath != "" {
		return ctx.watchedPath
	}
	return ctx.filename
}

func processDebounce(ctx *watchContext) {
	debounceWriteLogMutex.Lock()

	key := debounceKey(ctx)
	logEntry, found := debounceWriteLog[key]

	if !found {
		logEntry = &FsNotifyLogEntry{
			callbackComplete: false,
			callbackWrapper:  nil,
		}

		debounceWriteLog[key] = logEntry
	}

	log.Debugf("fsnotify event %+v", logEntry)

	if logEntry.callbackComplete || logEntry.callbackWrapper == nil {
		log.Debugf("fsnotify event callback queued within debounce delay: %v", ctx.filename)

		callback := ctx.callback
		eventName := ctx.event.Name
		logEntry.callbackComplete = false
		logEntry.callbackWrapper = time.AfterFunc(debounceDelay, func() {
			log.Debugf("fsnotify event callback being fired: %v", eventName)

			callback(eventName)

			debounceWriteLogMutex.Lock()
			logEntry.callbackComplete = true
			debounceWriteLogMutex.Unlock()
		})
	} else {
		log.Debugf("fsnotify event suppressed because it's within the debounce delay: %v", ctx.filename)
	}

	debounceWriteLogMutex.Unlock()
}
