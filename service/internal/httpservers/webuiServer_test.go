package httpservers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleWebuiSetsCacheControl(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "index.html"), []byte("<!doctype html>"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "app.js"), []byte("// app"), 0o600))

	s := &webUIServer{webuiDir: dir}

	for _, p := range []string{"/", "/app.js", "/logs"} {
		rec := httptest.NewRecorder()
		s.handleWebui(rec, httptest.NewRequest(http.MethodGet, p, nil))

		assert.Equal(t, http.StatusOK, rec.Code, p)
		assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"), p)
		assert.NotEmpty(t, rec.Header().Get("Last-Modified"), p)
	}
}
