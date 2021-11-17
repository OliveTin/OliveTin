package httpservers

import (
	"encoding/json"
	//	cors "github.com/jamesread/OliveTin/internal/cors"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"

	config "github.com/jamesread/OliveTin/internal/config"
	updatecheck "github.com/jamesread/OliveTin/internal/updatecheck"
)

type webUISettings struct {
	Rest             string
	ThemeName        string
	HideNavigation   bool
	AvailableVersion string
	CurrentVersion   string
	ShowNewVersions  bool
}

func findWebuiDir() string {
	directoriesToSearch := []string{
		"./webui",
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
	restAddress := ""

	if !cfg.UseSingleHTTPFrontend {
		restAddress = cfg.ExternalRestAddress
	}

	jsonRet, _ := json.Marshal(webUISettings{
		Rest:             restAddress + "/api/",
		ThemeName:        cfg.ThemeName,
		HideNavigation:   cfg.HideNavigation,
		AvailableVersion: updatecheck.AvailableVersion,
		CurrentVersion:   updatecheck.CurrentVersion,
		ShowNewVersions:  cfg.ShowNewVersions,
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
