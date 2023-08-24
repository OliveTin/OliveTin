package httpservers

import (
	"context"
	log "github.com/sirupsen/logrus"
	"net/http"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
	"time"
)

func handleWebsocket(w http.ResponseWriter, r *http.Request) bool {
	c, err := websocket.Accept(w, r, nil)

	if err != nil {
		log.Warnf("Websocket issue: %v", err)
		return false
	}

	defer c.Close(websocket.StatusInternalError, "Goodbye")

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*10)

	defer cancel()

	var v interface{}

	err = wsjson.Read(ctx, c, v)

	log.Printf("recv: %v", v)

	if err != nil {
		log.Warnf("Websocket issue: %v", err)
		return false
	}

	c.Close(websocket.StatusNormalClosure, "")
	return true
}
