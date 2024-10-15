package httpservers

import (
	config "github.com/OliveTin/OliveTin/internal/config"
	"golang.org/x/oauth2/endpoints"
)

var oauth2ProviderDatabase = map[string]config.OAuth2Provider{
	"github": {
		Icon:          "github",
		WhoamiUrl:     "https://api.github.com/user",
		TokenUrl:      endpoints.GitHub.TokenURL,
		AuthUrl:       endpoints.GitHub.AuthURL,
		Scopes:        []string{"profile", "email"},
		UsernameField: "login",
	},
	"google": {
		Icon:      "google",
		WhoamiUrl: "https://www.googleapis.com/oauth2/v3/userinfo",
		TokenUrl:  endpoints.Google.TokenURL,
		AuthUrl:   endpoints.Google.AuthURL,
		Scopes:    []string{"profile", "email"},
	},
}
