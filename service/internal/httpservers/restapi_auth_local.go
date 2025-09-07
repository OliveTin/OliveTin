package httpservers

import (
	"github.com/OliveTin/OliveTin/internal/config"
	"google.golang.org/grpc/metadata"
	"net/http"

	"github.com/google/uuid"
)

func parseLocalUserCookie(cfg *config.Config, req *http.Request) (string, string, string) {
	cookie, err := req.Cookie("olivetin-sid-local")

	if err != nil {
		return "", "", ""
	}

	cookieValue := cookie.Value

	user := getUserFromSession(cfg, "local", cookieValue)

	if user == nil {
		return "", "", ""
	}

	return user.Username, user.Usergroup, cookie.Value
}

func forwardResponseHandlerLoginLocalUser(cfg *config.Config, md metadata.MD, w http.ResponseWriter) error {
	setUsername := getMetadataKeyOrEmpty(md, "set-username")

	if setUsername != "" {
		user := cfg.FindUserByUsername(setUsername)

		if user == nil {
			return nil
		}

		sid := uuid.NewString()
		registerUserSession(cfg, "local", sid, user.Username)

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
