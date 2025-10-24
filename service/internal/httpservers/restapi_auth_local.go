package httpservers

import (
	"net/http"

	"github.com/OliveTin/OliveTin/internal/auth"
	"github.com/OliveTin/OliveTin/internal/config"
	log "github.com/sirupsen/logrus"
)

func parseLocalUserCookie(cfg *config.Config, req *http.Request) (string, string, string) {
	cookie, err := req.Cookie("olivetin-sid-local")

	if err != nil {
		return "", "", ""
	}

	cookieValue := cookie.Value

	session := auth.GetUserSession("local", cookieValue)
	if session == nil {
		return "", "", ""
	}

	user := cfg.FindUserByUsername(session.Username)
	if user == nil {
		log.WithFields(log.Fields{
			"username": session.Username,
		}).Warnf("User not found in config")
		return "", "", ""
	}

	return user.Username, user.Usergroup, cookie.Value
}
