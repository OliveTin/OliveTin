package httpservers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	config "github.com/OliveTin/OliveTin/internal/config"
)

func TestHandleWebuiSetsCacheControl(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "index.html"), []byte("<!doctype html>"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "app.js"), []byte("// app"), 0o600))

	s := &webUIServer{webuiDir: dir}

	for _, p := range []string{"/", "/app.js", "/logs"} {
		rec := httptest.NewRecorder()
		s.handleWebui(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, p, nil))

		assert.Equal(t, http.StatusOK, rec.Code, p)
		assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"), p)
		assert.NotEmpty(t, rec.Header().Get("ETag"), p)
		assert.Empty(t, rec.Header().Get("Last-Modified"), p)
	}
}

func getWithHeader(t *testing.T, handler http.Handler, urlPath string, header string, value string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, urlPath, nil)
	if header != "" {
		req.Header.Set(header, value)
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	return rec
}

func writeFileWithModTime(t *testing.T, filename string, content string, modTime time.Time) {
	t.Helper()

	require.NoError(t, os.WriteFile(filename, []byte(content), 0o600))
	require.NoError(t, os.Chtimes(filename, modTime, modTime))
}

func TestHandleWebuiRevalidatesByContent(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	appFile := filepath.Join(dir, "app.js")
	modTime := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	writeFileWithModTime(t, appFile, "// old app", modTime)

	s := &webUIServer{webuiDir: dir}
	handler := http.HandlerFunc(s.handleWebui)

	first := getWithHeader(t, handler, "/app.js", "", "")
	require.Equal(t, http.StatusOK, first.Code)
	oldETag := first.Header().Get("ETag")

	unchanged := getWithHeader(t, handler, "/app.js", "If-None-Match", oldETag)
	assert.Equal(t, http.StatusNotModified, unchanged.Code)

	writeFileWithModTime(t, appFile, "// new app", modTime.Add(-time.Hour))

	replaced := getWithHeader(t, handler, "/app.js", "If-None-Match", oldETag)
	assert.Equal(t, http.StatusOK, replaced.Code)
	assert.Equal(t, "// new app", replaced.Body.String())
	assert.NotEqual(t, oldETag, replaced.Header().Get("ETag"))

	legacy := getWithHeader(t, handler, "/app.js", "If-Modified-Since", modTime.Format(http.TimeFormat))
	assert.Equal(t, http.StatusOK, legacy.Code)
	assert.Equal(t, "// new app", legacy.Body.String())
}

func TestHandleWebuiMissingFile(t *testing.T) {
	t.Parallel()

	s := &webUIServer{webuiDir: t.TempDir()}

	rec := getWithHeader(t, http.HandlerFunc(s.handleWebui), "/missing.js", "", "")
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandleCustomWebuiRevalidatesByContent(t *testing.T) {
	t.Parallel()

	configDir := t.TempDir()
	iconsDir := filepath.Join(configDir, "custom-webui", "icons")
	require.NoError(t, os.MkdirAll(iconsDir, 0o750))

	iconFile := filepath.Join(iconsDir, "open.svg")
	modTime := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	writeFileWithModTime(t, iconFile, "<svg>old</svg>", modTime)

	cfg := config.DefaultConfig()
	cfg.SetDir(configDir)
	handler := (&webUIServer{cfg: cfg}).handleCustomWebui()

	first := getWithHeader(t, handler, "/custom-webui/icons/open.svg", "", "")
	require.Equal(t, http.StatusOK, first.Code)
	assert.Equal(t, "no-cache", first.Header().Get("Cache-Control"))
	assert.Empty(t, first.Header().Get("Last-Modified"))

	writeFileWithModTime(t, iconFile, "<svg>new</svg>", modTime)

	replaced := getWithHeader(t, handler, "/custom-webui/icons/open.svg", "If-None-Match", first.Header().Get("ETag"))
	assert.Equal(t, http.StatusOK, replaced.Code)
	assert.Equal(t, "<svg>new</svg>", replaced.Body.String())

	missing := getWithHeader(t, handler, "/custom-webui/icons/closed.svg", "", "")
	assert.Equal(t, http.StatusNotFound, missing.Code)
}
