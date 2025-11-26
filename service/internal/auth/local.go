package auth

import (
	"net/http"

	types "github.com/OliveTin/OliveTin/internal/auth/authpublic"
	log "github.com/sirupsen/logrus"
)

func getLocalSessionCookie(r *http.Request) (string, bool) {
	c, err := r.Cookie("olivetin-sid-local")
	if err != nil {
		return "", false
	}
	if c == nil {
		return "", false
	}
	if c.Value == "" {
		return "", false
	}
	return c.Value, true
}

func checkUserFromLocalSession(context *types.AuthCheckingContext) *types.AuthenticatedUser {
	u := &types.AuthenticatedUser{}

	sid, ok := getLocalSessionCookie(context.Request)
	if !ok {
		return u
	}

	sess := GetUserSession("local", sid)
	if sess == nil {
		log.WithFields(log.Fields{"sid": sid, "provider": "local"}).Warn("UserFromContext: stale local session")
		return u
	}

	cfgUser := context.Config.FindUserByUsername(sess.Username)
	if cfgUser == nil {
		log.WithFields(log.Fields{"username": sess.Username}).Warn("UserFromContext: local session user not in config")
		return u
	}

	u.Username = cfgUser.Username
	u.UsergroupLine = cfgUser.Usergroup
	u.Provider = "local"
	u.SID = sid
	return u
}
