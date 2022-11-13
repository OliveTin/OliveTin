package httpservers

import (
	"encoding/json"
	//	cors "github.com/OliveTin/OliveTin/internal/cors"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"

	config "github.com/OliveTin/OliveTin/internal/config"
	updatecheck "github.com/OliveTin/OliveTin/internal/updatecheck"
)

type webUISettings struct {
	Rest             string
	ThemeName        string
	ShowFooter       bool
	ShowNavigation   bool
	ShowNewVersions  bool
	AvailableVersion string
	CurrentVersion   string
	PageTitle        string
}

func findWebuiDir() string {
	directoriesToSearch := []string{
		"./webui",
		"/usr/share/OliveTin/webui/",
		"/var/www/OliveTin/",
		"/var/www/olivetin/",
		"/etc/OliveTin/webui/",
	}

	for _, dir := range directoriesToSearch {
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

func generateWebUISettings(w http.ResponseWriter, r *http.Request) {
	jsonRet, _ := json.Marshal(webUISettings{
		Rest:             cfg.ExternalRestAddress + "/api/",
		ThemeName:        cfg.ThemeName,
		ShowFooter:       cfg.ShowFooter,
		ShowNavigation:   cfg.ShowNavigation,
		ShowNewVersions:  cfg.ShowNewVersions,
		AvailableVersion: updatecheck.AvailableVersion,
		CurrentVersion:   updatecheck.CurrentVersion,
		PageTitle:        cfg.PageTitle,
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

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(findWebuiDir())))
	mux.HandleFunc("/webUiSettings.json", generateWebUISettings)

	srv := &http.Server{
		Addr:    cfg.ListenAddressWebUI,
		Handler: mux,
	}

	log.Fatal(srv.ListenAndServe())
}
