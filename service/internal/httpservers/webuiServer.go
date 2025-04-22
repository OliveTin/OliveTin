package httpservers

import (
	"encoding/json"
	"fmt"
	"strings"
	//	cors "github.com/OliveTin/OliveTin/internal/cors"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"path"
	"path/filepath"

	config "github.com/OliveTin/OliveTin/internal/config"
	installationinfo "github.com/OliveTin/OliveTin/internal/installationinfo"
	sv "github.com/OliveTin/OliveTin/internal/stringvariables"
)

var (
	customThemeCss     []byte
	customThemeCssRead = false
)

type webUISettings struct {
	BaseURL                string
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
	AuthOAuth2Providers    []publicOAuth2Provider
	AdditionalLinks        []*config.NavigationLink
}

func findWebuiDir() string {
	directoriesToSearch := []string{
		cfg.WebUIDir,
		"../webui/",
		"/usr/share/OliveTin/webui/",
		"/var/www/OliveTin/",
		"/var/www/olivetin/",
		"/etc/OliveTin/webui/",
	}

	// Use a classic i := 0 style for loop here instead of range, as the
	// search order must be deterministic - the order that the slice was defined in.
	for i := 0; i < len(directoriesToSearch); i++ {
		dir := directoriesToSearch[i]
		absdir, _ := filepath.Abs(dir)

		if _, err := os.Stat(absdir); !os.IsNotExist(err) {
			log.WithFields(log.Fields{
				"dir": absdir,
			}).Infof("Found the webui directory")

			sv.Set("internal.webuidir", absdir+" ("+dir+")")

			return dir
		}
	}

	log.Warnf("Did not find the webui directory, you will probably get 404 errors.")

	return "./webui" // Should not exist
}

func findCustomWebuiDir() string {
	dir := path.Join(cfg.GetDir(), "custom-webui")

	return dir
}

func setupCustomWebuiDir() {
	dir := findCustomWebuiDir()

	err := os.MkdirAll(path.Join(dir, "themes/"), 0775)

	if err != nil {
		log.Warnf("Could not create themes directory: %v", err)
		sv.Set("internal.themesdir", err.Error())
	} else {
		sv.Set("internal.themesdir", dir)
	}
}

func generateThemeCss(w http.ResponseWriter, r *http.Request) {
	themeCssFilename := path.Join(findCustomWebuiDir(), "themes", cfg.ThemeName, "theme.css")

	if !customThemeCssRead || cfg.ThemeCacheDisabled {
		customThemeCssRead = true

		if _, err := os.Stat(themeCssFilename); err == nil {
			customThemeCss, err = os.ReadFile(themeCssFilename)
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

func generateWebUISettings(w http.ResponseWriter, r *http.Request) {
	baseURL := baseURL()
	restURL := strings.TrimSuffix(baseURL, "/") + "/api/"
	jsonRet, _ := json.Marshal(webUISettings{
		BaseURL:                baseURL,
		Rest:                   restURL,
		ShowFooter:             cfg.ShowFooter,
		ShowNavigation:         cfg.ShowNavigation,
		ShowNewVersions:        cfg.ShowNewVersions,
		AvailableVersion:       installationinfo.Runtime.AvailableVersion,
		CurrentVersion:         installationinfo.Build.Version,
		PageTitle:              cfg.PageTitle,
		SectionNavigationStyle: cfg.SectionNavigationStyle,
		DefaultIconForBack:     cfg.DefaultIconForBack,
		EnableCustomJs:         cfg.EnableCustomJs,
		AuthLoginUrl:           cfg.AuthLoginUrl,
		AuthLocalLogin:         cfg.AuthLocalUsers.Enabled,
		AuthOAuth2Providers:    buildPublicOAuth2ProvidersList(cfg),
		AdditionalLinks:        cfg.AdditionalNavigationLinks,
	})

	w.Header().Add("Content-Type", "application/json")
	_, err := w.Write([]byte(jsonRet))

	if err != nil {
		log.Warnf("Could not write webui settings: %v", err)
	}
}

func startWebUIServer(cfg *config.Config) {
	log.WithFields(log.Fields{
		"address": cfg.ListenAddressWebUI,
	}).Info("Starting WebUI server")

	setupCustomWebuiDir()

	baseUrl := baseURL()
	mux := http.NewServeMux()
	mux.Handle(baseURLPath("/custom-webui/"), http.StripPrefix(baseURLPath("/custom-webui/"), http.FileServer(http.Dir(findCustomWebuiDir()))))
	mux.HandleFunc(baseURLPath("/theme.css"), generateThemeCss)
	mux.HandleFunc(baseURLPath("/webUiSettings.json"), generateWebUISettings)

	webuiDir := findWebuiDir()
	mux.HandleFunc(baseUrl, func(w http.ResponseWriter, r *http.Request) {
		// Mangle requests for any path like /logs or /config to load the webui index.html
		if path.Ext(r.URL.Path) == "" || strings.HasSuffix(r.URL.Path, "/") {
			log.Debugf("Mangling request for %s to /index.html", r.URL.Path)
			serveIndexHtmlWithBasePath(w)
		} else {
			http.StripPrefix(baseURLPath("/"), http.FileServer(http.Dir(webuiDir))).ServeHTTP(w, r)
		}
	})

	srv := &http.Server{
		Addr:    cfg.ListenAddressWebUI,
		Handler: mux,
	}

	log.Fatal(srv.ListenAndServe())
}

// serveIndexHtmlWithBasePath allows us to set the base href of the index.html when it is served
func serveIndexHtmlWithBasePath(w http.ResponseWriter) {
	webuiDir := findWebuiDir()
	// Read the file into memory
	data, err := os.ReadFile(path.Join(webuiDir, "index.html"))
	if err != nil {
		http.Error(w, "File not found", 404)
		return
	}
	baseUrl := baseURL()
	url := fmt.Sprintf("<base href=\"%s\">", baseUrl)

	// Replace the base href
	content := strings.ReplaceAll(string(data), "<base href=\"/\">", url)

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(content))
}

// baseURLPath returns the url including the baseUrl
func baseURLPath(path string) string {
	baseURL := strings.TrimSuffix(baseURL(), "/")
	path = strings.TrimPrefix(path, "/")
	return baseURL + "/" + path
}

// baseURL returns the baseUrl for the application based on the externalRestAddress and the subpath (if configured)
// It adds a trailing /
//
//gocyclo:ignore
func baseURL() string {
	baseURL := strings.TrimSuffix(cfg.ExternalRestAddress, "/")

	// Default externalRestAddress and Subpath
	if baseURL == "." && cfg.Subpath == "" {
		return "/"
	}
	if baseURL == "." {
		baseURL = ""
	}
	if cfg.Subpath != "" {
		baseURL += normaliseSubpath(cfg.Subpath)
		if !strings.HasSuffix(baseURL, "/") {
			baseURL += "/"
		}
	}
	return baseURL
}

// normaliseSubpath returns the base url with extra slashes removed
func normaliseSubpath(subpath string) string {
	return "/" + strings.Trim(subpath, "/")
}

// getCookiePath returns the appropriate path for cookies based on the configured subpath.
// If no subpath is configured, it returns "/", otherwise it returns the normalized subpath.
func getCookiePath() string {
	subpath := cfg.Subpath
	if subpath == "" {
		return "/"
	}
	return normaliseSubpath(subpath)
}
