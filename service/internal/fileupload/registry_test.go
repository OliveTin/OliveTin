package fileupload

import (
	"os"
	"testing"
	"time"

	config "github.com/OliveTin/OliveTin/internal/config"
)

func TestPruneExpiredRemovesStalePendingFile(t *testing.T) {
	r := mustRegistry(t)
	path := mustTempUnder(t, r.baseDir)
	seedExpiredPending(r, path)
	r.pruneExpired()
	assertNoFile(t, path)
	assertPendingGone(t, r)
}

func mustRegistry(t *testing.T) *Registry {
	t.Helper()
	r, err := NewRegistry(config.DefaultConfig())
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	return r
}

func mustTempUnder(t *testing.T, dir string) string {
	t.Helper()
	f, err := os.CreateTemp(dir, "olu-")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	path := f.Name()
	_ = f.Close()
	return path
}

func seedExpiredPending(r *Registry, path string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.pending["testtoken"] = &pendingEntry{
		path:    path,
		expires: time.Now().Add(-time.Minute),
	}
}

func assertNoFile(t *testing.T, path string) {
	t.Helper()
	_, err := os.Stat(path)
	if !os.IsNotExist(err) {
		t.Fatalf("expected temp file removed, stat err=%v", err)
	}
}

func assertPendingGone(t *testing.T, r *Registry) {
	t.Helper()
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.pending["testtoken"]; ok {
		t.Fatal("expected pending entry deleted")
	}
}
