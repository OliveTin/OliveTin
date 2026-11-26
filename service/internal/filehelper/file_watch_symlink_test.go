package filehelper

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	callbackTimeout = 2 * time.Second
	// quietPeriod is long enough for any debounced callback to have fired.
	quietPeriod = 3 * debounceDelay
)

type watchRecorder struct {
	calls chan string
}

func watchFileForTest(t *testing.T, path string) *watchRecorder {
	t.Helper()

	recorder := &watchRecorder{calls: make(chan string, 16)}
	require.NoError(t, WatchFileWrite(path, func(filename string) {
		recorder.calls <- filename
	}, WatchMeta{}))
	t.Cleanup(func() { StopFileWatch(path) })

	return recorder
}

// expectOneCall waits for a callback, then checks no more follow, so extra
// reloads are caught as well as missing ones.
func (r *watchRecorder) expectOneCall(t *testing.T) {
	t.Helper()

	select {
	case <-r.calls:
	case <-time.After(callbackTimeout):
		t.Fatal("expected a callback, got none")
	}
	r.expectNone(t)
}

func (r *watchRecorder) expectNone(t *testing.T) {
	t.Helper()

	select {
	case name := <-r.calls:
		t.Fatalf("unexpected callback for %s", name)
	case <-time.After(quietPeriod):
	}
}

func skipWithoutSymlinks(t *testing.T) {
	t.Helper()

	if runtime.GOOS == "windows" {
		t.Skip("symlinks need extra privileges on Windows")
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(content)
}

// replaceByRename writes a new file next to path and renames it over path, as
// editors, mv and many deployment tools do.
func replaceByRename(t *testing.T, path, content string) {
	t.Helper()
	temporary := path + ".tmp"
	writeFile(t, temporary, content)
	require.NoError(t, os.Rename(temporary, path))
}

// configMapDir lays out files the way the kubelet mounts a ConfigMap: each
// file is a symlink to ..data/<name>, and ..data is a symlink to a
// timestamped directory that is swapped on every update.
type configMapDir struct {
	dir     string
	name    string
	version int
}

// newConfigMapDir starts with the content "a: 1\n".
func newConfigMapDir(t *testing.T) *configMapDir {
	t.Helper()
	skipWithoutSymlinks(t)

	cm := &configMapDir{dir: t.TempDir(), name: "entities.yaml"}
	cm.writeVersion(t, "a: 1\n")
	require.NoError(t, os.Symlink(cm.versionDir(), filepath.Join(cm.dir, "..data")))
	require.NoError(t, os.Symlink(filepath.Join("..data", cm.name), cm.path()))

	return cm
}

func (cm *configMapDir) path() string {
	return filepath.Join(cm.dir, cm.name)
}

func (cm *configMapDir) versionDir() string {
	return fmt.Sprintf("..2026_11_26_00_00_%02d.%d", cm.version, cm.version)
}

func (cm *configMapDir) writeVersion(t *testing.T, content string) {
	t.Helper()
	cm.version++
	require.NoError(t, os.Mkdir(filepath.Join(cm.dir, cm.versionDir()), 0o755))
	writeFile(t, filepath.Join(cm.dir, cm.versionDir(), cm.name), content)
}

func (cm *configMapDir) pointDataAt(t *testing.T, target string) {
	t.Helper()
	temporary := filepath.Join(cm.dir, "..data_tmp")
	require.NoError(t, os.Symlink(target, temporary))
	require.NoError(t, os.Rename(temporary, filepath.Join(cm.dir, "..data")))
}

// update follows the kubelet's sequence: write a new directory, swap ..data
// to it, then delete the old directory.
func (cm *configMapDir) update(t *testing.T, content string) {
	t.Helper()
	previous := cm.versionDir()
	cm.writeVersion(t, content)
	cm.pointDataAt(t, cm.versionDir())
	require.NoError(t, os.RemoveAll(filepath.Join(cm.dir, previous)))
}

func TestWatchFileWriteReloadsConfigMapUpdate(t *testing.T) {
	cm := newConfigMapDir(t)
	recorder := watchFileForTest(t, cm.path())

	cm.update(t, "a: 2\n")

	recorder.expectOneCall(t)
	assert.Equal(t, "a: 2\n", readFile(t, cm.path()))
}

func TestWatchFileWriteReloadsEveryConfigMapUpdate(t *testing.T) {
	cm := newConfigMapDir(t)
	recorder := watchFileForTest(t, cm.path())

	for version := 2; version <= 4; version++ {
		content := fmt.Sprintf("a: %d\n", version)
		cm.update(t, content)
		recorder.expectOneCall(t)
		assert.Equal(t, content, readFile(t, cm.path()))
	}
}

func TestWatchFileWriteWaitsForDanglingConfigMapLink(t *testing.T) {
	cm := newConfigMapDir(t)
	recorder := watchFileForTest(t, cm.path())

	cm.pointDataAt(t, "..missing")
	recorder.expectNone(t)

	cm.writeVersion(t, "a: 2\n")
	cm.pointDataAt(t, cm.versionDir())
	recorder.expectOneCall(t)
}

func TestWatchFileWriteWaitsForDanglingReplacement(t *testing.T) {
	skipWithoutSymlinks(t)

	dir := t.TempDir()
	path := filepath.Join(dir, "entities.yaml")
	writeFile(t, path, "a: 1\n")
	recorder := watchFileForTest(t, path)

	temporary := filepath.Join(dir, "entities.yaml.tmp")
	require.NoError(t, os.Symlink(filepath.Join(dir, "missing.yaml"), temporary))
	require.NoError(t, os.Rename(temporary, path))
	recorder.expectNone(t)

	writeFile(t, filepath.Join(dir, "missing.yaml"), "a: 2\n")
	recorder.expectOneCall(t)
}

func TestWatchFileWriteReloadsRenameSave(t *testing.T) {
	path := filepath.Join(t.TempDir(), "entities.yaml")
	writeFile(t, path, "a: 1\n")
	recorder := watchFileForTest(t, path)

	replaceByRename(t, path, "a: 2\n")

	recorder.expectOneCall(t)
}

func TestWatchFileWriteReloadsRecreatedFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "entities.yaml")
	writeFile(t, path, "a: 1\n")
	recorder := watchFileForTest(t, path)

	require.NoError(t, os.Remove(path))
	recorder.expectNone(t)

	replaceByRename(t, path, "a: 2\n")
	recorder.expectOneCall(t)
}

