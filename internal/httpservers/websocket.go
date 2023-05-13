package httpservers

import (
	"time"
	log "github.com/sirupsen/logrus"
	"net/http"
	"nhooyr.io/websocket"
)

func handleWebsocket(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, nil)

	if err != nil {
		Log.Warnf("Websocket issue: %v", err)
		return
	}

	defer c.Close(websocket.StatusInternalError, "Goodbye")

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*10)

	defer cancel()

	var v interface{}

	err = wsjson.Read(Ctx, c, v)

	if err != nil {
		Log.Warnf("Websocket issue: %v", err)
		return
	}

	log.Printf("recv: %v", v)

	c.Close(websocket.StatusNormalClosure, "")
}
