package staticWebserverForUi

import (
	"net/http"
	"encoding/json"
	log "github.com/sirupsen/logrus"
)

type WebUiSettings struct {
	Rest string
}

func Start(listenAddress string, listenAddressRest string) {
	http.Handle("/", http.FileServer(http.Dir("webui")))

	http.HandleFunc("/webUiSettings.json", func(w http.ResponseWriter, r *http.Request) {
		ret := WebUiSettings {
			Rest: "http://" + listenAddressRest + "/",
		}

		jsonRet, _ := json.Marshal(ret)

		w.Write([]byte(jsonRet));
	})

	log.Fatal(http.ListenAndServe(listenAddress, nil))
}
