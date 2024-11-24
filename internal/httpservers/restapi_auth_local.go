package httpservers

import (
	"google.golang.org/grpc/metadata"
	"net/http"

	"github.com/google/uuid"
	"github.com/OliveTin/OliveTin/internal/config"
	log "github.com/sirupsen/logrus"
)

var (
	localUserSessions = make(map[string]*config.LocalUser)
)

func parseLocalUserCookie(req *http.Request) (string, string, string) {
	cookie, err := req.Cookie("olivetin-sid-local")

	if err != nil {
		return "", "", ""
	}

	cookieValue := cookie.Value

	user, ok := localUserSessions[cookieValue]

	if !ok {
		log.WithFields(log.Fields{
			"sid":      cookieValue,
			"provider": "local",
		}).Warnf("Stale session")
		return "", "", ""
	}

	return user.Username, user.Usergroup, cookie.Value
}

func findUserByUsername(searchUsername string) *config.LocalUser {
	for _, user := range cfg.AuthLocalUsers.Users {
		if user.Username == searchUsername {
			return user
		}
	}

	return nil
}

func forwardResponseHandlerLoginLocalUser(md metadata.MD, w http.ResponseWriter) error {
	setUsername := getMetadataKeyOrEmpty(md, "set-username")

	if setUsername != "" {
		user := findUserByUsername(setUsername)

		if user == nil {
			return nil
		}

		sid := uuid.NewString()
		localUserSessions[sid] = user

		http.SetCookie(
			w,
			&http.Cookie{
				Name:     "olivetin-sid-local",
				Value:    sid,
				MaxAge:   31556952, // 1 year
				HttpOnly: true,
				Path:     "/",
			},
		)
	}

	return nil
}
