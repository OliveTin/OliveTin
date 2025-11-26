package auth

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	types "github.com/OliveTin/OliveTin/internal/auth/authpublic"
	otjwt "github.com/OliveTin/OliveTin/internal/auth/otjwt"
	"github.com/OliveTin/OliveTin/internal/config"
	log "github.com/sirupsen/logrus"
)

var authChain = []func(*types.AuthCheckingContext) *types.AuthenticatedUser{
	checkUserFromHeaders,
	checkUserFromLocalSession,
	otjwt.CheckUserFromJwtHeader,
	otjwt.CheckUserFromJwtCookie,
}

// Handlers like the OAuth2's handler are "instance methods", so they need to be added to the auth chain after the other handlers.
func AddAuthChainFunction(check func(*types.AuthCheckingContext) *types.AuthenticatedUser) {
	authChain = append(authChain, check)
}

func runAuthChain[T any](req *connect.Request[T], cfg *config.Config) *types.AuthenticatedUser {
	var user *types.AuthenticatedUser

	authCtx := &types.AuthCheckingContext{
		Request: &http.Request{Header: req.Header()},
		Config:  cfg,
	}

	for _, check := range authChain {
		user = check(authCtx)

		if user != nil && user.Username != "" {
			return user
		}
	}

	return nil
}

func UserFromApiCall[T any](ctx context.Context, req *connect.Request[T], cfg *config.Config) *types.AuthenticatedUser {
	user := runAuthChain(req, cfg)

	log.Tracef("UserFromApiCall Context: %+v", ctx)

	if user == nil || user.Username == "" {
		user = UserGuest(cfg)
	} else {
		user.BuildUserAcls(cfg)
	}

	path := ""
	if req != nil {
		path = req.Spec().Procedure
	}

	log.WithFields(log.Fields{
		"username":      user.Username,
		"usergroupLine": user.UsergroupLine,
		"provider":      user.Provider,
		"acls":          user.Acls,
		"path":          path,
	}).Debugf("Authenticated API request")

	return user
}
