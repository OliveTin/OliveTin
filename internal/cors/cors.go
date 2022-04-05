package cors

import (
	log "github.com/sirupsen/logrus"
	"net/http"
)

// AllowCors takes a HTTP handler and adds Access-Control-Allow-Origin headers to
// responses.
//
// Note: HTTP OPTIONS requests (which need to be preflighted" for CORS) are not
// handled because this app does not use HTTP PUT/PATCH/etc.
func AllowCors(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			log.Debugf("Adding CORS header origin: %q", origin)

			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		h.ServeHTTP(w, r)
	})
}
