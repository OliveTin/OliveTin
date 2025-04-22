package httpservers

import (
	"google.golang.org/grpc/metadata"
	"net/http"

	"github.com/google/uuid"
)

func parseLocalUserCookie(req *http.Request) (string, string, string) {
	cookie, err := req.Cookie("olivetin-sid-local")

	if err != nil {
		return "", "", ""
	}

	cookieValue := cookie.Value

	user := getUserFromSession("local", cookieValue)

	if user == nil {
		return "", "", ""
	}

	return user.Username, user.Usergroup, cookie.Value
}

func forwardResponseHandlerLoginLocalUser(md metadata.MD, w http.ResponseWriter) error {
	setUsername := getMetadataKeyOrEmpty(md, "set-username")

	if setUsername != "" {
		user := cfg.FindUserByUsername(setUsername)

		if user == nil {
			return nil
		}

		sid := uuid.NewString()
		registerUserSession("local", sid, user.Username)

		http.SetCookie(
			w,
			&http.Cookie{
				Name:     "olivetin-sid-local",
				Value:    sid,
				MaxAge:   31556952, // 1 year
				HttpOnly: true,
				Path:     getCookiePath(),
			},
		)
	}

	return nil
}
