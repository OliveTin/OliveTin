package httpservers

import (
	"encoding/json"
	//	cors "github.com/OliveTin/OliveTin/internal/cors"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"path"

	"github.com/jamesread/golure/pkg/dirs"

	config "github.com/OliveTin/OliveTin/internal/config"
	installationinfo "github.com/OliveTin/OliveTin/internal/installationinfo"
	sv "github.com/OliveTin/OliveTin/internal/stringvariables"
)

type webUIServer struct {
	cfg *config.Config

	webuiDir string
}

var (
	customThemeCss     []byte
	customThemeCssRead = false
)

type webUISettings struct {
	Rest                   string
	ShowFooter             bool
	ShowNavigation         bool
	ShowNewVersions        bool
	AvailableVersion       string
	CurrentVersion         string
	PageTitle              string
	SectionNavigationStyle string
	DefaultIconForBack     string
	EnableCustomJs         bool
	AuthLoginUrl           string
	AuthLocalLogin         bool
	StyleMods              []string
	AuthOAuth2Providers    []publicOAuth2Provider
	AdditionalLinks        []*config.NavigationLink
}

func NewWebUIServer(cfg *config.Config) *webUIServer {
	s := &webUIServer{
		cfg: cfg,
	}

	s.webuiDir = s.findWebuiDir()
	s.setupCustomWebuiDir()

	return s
}

func (s *webUIServer) handleWebui(w http.ResponseWriter, r *http.Request) {
	//dirName := path.Dir(r.URL.Path)

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
		sv.Set("internal.themesdir", err.Error())
	} else {
		sv.Set("internal.themesdir", dir)
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

type publicOAuth2Provider struct {
	Name  string
	Title string
	Icon  string
}

func buildPublicOAuth2ProvidersList(cfg *config.Config) []publicOAuth2Provider {
	var publicProviders []publicOAuth2Provider

	for _, provider := range cfg.AuthOAuth2Providers {
		publicProviders = append(publicProviders, publicOAuth2Provider{
			Name:  provider.Name,
			Title: provider.Title,
			Icon:  provider.Icon,
		})
	}

	return publicProviders
}

func (s *webUIServer) generateWebUISettings(w http.ResponseWriter, r *http.Request) {
	log.Infof("Generating webui settings for %s", r.RemoteAddr)

	jsonRet, _ := json.Marshal(webUISettings{
		Rest:                   s.cfg.ExternalRestAddress + "/api/",
		ShowFooter:             s.cfg.ShowFooter,
		ShowNavigation:         s.cfg.ShowNavigation,
		ShowNewVersions:        s.cfg.ShowNewVersions,
		AvailableVersion:       installationinfo.Runtime.AvailableVersion,
		CurrentVersion:         installationinfo.Build.Version,
		PageTitle:              s.cfg.PageTitle,
		SectionNavigationStyle: s.cfg.SectionNavigationStyle,
		DefaultIconForBack:     s.cfg.DefaultIconForBack,
		EnableCustomJs:         s.cfg.EnableCustomJs,
		AuthLoginUrl:           s.cfg.AuthLoginUrl,
		AuthLocalLogin:         s.cfg.AuthLocalUsers.Enabled,
		AuthOAuth2Providers:    buildPublicOAuth2ProvidersList(s.cfg),
		AdditionalLinks:        s.cfg.AdditionalNavigationLinks,
		StyleMods:              s.cfg.StyleMods,
	})

	w.Header().Add("Content-Type", "application/json")
	_, err := w.Write([]byte(jsonRet))

	if err != nil {
		log.Warnf("Could not write webui settings: %v", err)
	}
}

func (s *webUIServer) handleCustomWebui() (http.Handler) {
	return http.StripPrefix("/custom-webui/", http.FileServer(http.Dir(s.findCustomWebuiDir())))
}
