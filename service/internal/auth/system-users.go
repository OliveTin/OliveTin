package auth

import (
	authpublic "github.com/OliveTin/OliveTin/internal/auth/authpublic"
	config "github.com/OliveTin/OliveTin/internal/config"
)

func UserGuest(cfg *config.Config) *authpublic.AuthenticatedUser {
	ret := &authpublic.AuthenticatedUser{}
	ret.Username = "guest"
	ret.UsergroupLine = "guest"
	ret.Provider = "system"

	ret.BuildUserAcls(cfg)

	return ret
}

func UserFromSystem(cfg *config.Config, username string) *authpublic.AuthenticatedUser {
	ret := &authpublic.AuthenticatedUser{
		Username:      username,
		UsergroupLine: "system",
		Provider:      "system",
	}

	ret.BuildUserAcls(cfg)

	return ret
}
