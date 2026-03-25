package fileupload

import (
	"os"
	"strings"
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

func TestSanitizeUploadFilenameShellSafe(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   string
		want string
	}{
		{"normal.txt", "normal.txt"},
		{"My-Document_2.pdf", "My-Document_2.pdf"},
		{"", "upload"},
		{".", "upload"},
		{"../../../etc/passwd", "passwd"},
		{"foo;rm -rf /", "foo_rm_-rf_"},
		{"a$b`x$(y)", "a_b_x__y_"},
		{"x\ny\tz", "x_y_z"},
		{"a|b&c>d<e", "a_b_c_d_e"},
		{`a\b`, "a_b"},
		{`'quote"`, "_quote_"},
		{strings.Repeat("n", 300), strings.Repeat("n", 255)},
	}
	for _, tc := range cases {
		got := SanitizeUploadFilename(tc.in)
		if got != tc.want {
			t.Errorf("SanitizeUploadFilename(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
