package httpservers

import (
	"net/http"
	"os"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/jamesread/golure/pkg/dirs"

	config "github.com/OliveTin/OliveTin/internal/config"
)

type webUIServer struct {
	cfg *config.Config

	webuiDir string
}

var (
	customThemeCss     []byte
	customThemeCssRead = false
)

func NewWebUIServer(cfg *config.Config) *webUIServer {
	s := &webUIServer{
		cfg: cfg,
	}

	s.webuiDir = s.findWebuiDir()
	s.setupCustomWebuiDir()

	return s
}

// Web UI files keep their names between releases, and some replacement UIs
// (webUIDir) do not hash their asset names at all. Without Cache-Control,
// browsers cache responses that only carry Last-Modified heuristically, and
// can keep an old UI after an upgrade. no-cache makes them revalidate every
// time; unchanged files still come back as a cheap 304 (see webui_etag.go).
const webuiCacheControl = "no-cache"

func (s *webUIServer) handleWebui(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", webuiCacheControl)

	// http.Dir rejects path traversal; do not Join raw URL paths into ServeFile.
	root := http.Dir(s.webuiDir)

	log.Tracef("Serving webui from %s for %s", s.webuiDir, r.URL.Path)
	serveFileWithContentETag(w, r, root, webuiFileName(r.URL.Path), http.FileServer(root))
}

// Mangle requests for any path like /logs or /config to load the webui index.html
func webuiFileName(urlPath string) string {
	if path.Ext(urlPath) == "" {
		return "/index.html"
	}

	return urlPath
}

func (s *webUIServer) findWebuiDir() string {
	directoriesToSearch := []string{
		s.cfg.WebUIDir,
		"../frontend/dist/",
		"../webui/",
		"../frontend/",
		"/usr/share/OliveTin/frontend/",
		"/usr/share/OliveTin/webui/",
		"/var/www/OliveTin/",
		"/var/www/olivetin/",
		"/etc/OliveTin/frontend/",
	}

	dir, err := dirs.GetFirstExistingDirectory("webui", directoriesToSearch)

	if err != nil {
		log.Warnf("Did not find the webui directory, you will probably get 404 errors.")

		return "./webui" // Should not exist
	}

	log.Infof("Using webui directory: %s", dir)

	return dir
}

func (s *webUIServer) findCustomWebuiDir() string {
	dir := path.Join(s.cfg.GetDir(), "custom-webui")

	return dir
}

func (s *webUIServer) setupCustomWebuiDir() {
	dir := s.findCustomWebuiDir()

	err := os.MkdirAll(path.Join(dir, "themes/"), 0o750)

	if err != nil {
		log.Warnf("Could not create themes directory: %v", err)
	}
}

func shouldReloadThemeCss() bool {
	return !customThemeCssRead
}

func loadThemeCssFromFile(filename string) []byte {
	_, err := os.Stat(filename)
	if err == nil {
		css, err := os.ReadFile(filename)

		if err != nil {
			log.Tracef("Theme CSS file not read: %s", filename)
			return nil
		}

		return css
	}

	log.Tracef("Theme CSS file not found: %s", filename)
	return nil
}

func (s *webUIServer) generateThemeCss(w http.ResponseWriter, r *http.Request) {
	themeCssFilename := path.Join(s.findCustomWebuiDir(), "themes", s.cfg.ThemeName, "theme.css")

	if shouldReloadThemeCss() || s.cfg.ThemeCacheDisabled {
		customThemeCssRead = true
		customThemeCss = loadThemeCssFromFile(themeCssFilename)
	}

	w.Header().Add("Content-Type", "text/css")
	w.Header().Set("Cache-Control", webuiCacheControl)
	_, err := w.Write(customThemeCss)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Warnf("Failed to write theme CSS")
	}
}

func (s *webUIServer) handleCustomWebui() http.Handler {
	root := http.Dir(s.findCustomWebuiDir())
	fileServer := http.StripPrefix("/custom-webui/", http.FileServer(root))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", webuiCacheControl)
		name := "/" + strings.TrimPrefix(r.URL.Path, "/custom-webui/")
		serveFileWithContentETag(w, r, root, name, fileServer)
	})
}
