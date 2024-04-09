package httpservers

import (
	"encoding/json"
	//	cors "github.com/OliveTin/OliveTin/internal/cors"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"path"

	config "github.com/OliveTin/OliveTin/internal/config"
	updatecheck "github.com/OliveTin/OliveTin/internal/updatecheck"
)

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

		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			log.WithFields(log.Fields{
				"dir": dir,
			}).Infof("Found the webui directory")

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
	}
}

func generateThemeCss(w http.ResponseWriter, r *http.Request) {
	themeCssFilename := path.Join(findCustomWebuiDir(), "themes", cfg.ThemeName, "theme.css")

	if !customThemeCssRead {
		customThemeCssRead = true

		if _, err := os.Stat(themeCssFilename); err == nil {
			customThemeCss, err = ioutil.ReadFile(themeCssFilename)
		} else {
			log.Debugf("Theme CSS not read: %v", err)
			customThemeCss = []byte("/* not found */")
		}
	}

	w.Header().Add("Content-Type", "text/css")
	w.Write(customThemeCss)
}

func generateWebUISettings(w http.ResponseWriter, r *http.Request) {
	jsonRet, _ := json.Marshal(webUISettings{
		Rest:                   cfg.ExternalRestAddress + "/api/",
		ShowFooter:             cfg.ShowFooter,
		ShowNavigation:         cfg.ShowNavigation,
		ShowNewVersions:        cfg.ShowNewVersions,
		AvailableVersion:       updatecheck.AvailableVersion,
		CurrentVersion:         updatecheck.CurrentVersion,
		PageTitle:              cfg.PageTitle,
		SectionNavigationStyle: cfg.SectionNavigationStyle,
	})

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

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(findWebuiDir())))
	mux.Handle("/custom-webui/", http.FileServer(http.Dir(findCustomWebuiDir())))
	mux.HandleFunc("/theme.css", generateThemeCss)
	mux.HandleFunc("/webUiSettings.json", generateWebUISettings)

	srv := &http.Server{
		Addr:    cfg.ListenAddressWebUI,
		Handler: mux,
	}

	log.Fatal(srv.ListenAndServe())
}