func TestWatchFileWriteReloadsInPlaceWrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "entities.yaml")
	writeFile(t, path, "a: 1\n")
	recorder := watchFileForTest(t, path)

	writeFile(t, path, "a: 2\n")

	recorder.expectOneCall(t)
}

func TestWatchFileWriteReloadsNewFileOnce(t *testing.T) {
	path := filepath.Join(t.TempDir(), "entities.yaml")
	recorder := watchFileForTest(t, path)

	writeFile(t, path, "a: 1\n")

	recorder.expectOneCall(t)
}

func TestWatchFileWriteFollowsSymlinkToAnotherDirectory(t *testing.T) {
	skipWithoutSymlinks(t)

	linkDir, firstDir, secondDir := t.TempDir(), t.TempDir(), t.TempDir()
	firstTarget := filepath.Join(firstDir, "real.yaml")
	secondTarget := filepath.Join(secondDir, "real.yaml")
	writeFile(t, firstTarget, "a: 1\n")
	writeFile(t, secondTarget, "b: 1\n")

	path := filepath.Join(linkDir, "entities.yaml")
	require.NoError(t, os.Symlink(firstTarget, path))
	recorder := watchFileForTest(t, path)

	writeFile(t, firstTarget, "a: 2\n")
	recorder.expectOneCall(t)

	temporary := filepath.Join(linkDir, "entities.yaml.tmp")
	require.NoError(t, os.Symlink(secondTarget, temporary))
	require.NoError(t, os.Rename(temporary, path))
	recorder.expectOneCall(t)

	writeFile(t, secondTarget, "b: 2\n")
	recorder.expectOneCall(t)

	writeFile(t, firstTarget, "a: 3\n")
	recorder.expectNone(t)
}

func TestWatchFileWriteIgnoresUnrelatedChanges(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "entities.yaml")
	writeFile(t, path, "a: 1\n")
	recorder := watchFileForTest(t, path)

	other := filepath.Join(dir, "other.yaml")
	writeFile(t, other, "b: 1\n")
	writeFile(t, other, "b: 2\n")
	replaceByRename(t, other, "b: 3\n")
	require.NoError(t, os.Remove(other))
	require.NoError(t, os.Chmod(path, 0o640))

	recorder.expectNone(t)
}

func TestWatchFileWriteKeepsSameNamedFilesApart(t *testing.T) {
	first := filepath.Join(t.TempDir(), "entities.yaml")
	second := filepath.Join(t.TempDir(), "entities.yaml")
	writeFile(t, first, "a: 1\n")
	writeFile(t, second, "b: 1\n")
	firstRecorder := watchFileForTest(t, first)
	secondRecorder := watchFileForTest(t, second)

	writeFile(t, first, "a: 2\n")
	writeFile(t, second, "b: 2\n")

	firstRecorder.expectOneCall(t)
	secondRecorder.expectOneCall(t)
}

func TestWatchFileWriteStopsAfterStopFileWatch(t *testing.T) {
	cm := newConfigMapDir(t)
	recorder := watchFileForTest(t, cm.path())

	StopFileWatch(cm.path())
	cm.update(t, "a: 2\n")

	recorder.expectNone(t)
}

func TestWatchDirectoryCreateStillFiresForNewFiles(t *testing.T) {
	debounceWriteLogMutex.Lock()
	debounceWriteLog = make(map[string]*FsNotifyLogEntry)
	debounceWriteLogMutex.Unlock()

	dir := t.TempDir()
	calls := make(chan string, 16)
	done := make(chan struct{})
	t.Cleanup(func() { close(done) })

	// The watch is added before the event loop starts, so the file created
	// below cannot be missed.
	watcher, err := fsnotify.NewWatcher()
	require.NoError(t, err)
	require.NoError(t, watcher.Add(dir))
	go runWatcher(&watchContext{
		filedir:         dir,
		callback:        func(filename string) { calls <- filename },
		interestedEvent: fsnotify.Create,
	}, watcher, done)

	created := filepath.Join(dir, "new.txt")
	writeFile(t, created, "x")

	select {
	case name := <-calls:
		assert.Equal(t, created, name)
	case <-time.After(callbackTimeout):
		t.Fatal("directory watch did not fire for a created file")
	}
}
