package httpservers

import (

	//	cors "github.com/OliveTin/OliveTin/internal/cors"
	"net/http"
	"os"
	"path"

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

func (s *webUIServer) handleWebui(w http.ResponseWriter, r *http.Request) {
	// dirName := path.Dir(r.URL.Path)

	// Mangle requests for any path like /logs or /config to load the webui index.html
	if path.Ext(r.URL.Path) == "" && r.URL.Path != "/" {
		log.Debugf("Mangling request for %s to /index.html", r.URL.Path)

		http.ServeFile(w, r, path.Join(s.webuiDir, "index.html"))
	} else {
		log.Infof("Serving webui from %s for %s", s.webuiDir, r.URL.Path)
		http.ServeFile(w, r, path.Join(s.webuiDir, r.URL.Path))
		//		http.StripPrefix(dirName, http.FileServer(http.Dir(s.webuiDir))).ServeHTTP(w, r)
	}
}

func (s *webUIServer) findWebuiDir() string {
	directoriesToSearch := []string{
		s.cfg.WebUIDir,
		"../frontend/dist/",
		"../frontend/",
		"/usr/share/OliveTin/frontend/",
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

	err := os.MkdirAll(path.Join(dir, "themes/"), 0775)

	if err != nil {
		log.Warnf("Could not create themes directory: %v", err)
	}
}

func (s *webUIServer) generateThemeCss(w http.ResponseWriter, r *http.Request) {
	themeCssFilename := path.Join(s.findCustomWebuiDir(), "themes", s.cfg.ThemeName, "theme.css")

	if !customThemeCssRead || s.cfg.ThemeCacheDisabled {
		customThemeCssRead = true

		if _, err := os.Stat(themeCssFilename); err == nil {
			customThemeCss, _ = os.ReadFile(themeCssFilename)
		} else {
			log.Debugf("Theme CSS not read: %v", err)
			customThemeCss = []byte("/* not found */")
		}
	}

	w.Header().Add("Content-Type", "text/css")
	w.Write(customThemeCss)
}

func (s *webUIServer) handleCustomWebui() http.Handler {
	return http.StripPrefix("/custom-webui/", http.FileServer(http.Dir(s.findCustomWebuiDir())))
}
