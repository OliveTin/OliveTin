package api

import (
	ctx "context"
	"errors"
	"net/http"

	"connectrpc.com/connect"

	apiv1connect "github.com/OliveTin/OliveTin/gen/olivetin/api/v1/apiv1connect"
	auth "github.com/OliveTin/OliveTin/internal/auth"
)

// When authRequireGuestsToLogin is enabled, unauthenticated visitors may only
// call these procedures, which are needed to log in or to check the server is
// up. Every other procedure is refused before its handler runs, so a handler
// cannot forget the check, and a lookup cannot reveal whether an ID exists.
//
// Adding a procedure here makes it public. TestEveryProcedureRequiresLoginUnlessPublic
// lists the same procedures and fails if the two lists differ.
var loginNotRequiredProcedures = map[string]bool{
	apiv1connect.OliveTinApiServiceInitProcedure:           true,
	apiv1connect.OliveTinApiServiceGetReadyzProcedure:      true,
	apiv1connect.OliveTinApiServiceLocalUserLoginProcedure: true,
	apiv1connect.OliveTinApiServiceLogoutProcedure:         true,
	apiv1connect.OliveTinApiServicePasswordHashProcedure:   true,
}

var errLoginRequired = errors.New("login is required")

type loginRequiredInterceptor struct {
	api *oliveTinAPI
}

func (interceptor *loginRequiredInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(callCtx ctx.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		if err := interceptor.checkLoginRequired(req.Spec().Procedure, req.Header()); err != nil {
			return nil, err
		}

		return next(callCtx, req)
	}
}

func (interceptor *loginRequiredInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func (interceptor *loginRequiredInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(callCtx ctx.Context, conn connect.StreamingHandlerConn) error {
		if err := interceptor.checkLoginRequired(conn.Spec().Procedure, conn.RequestHeader()); err != nil {
			return err
		}

		return next(callCtx, conn)
	}
}

func (interceptor *loginRequiredInterceptor) checkLoginRequired(procedure string, header http.Header) error {
	if loginNotRequiredProcedures[procedure] {
		return nil
	}

	cfg := interceptor.api.cfg
	user := auth.UserFromHeaders(header, procedure, cfg)

	if user.MustLogIn(cfg) {
		return connect.NewError(connect.CodePermissionDenied, errLoginRequired)
	}

	return nil
}
