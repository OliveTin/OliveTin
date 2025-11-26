package auth

import (
	"net/http"

	types "github.com/OliveTin/OliveTin/internal/auth/authpublic"
)

//gocyclo:ignore
func checkUserFromHeaders(context *types.AuthCheckingContext) *types.AuthenticatedUser {
	u := &types.AuthenticatedUser{}

	if context.Config.AuthHttpHeaderUsername != "" {
		u.Username = getHeaderKeyOrEmpty(context.Request.Header, context.Config.AuthHttpHeaderUsername)
	}

	if context.Config.AuthHttpHeaderUserGroup != "" {
		u.UsergroupLine = getHeaderKeyOrEmpty(context.Request.Header, context.Config.AuthHttpHeaderUserGroup)
	}

	if prov := getHeaderKeyOrEmpty(context.Request.Header, "provider"); prov != "" {
		u.Provider = prov
	}
	return u
}

func getHeaderKeyOrEmpty(headers http.Header, key string) string {
	values := headers.Values(key)
	if len(values) > 0 {
		return values[0]
	}
	return ""
}
