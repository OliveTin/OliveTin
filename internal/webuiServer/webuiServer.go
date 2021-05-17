package webuiServer

import (
	"encoding/json"
	cors "github.com/jamesread/OliveTin/internal/cors"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
)

type WebUISettings struct {
	Rest string
}

func findWebuiDir() string {
	directoriesToSearch := []string{
		"./webui",
		"/var/www/olivetin/",
	}

	for _, dir := range directoriesToSearch {
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			log.Infof("Found the webui directory here: %v", dir)

			return dir
		}
	}

	log.Warnf("Did not find the webui directory, you will probably get 404 errors.")

	return "./webui" // Should not exist
}

func Start(listenAddress string, listenAddressRest string) {
	http.Handle("/", cors.AllowCors(http.FileServer(http.Dir(findWebuiDir()))))

	http.HandleFunc("/webUiSettings.json", func(w http.ResponseWriter, r *http.Request) {
		ret := WebUISettings{
			Rest: "http://" + listenAddressRest + "/",
		}

		jsonRet, _ := json.Marshal(ret)

		w.Write([]byte(jsonRet))
	})

	log.Fatal(http.ListenAndServe(listenAddress, nil))
}
