package filehelper_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/OliveTin/OliveTin/internal/configissues"
	"github.com/OliveTin/OliveTin/internal/filehelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWatchDirectoryCreateReportsMissingPath(t *testing.T) {
	configissues.Replace(nil)

	missing := filepath.Join(t.TempDir(), "missing-watch-dir")
	done := make(chan struct{})
	go func() {
		filehelper.WatchDirectoryCreate(missing, func(string) {}, filehelper.WatchMeta{})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("WatchDirectoryCreate did not return after watcher failure")
	}

	issues := configissues.List()
	require.Len(t, issues, 1)
	assert.Equal(t, configissues.CodeWatcherPath, issues[0].Code)
	assert.Equal(t, missing, issues[0].Source)
	assert.Contains(t, issues[0].Message, missing)
	assert.Empty(t, issues[0].ActionID)
}

func TestWatchDirectoryCreateReportsActionMeta(t *testing.T) {
	configissues.Replace(nil)

	missing := filepath.Join(t.TempDir(), "missing-secret-watch")
	done := make(chan struct{})
	go func() {
		filehelper.WatchDirectoryCreate(missing, func(string) {}, filehelper.WatchMeta{
			ActionID:    "secret_action",
			ActionTitle: "Secret Action",
			ConfigFile:  "/etc/OliveTin/secret.yaml",
		})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("WatchDirectoryCreate did not return after watcher failure")
	}

	issues := configissues.List()
	require.Len(t, issues, 1)
	assert.Equal(t, "secret_action", issues[0].ActionID)
	assert.Equal(t, "Secret Action", issues[0].ActionTitle)
	assert.Equal(t, "/etc/OliveTin/secret.yaml", issues[0].ConfigFile)
}
