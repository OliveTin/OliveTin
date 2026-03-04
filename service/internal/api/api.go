package api

import (
	ctx "context"
	"encoding/json"
	"errors"
	"os"
	"path"
	"sort"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/encoding/protojson"

	apiv1 "github.com/OliveTin/OliveTin/gen/olivetin/api/v1"
	apiv1connect "github.com/OliveTin/OliveTin/gen/olivetin/api/v1/apiv1connect"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"fmt"
	"net/http"
	"sync"
	"time"

	acl "github.com/OliveTin/OliveTin/internal/acl"
	auth "github.com/OliveTin/OliveTin/internal/auth"
	authpublic "github.com/OliveTin/OliveTin/internal/auth/authpublic"
	config "github.com/OliveTin/OliveTin/internal/config"
	entities "github.com/OliveTin/OliveTin/internal/entities"
	executor "github.com/OliveTin/OliveTin/internal/executor"
	installationinfo "github.com/OliveTin/OliveTin/internal/installationinfo"
	"github.com/OliveTin/OliveTin/internal/tpl"
	connectproto "go.akshayshah.org/connectproto"
)

type oliveTinAPI struct {
	executor *executor.Executor
	cfg      *config.Config

	// streamingClients is a set of currently connected clients.
	// The empty struct value models set semantics (keys only) and keeps add/remove O(1).
	// We use a map for efficient membership and deletion; ordering is not required.
	streamingClients      map[*streamingClient]struct{}
	streamingClientsMutex sync.RWMutex
}

// This is used to avoid race conditions when iterating over the connectedClients map.
// and holds the lock for as minimal time as possible to avoid blocking the API for too long.
func (api *oliveTinAPI) copyOfStreamingClients() []*streamingClient {
	api.streamingClientsMutex.RLock()
	defer api.streamingClientsMutex.RUnlock()
	clients := make([]*streamingClient, 0, len(api.streamingClients))
	for client := range api.streamingClients {
		clients = append(clients, client)
	}
	return clients
}

type streamingClient struct {
	channel           chan *apiv1.EventStreamResponse
	AuthenticatedUser *authpublic.AuthenticatedUser
}

func (api *oliveTinAPI) KillAction(ctx ctx.Context, req *connect.Request[apiv1.KillActionRequest]) (*connect.Response[apiv1.KillActionResponse], error) {
	ret := &apiv1.KillActionResponse{
		ExecutionTrackingId: req.Msg.ExecutionTrackingId,
	}

	var execReqLogEntry *executor.InternalLogEntry

	execReqLogEntry, ret.Found = api.executor.GetLog(req.Msg.ExecutionTrackingId)

	if !ret.Found {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("execution not found for tracking ID %s", req.Msg.ExecutionTrackingId))
	}

	if execReqLogEntry.Binding == nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("log entry has no binding for tracking ID %s", req.Msg.ExecutionTrackingId))
	}

	action := execReqLogEntry.Binding.Action

	if action == nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("action not found for tracking ID %s", req.Msg.ExecutionTrackingId))
	}

	log.Warnf("Killing execution request by tracking ID: %v", req.Msg.ExecutionTrackingId)

	user := auth.UserFromApiCall(ctx, req, api.cfg)

	api.killActionByTrackingId(user, action, execReqLogEntry, ret)

	return connect.NewResponse(ret), nil
}

func (api *oliveTinAPI) killActionByTrackingId(user *authpublic.AuthenticatedUser, action *config.Action, execReqLogEntry *executor.InternalLogEntry, ret *apiv1.KillActionResponse) {
	if !acl.IsAllowedKill(api.cfg, user, action) {
		log.Warnf("Killing execution request not possible - user not allowed to kill this action: %v", execReqLogEntry.ExecutionTrackingID)
		ret.Killed = false
		return
	}

	err := api.executor.Kill(execReqLogEntry)

	if err != nil {
		log.Warnf("Killing execution request err: %v", err)
		ret.AlreadyCompleted = true
		ret.Killed = false
	} else {
		ret.Killed = true
	}
}

func (api *oliveTinAPI) StartAction(ctx ctx.Context, req *connect.Request[apiv1.StartActionRequest]) (*connect.Response[apiv1.StartActionResponse], error) {
	args := make(map[string]string)

	for _, arg := range req.Msg.Arguments {
		args[arg.Name] = arg.Value
	}

	pair := api.executor.FindBindingByID(req.Msg.BindingId)

	if pair == nil || pair.Action == nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("action with ID %s not found", req.Msg.BindingId))
	}

	authenticatedUser := auth.UserFromApiCall(ctx, req, api.cfg)

	execReq := executor.ExecutionRequest{
		Binding:           pair,
		TrackingID:        req.Msg.UniqueTrackingId,
		Arguments:         args,
		AuthenticatedUser: authenticatedUser,
		Cfg:               api.cfg,
	}

	api.executor.ExecRequest(&execReq)

	ret := &apiv1.StartActionResponse{
		ExecutionTrackingId: execReq.TrackingID,
	}

	return connect.NewResponse(ret), nil
}

func (api *oliveTinAPI) PasswordHash(ctx ctx.Context, req *connect.Request[apiv1.PasswordHashRequest]) (*connect.Response[apiv1.PasswordHashResponse], error) {
	hash, err := createHash(req.Msg.Password)

	if err != nil {
		if errors.Is(err, ErrArgon2Busy) {
			return nil, connect.NewError(connect.CodeResourceExhausted, err)
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("error creating hash: %w", err))
	}

	ret := &apiv1.PasswordHashResponse{
		Hash: hash,
	}

	return connect.NewResponse(ret), nil
}

func (api *oliveTinAPI) cookieSecure(header http.Header) bool {
	useTLS := header.Get("X-Forwarded-Proto") == "https"
	return useTLS || api.cfg.Security.ForceSecureCookies
}

func (api *oliveTinAPI) applyLocalLoginResult(req *apiv1.LocalUserLoginRequest, response *connect.Response[apiv1.LocalUserLoginResponse], match bool, secure bool) {
	if match {
		user := api.cfg.FindUserByUsername(req.Username)
		if user != nil {
			sid := uuid.NewString()
			auth.RegisterUserSession(api.cfg, "local", sid, user.Username)
			log.WithFields(log.Fields{"username": user.Username}).Info("LocalUserLogin: Session created and registered")
			cookie := &http.Cookie{
				Name:     "olivetin-sid-local",
				Value:    sid,
				MaxAge:   31556952,
				HttpOnly: true,
				Path:     "/",
				Secure:   secure,
				SameSite: http.SameSiteLaxMode,
			}
			response.Header().Set("Set-Cookie", cookie.String())
			log.WithFields(log.Fields{"username": user.Username}).Info("LocalUserLogin: User logged in successfully.")
		} else {
			log.WithFields(log.Fields{"username": req.Username}).Warn("LocalUserLogin: Password matched but user lookup failed.")
		}
	} else {
		log.WithFields(log.Fields{"username": req.Username}).Warn("LocalUserLogin: User login failed.")
	}
}

func (api *oliveTinAPI) LocalUserLogin(ctx ctx.Context, req *connect.Request[apiv1.LocalUserLoginRequest]) (*connect.Response[apiv1.LocalUserLoginResponse], error) {
	if !api.cfg.AuthLocalUsers.Enabled {
		return connect.NewResponse(&apiv1.LocalUserLoginResponse{Success: false}), nil
	}
	match, err := checkUserPassword(api.cfg, req.Msg.Username, req.Msg.Password)
	if err != nil {
		if errors.Is(err, ErrArgon2Busy) {
			return nil, connect.NewError(connect.CodeResourceExhausted, err)
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("checking password: %w", err))
	}
	response := connect.NewResponse(&apiv1.LocalUserLoginResponse{Success: match})
	api.applyLocalLoginResult(req.Msg, response, match, api.cookieSecure(req.Header()))
	return response, nil
}

func (api *oliveTinAPI) startActionAndWaitRun(binding *executor.ActionBinding, args map[string]string, user *authpublic.AuthenticatedUser) (*executor.InternalLogEntry, bool) {
	execReq := executor.ExecutionRequest{
		Binding:           binding,
		TrackingID:        uuid.NewString(),
		Arguments:         args,
		AuthenticatedUser: user,
		Cfg:               api.cfg,
	}
	wg, _ := api.executor.ExecRequest(&execReq)
	wg.Wait()
	return api.executor.GetLog(execReq.TrackingID)
}

func (api *oliveTinAPI) findBindingOrNotFound(actionId string) (*executor.ActionBinding, error) {
	binding := api.executor.FindBindingByID(actionId)
	if binding == nil || binding.Action == nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("action with ID %s not found", actionId))
	}
	return binding, nil
}

func (api *oliveTinAPI) StartActionAndWait(ctx ctx.Context, req *connect.Request[apiv1.StartActionAndWaitRequest]) (*connect.Response[apiv1.StartActionAndWaitResponse], error) {
	binding, err := api.findBindingOrNotFound(req.Msg.ActionId)
	if err != nil {
		return nil, err
	}

	args := make(map[string]string)
	for _, arg := range req.Msg.Arguments {
		args[arg.Name] = arg.Value
	}
	user := auth.UserFromApiCall(ctx, req, api.cfg)

	internalLogEntry, ok := api.startActionAndWaitRun(binding, args, user)
	if !ok {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("execution not found"))
	}
	return connect.NewResponse(&apiv1.StartActionAndWaitResponse{
		LogEntry: api.internalLogEntryToPb(internalLogEntry, user),
	}), nil
}

func (api *oliveTinAPI) StartActionByGet(ctx ctx.Context, req *connect.Request[apiv1.StartActionByGetRequest]) (*connect.Response[apiv1.StartActionByGetResponse], error) {
	binding := api.executor.FindBindingByID(req.Msg.ActionId)
	if binding == nil || binding.Action == nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("action with ID %s not found", req.Msg.ActionId))
	}

	args := make(map[string]string)

	execReq := executor.ExecutionRequest{
		Binding:           binding,
		TrackingID:        uuid.NewString(),
		Arguments:         args,
		AuthenticatedUser: auth.UserFromApiCall(ctx, req, api.cfg),
		Cfg:               api.cfg,
	}

	_, uniqueTrackingId := api.executor.ExecRequest(&execReq)

	return connect.NewResponse(&apiv1.StartActionByGetResponse{
		ExecutionTrackingId: uniqueTrackingId,
	}), nil
}

func (api *oliveTinAPI) StartActionByGetAndWait(ctx ctx.Context, req *connect.Request[apiv1.StartActionByGetAndWaitRequest]) (*connect.Response[apiv1.StartActionByGetAndWaitResponse], error) {
	binding := api.executor.FindBindingByID(req.Msg.ActionId)
	if binding == nil || binding.Action == nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("action with ID %s not found", req.Msg.ActionId))
	}

	args := make(map[string]string)

	user := auth.UserFromApiCall(ctx, req, api.cfg)

	execReq := executor.ExecutionRequest{
		Binding:           binding,
		TrackingID:        uuid.NewString(),
		Arguments:         args,
		AuthenticatedUser: user,
		Cfg:               api.cfg,
	}

	wg, _ := api.executor.ExecRequest(&execReq)
	wg.Wait()

	internalLogEntry, ok := api.executor.GetLog(execReq.TrackingID)

	if ok {
		return connect.NewResponse(&apiv1.StartActionByGetAndWaitResponse{
			LogEntry: api.internalLogEntryToPb(internalLogEntry, user),
		}), nil
	}
	return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("execution not found"))
}

func calculateRateLimitExpires(api *oliveTinAPI, logEntry *executor.InternalLogEntry) string {
	if logEntry.Binding == nil || logEntry.Binding.Action == nil {
		return ""
	}

	expiryUnix := api.executor.GetTimeUntilAvailable(logEntry.Binding)
	if expiryUnix <= 0 {
		return ""
	}

	return time.Unix(expiryUnix, 0).Format("2006-01-02 15:04:05")
}

func (api *oliveTinAPI) internalLogEntryToPb(logEntry *executor.InternalLogEntry, authenticatedUser *authpublic.AuthenticatedUser) *apiv1.LogEntry {
	pble := &apiv1.LogEntry{
		ActionTitle:              logEntry.ActionTitle,
		ActionIcon:               logEntry.ActionIcon,
		DatetimeStarted:          logEntry.DatetimeStarted.Format("2006-01-02 15:04:05"),
		DatetimeFinished:         logEntry.DatetimeFinished.Format("2006-01-02 15:04:05"),
		DatetimeIndex:            logEntry.Index,
		Output:                   logEntry.Output,
		TimedOut:                 logEntry.TimedOut,
		Blocked:                  logEntry.Blocked,
		ExitCode:                 logEntry.ExitCode,
		Tags:                     logEntry.Tags,
		ExecutionTrackingId:      logEntry.ExecutionTrackingID,
		ExecutionStarted:         logEntry.ExecutionStarted,
		ExecutionFinished:        logEntry.ExecutionFinished,
		User:                     logEntry.Username,
		BindingId:                logEntry.GetBindingId(),
		DatetimeRateLimitExpires: calculateRateLimitExpires(api, logEntry),
	}

	if !pble.ExecutionFinished && logEntry.Binding != nil && logEntry.Binding.Action != nil {
		pble.CanKill = acl.IsAllowedKill(api.cfg, authenticatedUser, logEntry.Binding.Action)
	}

	return pble
}

func getExecutionStatusByTrackingID(api *oliveTinAPI, executionTrackingId string) *executor.InternalLogEntry {
	logEntry, ok := api.executor.GetLog(executionTrackingId)

	if !ok {
		return nil
	}

	return logEntry
}

// This is the actual action ID, not the binding ID.
func getMostRecentExecutionStatusByActionId(api *oliveTinAPI, actionId string) *executor.InternalLogEntry {
	var ile *executor.InternalLogEntry

	binding := api.executor.FindBindingByID(actionId)
	if binding == nil {
		return nil
	}

	logs := api.executor.GetLogsByBindingId(binding.ID)

	if len(logs) == 0 {
		return nil
	}

	if len(logs) == 0 {
		return nil
	} else {
		// Get last log entry
		ile = logs[len(logs)-1]
	}

	return ile
}

func (api *oliveTinAPI) resolveExecutionStatusForView(msg *apiv1.ExecutionStatusRequest, user *authpublic.AuthenticatedUser) (*executor.InternalLogEntry, error) {
	ile := api.getExecutionStatusByRequest(msg)
	if ile == nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("execution not found for tracking ID %s or action ID %s", msg.ExecutionTrackingId, msg.ActionId))
	}
	if !isValidLogEntry(ile) || !api.isLogEntryAllowed(ile, user) {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("permission denied to view this execution"))
	}
	return ile, nil
}

func (api *oliveTinAPI) getExecutionStatusByRequest(msg *apiv1.ExecutionStatusRequest) *executor.InternalLogEntry {
	if msg.ExecutionTrackingId != "" {
		return getExecutionStatusByTrackingID(api, msg.ExecutionTrackingId)
	}
	return getMostRecentExecutionStatusByActionId(api, msg.ActionId)
}

func (api *oliveTinAPI) ExecutionStatus(ctx ctx.Context, req *connect.Request[apiv1.ExecutionStatusRequest]) (*connect.Response[apiv1.ExecutionStatusResponse], error) {
	user := auth.UserFromApiCall(ctx, req, api.cfg)
	if err := api.checkDashboardAccess(user); err != nil {
		return nil, err
	}
	ile, err := api.resolveExecutionStatusForView(req.Msg, user)
	if err != nil {
		return nil, err
	}
	res := &apiv1.ExecutionStatusResponse{
		LogEntry: api.internalLogEntryToPb(ile, user),
	}
	return connect.NewResponse(res), nil
}

func (api *oliveTinAPI) Logout(ctx ctx.Context, req *connect.Request[apiv1.LogoutRequest]) (*connect.Response[apiv1.LogoutResponse], error) {
	user := auth.UserFromApiCall(ctx, req, api.cfg)

	log.WithFields(log.Fields{
		"username": user.Username,
		"provider": user.Provider,
	}).Info("Logout: User logged out")

	response := connect.NewResponse(&apiv1.LogoutResponse{})
	secure := api.cookieSecure(req.Header())

	// Clear the local authentication cookie by setting it to expire
	localCookie := &http.Cookie{
		Name:     "olivetin-sid-local",
		Value:    "",
		MaxAge:   -1, // This tells the browser to delete the cookie
		HttpOnly: true,
		Path:     "/",
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	}
	response.Header().Set("Set-Cookie", localCookie.String())

	// Clear the OAuth2 authentication cookie by setting it to expire
	oauth2Cookie := &http.Cookie{
		Name:     "olivetin-sid-oauth",
		Value:    "",
		MaxAge:   -1, // This tells the browser to delete the cookie
		HttpOnly: true,
		Path:     "/",
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	}
	response.Header().Add("Set-Cookie", oauth2Cookie.String())

	return response, nil
}

func (api *oliveTinAPI) GetActionBinding(ctx ctx.Context, req *connect.Request[apiv1.GetActionBindingRequest]) (*connect.Response[apiv1.GetActionBindingResponse], error) {
	user := auth.UserFromApiCall(ctx, req, api.cfg)

	if err := api.checkDashboardAccess(user); err != nil {
		return nil, err
	}

	binding := api.executor.FindBindingByID(req.Msg.BindingId)

	if binding == nil || binding.Action == nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("action with ID %s not found", req.Msg.BindingId))
	}

	return connect.NewResponse(&apiv1.GetActionBindingResponse{
		Action: buildAction(binding, &DashboardRenderRequest{
			cfg:               api.cfg,
			AuthenticatedUser: user,
			ex:                api.executor,
		}),
	}), nil
}

func (api *oliveTinAPI) GetDashboard(ctx ctx.Context, req *connect.Request[apiv1.GetDashboardRequest]) (*connect.Response[apiv1.GetDashboardResponse], error) {
	user := auth.UserFromApiCall(ctx, req, api.cfg)

	if err := api.checkDashboardAccess(user); err != nil {
		return nil, err
	}

	entityType := ""
	entityKey := ""
	if req.Msg != nil {
		entityType = req.Msg.EntityType
		entityKey = req.Msg.EntityKey
	}
	dashboardRenderRequest := api.createDashboardRenderRequest(user, entityType, entityKey)

	if api.isDefaultDashboard(req.Msg.Title) {
		return api.buildDefaultDashboardResponse(dashboardRenderRequest)
	}

	return api.buildCustomDashboardResponse(dashboardRenderRequest, req.Msg.Title)
}

func (api *oliveTinAPI) checkDashboardAccess(user *authpublic.AuthenticatedUser) error {
	if user.IsGuest() && api.cfg.AuthRequireGuestsToLogin {
		return connect.NewError(connect.CodePermissionDenied, fmt.Errorf("guests are not allowed to access the dashboard"))
	}
	return nil
}

func (api *oliveTinAPI) createDashboardRenderRequest(user *authpublic.AuthenticatedUser, entityType, entityKey string) *DashboardRenderRequest {
	return &DashboardRenderRequest{
		AuthenticatedUser: user,
		cfg:               api.cfg,
		ex:                api.executor,
		EntityType:        entityType,
		EntityKey:         entityKey,
	}
}

func (api *oliveTinAPI) isDefaultDashboard(title string) bool {
	return title == "default" || title == "" || title == "Actions"
}

func (api *oliveTinAPI) buildDefaultDashboardResponse(rr *DashboardRenderRequest) (*connect.Response[apiv1.GetDashboardResponse], error) {
	db := buildDefaultDashboard(rr)
	res := &apiv1.GetDashboardResponse{
		Dashboard: db,
	}
	return connect.NewResponse(res), nil
}

func (api *oliveTinAPI) buildCustomDashboardResponse(rr *DashboardRenderRequest, title string) (*connect.Response[apiv1.GetDashboardResponse], error) {
	res := &apiv1.GetDashboardResponse{
		Dashboard: renderDashboard(rr, title),
	}
	return connect.NewResponse(res), nil
}

func resolveLogsPageSize(requestPageSize, defaultPageSize int64) int64 {
	if requestPageSize == 0 {
		return defaultPageSize
	}
	if requestPageSize < 10 {
		return 10
	}
	if requestPageSize > 100 {
		return 100
	}
	return requestPageSize
}

func (api *oliveTinAPI) GetLogs(ctx ctx.Context, req *connect.Request[apiv1.GetLogsRequest]) (*connect.Response[apiv1.GetLogsResponse], error) {
	user := auth.UserFromApiCall(ctx, req, api.cfg)

	if err := api.checkDashboardAccess(user); err != nil {
		return nil, err
	}

	pageSize := resolveLogsPageSize(req.Msg.GetPageSize(), api.cfg.LogHistoryPageSize)
	logEntries, paging := api.executor.GetLogTrackingIdsACL(api.cfg, user, req.Msg.StartOffset, pageSize, req.Msg.DateFilter)
	ret := &apiv1.GetLogsResponse{}
	for _, le := range logEntries {
		ret.Logs = append(ret.Logs, api.internalLogEntryToPb(le, user))
	}
	ret.CountRemaining = paging.CountRemaining
	ret.PageSize = paging.PageSize
	ret.TotalCount = paging.TotalCount
	ret.StartOffset = paging.StartOffset
	return connect.NewResponse(ret), nil
}

// isValidLogEntry checks if a log entry has all required fields populated.
func isValidLogEntry(e *executor.InternalLogEntry) bool {
	return e != nil && e.Binding != nil && e.Binding.Action != nil
}

// isLogEntryAllowed checks if a log entry is allowed to be viewed by the user.
func (api *oliveTinAPI) isLogEntryAllowed(e *executor.InternalLogEntry, user *authpublic.AuthenticatedUser) bool {
	return acl.IsAllowedLogs(api.cfg, user, e.Binding.Action)
}

// buildEmptyPageResponse creates a response for an empty page.
func buildEmptyPageResponse(page pageInfo) *apiv1.GetActionLogsResponse {
	return &apiv1.GetActionLogsResponse{
		CountRemaining: 0,
		PageSize:       page.size,
		TotalCount:     page.total,
		StartOffset:    page.start,
	}
}

// calculateReversedIndices computes the reversed indices for newest-first pagination.
func calculateReversedIndices(page pageInfo, filteredLen int) (int64, int64) {
	startIdx := page.total - page.end
	endIdx := page.total - page.start
	if startIdx < 0 {
		startIdx = 0
	}
	if endIdx > int64(filteredLen) {
		endIdx = int64(filteredLen)
	}
	return startIdx, endIdx
}

// buildActionLogsResponse builds the response with paginated log entries.
func (api *oliveTinAPI) buildActionLogsResponse(filtered []*executor.InternalLogEntry, page pageInfo, user *authpublic.AuthenticatedUser) *apiv1.GetActionLogsResponse {
	startIdx, endIdx := calculateReversedIndices(page, len(filtered))
	ret := &apiv1.GetActionLogsResponse{}
	for _, le := range filtered[startIdx:endIdx] {
		ret.Logs = append(ret.Logs, api.internalLogEntryToPb(le, user))
	}
	ret.CountRemaining = page.start
	ret.PageSize = page.size
	ret.TotalCount = page.total
	ret.StartOffset = page.start
	return ret
}

func (api *oliveTinAPI) GetActionLogs(ctx ctx.Context, req *connect.Request[apiv1.GetActionLogsRequest]) (*connect.Response[apiv1.GetActionLogsResponse], error) {
	user := auth.UserFromApiCall(ctx, req, api.cfg)

	if err := api.checkDashboardAccess(user); err != nil {
		return nil, err
	}

	filtered := api.filterLogsByACL(api.executor.GetLogsByBindingId(req.Msg.ActionId), user)
	page := paginate(int64(len(filtered)), api.cfg.LogHistoryPageSize, req.Msg.StartOffset)
	if page.empty {
		return connect.NewResponse(buildEmptyPageResponse(page)), nil
	}

	return connect.NewResponse(api.buildActionLogsResponse(filtered, page, user)), nil
}

func (api *oliveTinAPI) filterLogsByACL(entries []*executor.InternalLogEntry, user *authpublic.AuthenticatedUser) []*executor.InternalLogEntry {
	filtered := make([]*executor.InternalLogEntry, 0, len(entries))
	for _, e := range entries {
		if !isValidLogEntry(e) {
			continue
		}
		if api.isLogEntryAllowed(e, user) {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

type pageInfo struct {
	total int64
	size  int64
	start int64
	end   int64
	empty bool
}

func paginate(total int64, size int64, start int64) pageInfo {
	if start < 0 {
		start = 0
	}
	if start >= total {
		return pageInfo{total: total, size: size, start: start, end: start, empty: true}
	}
	end := start + size
	if end > total {
		end = total
	}
	return pageInfo{total: total, size: size, start: start, end: end, empty: false}
}

/*
This function is ONLY a helper for the UI - the arguments are validated properly
on the StartAction -> Executor chain. This is here basically to provide helpful
error messages more quickly before starting the action.

It uses the same validation logic as the executor, including mangling argument
values (e.g., datetime formatting, checkbox title-to-value conversion).
*/
func (api *oliveTinAPI) argumentNotFoundForValidation(msg *apiv1.ValidateArgumentTypeRequest) bool {
	arg, _ := api.findArgumentForValidation(msg.BindingId, msg.ArgumentName)
	return arg == nil && (msg.BindingId != "" || msg.ArgumentName != "")
}

func (api *oliveTinAPI) ValidateArgumentType(ctx ctx.Context, req *connect.Request[apiv1.ValidateArgumentTypeRequest]) (*connect.Response[apiv1.ValidateArgumentTypeResponse], error) {
	if api.argumentNotFoundForValidation(req.Msg) {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("action or argument not found for binding ID %s", req.Msg.BindingId))
	}

	err := api.validateArgumentTypeInternal(req.Msg)
	desc := ""
	if err != nil {
		desc = err.Error()
	}

	return connect.NewResponse(&apiv1.ValidateArgumentTypeResponse{
		Valid:       err == nil,
		Description: desc,
	}), nil
}

func (api *oliveTinAPI) validateArgumentTypeInternal(msg *apiv1.ValidateArgumentTypeRequest) error {
	if msg.BindingId == "" || msg.ArgumentName == "" {
		return executor.TypeSafetyCheck("", msg.Value, msg.Type)
	}

	arg, action := api.findArgumentForValidation(msg.BindingId, msg.ArgumentName)
	if arg == nil {
		return fmt.Errorf("argument not found")
	}

	return executor.ValidateArgument(arg, msg.Value, action)
}

func (api *oliveTinAPI) findArgumentForValidation(bindingId string, argumentName string) (*config.ActionArgument, *config.Action) {
	binding := api.executor.FindBindingByID(bindingId)
	if binding == nil || binding.Action == nil {
		return nil, nil
	}

	arg := api.findArgumentByName(binding.Action, argumentName)
	return arg, binding.Action
}

func (api *oliveTinAPI) findArgumentByName(action *config.Action, name string) *config.ActionArgument {
	for i := range action.Arguments {
		if action.Arguments[i].Name == name {
			return &action.Arguments[i]
		}
	}
	return nil
}

func (api *oliveTinAPI) WhoAmI(ctx ctx.Context, req *connect.Request[apiv1.WhoAmIRequest]) (*connect.Response[apiv1.WhoAmIResponse], error) {
	user := auth.UserFromApiCall(ctx, req, api.cfg)

	if err := api.checkDashboardAccess(user); err != nil {
		return nil, err
	}

	res := &apiv1.WhoAmIResponse{
		AuthenticatedUser: user.Username,
		Usergroup:         user.UsergroupLine,
		Provider:          user.Provider,
		Sid:               user.SID,
		Acls:              user.Acls,
	}

	return connect.NewResponse(res), nil
}

func (api *oliveTinAPI) SosReport(ctx ctx.Context, req *connect.Request[apiv1.SosReportRequest]) (*connect.Response[apiv1.SosReportResponse], error) {
	user := auth.UserFromApiCall(ctx, req, api.cfg)
	redactVersion := !user.EffectivePolicy.ShowVersionNumber
	sos := installationinfo.GetSosReport(redactVersion)

	if !api.cfg.InsecureAllowDumpSos {
		log.Info(sos)
		sos = "Your SOS Report has been logged to OliveTin logs.\n\nIf you are in a safe network, you can temporarily set `insecureAllowDumpSos: true` in your config.yaml, restart OliveTin, and refresh this page - it will put the output directly in the browser."
	}

	ret := &apiv1.SosReportResponse{
		Alert: sos,
	}

	return connect.NewResponse(ret), nil
}

func (api *oliveTinAPI) DumpVars(ctx ctx.Context, req *connect.Request[apiv1.DumpVarsRequest]) (*connect.Response[apiv1.DumpVarsResponse], error) {
	res := &apiv1.DumpVarsResponse{}

	if !api.cfg.InsecureAllowDumpVars {
		res.Alert = "Dumping variables is not allowed by default because it is insecure."

		return connect.NewResponse(res), nil
	}

	jsonstring, err := json.MarshalIndent(tpl.GetNewGeneralTemplateContext(), "", "  ")
	if err != nil {
		log.WithError(err).Error("DumpVars: failed to marshal template context from GetNewGeneralTemplateContext")
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("dump vars: marshal template context: %w", err))
	}
	fmt.Printf("%s", jsonstring)

	res.Alert = "Dumping variables has been enabled in the configuration. Please set InsecureAllowDumpVars = false again after you don't need it anymore"

	return connect.NewResponse(res), nil
}

func debugBindingActionTitle(binding *executor.ActionBinding) string {
	if binding == nil || binding.Action == nil {
		return ""
	}
	return binding.Action.Title
}

func (api *oliveTinAPI) DumpPublicIdActionMap(ctx ctx.Context, req *connect.Request[apiv1.DumpPublicIdActionMapRequest]) (*connect.Response[apiv1.DumpPublicIdActionMapResponse], error) {
	res := &apiv1.DumpPublicIdActionMapResponse{}
	res.Contents = make(map[string]*apiv1.DebugBinding)

	if !api.cfg.InsecureAllowDumpActionMap {
		res.Alert = "Dumping Public IDs is disallowed."

		return connect.NewResponse(res), nil
	}

	api.executor.MapActionBindingsLock.RLock()

	for k, v := range api.executor.MapActionBindings {
		res.Contents[k] = &apiv1.DebugBinding{
			ActionTitle: debugBindingActionTitle(v),
		}
	}

	api.executor.MapActionBindingsLock.RUnlock()

	res.Alert = "Dumping variables has been enabled in the configuration. Please set InsecureAllowDumpActionMap = false again after you don't need it anymore"

	return connect.NewResponse(res), nil
}

func (api *oliveTinAPI) GetReadyz(ctx ctx.Context, req *connect.Request[apiv1.GetReadyzRequest]) (*connect.Response[apiv1.GetReadyzResponse], error) {
	res := &apiv1.GetReadyzResponse{
		Status: "OK",
	}

	return connect.NewResponse(res), nil
}

func (api *oliveTinAPI) EventStream(ctx ctx.Context, req *connect.Request[apiv1.EventStreamRequest], srv *connect.ServerStream[apiv1.EventStreamResponse]) error {
	log.Debugf("EventStream: %v", req.Msg)

	// Set X-Accel-Buffering header to disable nginx buffering for this stream
	// https://github.com/OliveTin/OliveTin/issues/765
	srv.ResponseHeader().Set("X-Accel-Buffering", "no")

	user := auth.UserFromApiCall(ctx, req, api.cfg)

	if err := api.checkDashboardAccess(user); err != nil {
		return err
	}

	client := &streamingClient{
		channel:           make(chan *apiv1.EventStreamResponse, 10), // Buffered channel to hold Events
		AuthenticatedUser: user,
	}

	log.WithFields(log.Fields{
		"authenticatedUser": user.Username,
	}).Debugf("EventStream: client connected")

	api.streamingClientsMutex.Lock()
	api.streamingClients[client] = struct{}{}
	api.streamingClientsMutex.Unlock()

	// loop over client channel and send events to connectedClient
	for msg := range client.channel {
		log.Debugf("Sending event to client: %v", msg)
		if err := srv.Send(msg); err != nil {
			log.Errorf("Error sending event to client: %v", err)
			// Remove disconnected client from the list
			api.removeClient(client)
			break
		}
	}

	log.Infof("EventStream: client disconnected")

	return nil
}

func (api *oliveTinAPI) removeClient(clientToRemove *streamingClient) {
	api.streamingClientsMutex.Lock()
	delete(api.streamingClients, clientToRemove)
	api.streamingClientsMutex.Unlock()
	close(clientToRemove.channel)
}

func (api *oliveTinAPI) OnActionMapRebuilt() {
	toRemove := []*streamingClient{}

	for _, client := range api.copyOfStreamingClients() {
		select {
		case client.channel <- &apiv1.EventStreamResponse{
			Event: &apiv1.EventStreamResponse_ConfigChanged{
				ConfigChanged: &apiv1.EventConfigChanged{},
			},
		}:
		default:
			log.Warnf("EventStream: client channel is full, removing client")
			toRemove = append(toRemove, client)
		}
	}

	for _, client := range toRemove {
		api.removeClient(client)
	}
}

func (api *oliveTinAPI) OnExecutionStarted(ex *executor.InternalLogEntry) {
	toRemove := []*streamingClient{}

	for _, client := range api.copyOfStreamingClients() {
		select {
		case client.channel <- &apiv1.EventStreamResponse{
			Event: &apiv1.EventStreamResponse_ExecutionStarted{
				ExecutionStarted: &apiv1.EventExecutionStarted{
					LogEntry: api.internalLogEntryToPb(ex, client.AuthenticatedUser),
				},
			},
		}:
		default:
			log.Warnf("EventStream: client channel is full, removing client")
			toRemove = append(toRemove, client)
		}
	}

	for _, client := range toRemove {
		api.removeClient(client)
	}
}

func (api *oliveTinAPI) OnExecutionFinished(ile *executor.InternalLogEntry) {
	toRemove := []*streamingClient{}

	for _, client := range api.copyOfStreamingClients() {
		select {
		case client.channel <- &apiv1.EventStreamResponse{
			Event: &apiv1.EventStreamResponse_ExecutionFinished{
				ExecutionFinished: &apiv1.EventExecutionFinished{
					LogEntry: api.internalLogEntryToPb(ile, client.AuthenticatedUser),
				},
			},
		}:
		default:
			log.Warnf("EventStream: client channel is full, removing client")
			toRemove = append(toRemove, client)
		}
	}

	for _, client := range toRemove {
		api.removeClient(client)
	}
}

func (api *oliveTinAPI) GetDiagnostics(ctx ctx.Context, req *connect.Request[apiv1.GetDiagnosticsRequest]) (*connect.Response[apiv1.GetDiagnosticsResponse], error) {
	user := auth.UserFromApiCall(ctx, req, api.cfg)
	if err := api.checkDashboardAccess(user); err != nil {
		return nil, err
	}
	if !user.EffectivePolicy.ShowDiagnostics {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("diagnostics are not available for your account"))
	}
	res := &apiv1.GetDiagnosticsResponse{
		SshFoundKey:    installationinfo.Runtime.SshFoundKey,
		SshFoundConfig: installationinfo.Runtime.SshFoundConfig,
	}
	return connect.NewResponse(res), nil
}

func (api *oliveTinAPI) Init(ctx ctx.Context, req *connect.Request[apiv1.InitRequest]) (*connect.Response[apiv1.InitResponse], error) {
	user := auth.UserFromApiCall(ctx, req, api.cfg)

	loginRequired := user.IsGuest() && api.cfg.AuthRequireGuestsToLogin

	showVersion := user.EffectivePolicy.ShowVersionNumber
	currentVersion := ""
	availableVersion := ""
	if showVersion {
		currentVersion = installationinfo.Build.Version
		availableVersion = installationinfo.Runtime.AvailableVersion
	}
	res := &apiv1.InitResponse{
		ShowFooter:                api.cfg.ShowFooter,
		ShowNavigation:            api.cfg.ShowNavigation,
		ShowNewVersions:           showVersion && api.cfg.ShowNewVersions,
		AvailableVersion:          availableVersion,
		CurrentVersion:            currentVersion,
		PageTitle:                 api.cfg.PageTitle,
		SectionNavigationStyle:    api.cfg.SectionNavigationStyle,
		DefaultIconForBack:        api.cfg.DefaultIconForBack,
		EnableCustomJs:            api.cfg.EnableCustomJs,
		AuthLoginUrl:              api.cfg.AuthLoginUrl,
		AuthLocalLogin:            api.cfg.AuthLocalUsers.Enabled,
		OAuth2Providers:           buildPublicOAuth2ProvidersList(api.cfg),
		AdditionalLinks:           buildAdditionalLinks(api.cfg.AdditionalNavigationLinks),
		StyleMods:                 api.cfg.StyleMods,
		RootDashboards:            api.buildRootDashboards(user, api.cfg.Dashboards),
		AuthenticatedUser:         user.Username,
		AuthenticatedUserProvider: user.Provider,
		EffectivePolicy:           buildEffectivePolicy(user.EffectivePolicy),
		BannerMessage:             api.cfg.BannerMessage,
		BannerCss:                 api.cfg.BannerCSS,
		ShowDiagnostics:           user.EffectivePolicy.ShowDiagnostics,
		ShowLogList:               user.EffectivePolicy.ShowLogList,
		LoginRequired:             loginRequired,
		AvailableThemes:           discoverAvailableThemes(api.cfg),
		ShowNavigateOnStartIcons:  api.cfg.ShowNavigateOnStartIcons,
	}

	return connect.NewResponse(res), nil
}

// discoverAvailableThemes finds all available themes in the custom-webui/themes directory.
// A theme is considered available if it has a theme.css file.
func discoverAvailableThemes(cfg *config.Config) []string {
	configDir := cfg.GetDir()
	if configDir == "" {
		return []string{}
	}

	themesDir := path.Join(configDir, "custom-webui", "themes")
	entries, err := os.ReadDir(themesDir)
	if err != nil {
		log.WithFields(log.Fields{
			"themesDir": themesDir,
			"error":     err,
		}).Tracef("Could not read themes directory")
		return []string{}
	}

	themes := collectValidThemes(themesDir, entries)
	sort.Strings(themes)
	return themes
}

// collectValidThemes collects theme names from directory entries that have a theme.css file.
func collectValidThemes(themesDir string, entries []os.DirEntry) []string {
	var themes []string
	for _, entry := range entries {
		if themeName := getValidThemeName(themesDir, entry); themeName != "" {
			themes = append(themes, themeName)
		}
	}
	return themes
}

// getValidThemeName returns the theme name if the entry is a valid theme directory with theme.css, otherwise returns empty string.
func getValidThemeName(themesDir string, entry os.DirEntry) string {
	if !entry.IsDir() {
		return ""
	}

	themeName := entry.Name()
	themeCssPath := path.Join(themesDir, themeName, "theme.css")

	if _, err := os.Stat(themeCssPath); err != nil {
		return ""
	}

	return themeName
}

func (api *oliveTinAPI) buildRootDashboards(user *authpublic.AuthenticatedUser, dashboards []*config.DashboardComponent) []string {
	var rootDashboards []string
	dashboardRenderRequest := api.createDashboardRenderRequest(user, "", "")

	api.addDefaultDashboardIfNeeded(&rootDashboards, dashboardRenderRequest)
	api.addCustomDashboards(&rootDashboards, dashboards, dashboardRenderRequest)

	return rootDashboards
}

func (api *oliveTinAPI) addDefaultDashboardIfNeeded(rootDashboards *[]string, rr *DashboardRenderRequest) {
	defaultDashboard := buildDefaultDashboard(rr)
	if defaultDashboard != nil && len(defaultDashboard.Contents) > 0 {
		log.Tracef("defaultDashboard: %+v", defaultDashboard.Contents)
		*rootDashboards = append(*rootDashboards, "Actions")
	}
}

func (api *oliveTinAPI) addCustomDashboards(rootDashboards *[]string, dashboards []*config.DashboardComponent, rr *DashboardRenderRequest) {
	for _, dashboard := range dashboards {
		// We have to build the dashboard response instead of just looping over config.dashboards,
		// because we need to check if the user has access to the dashboard
		db := renderDashboard(rr, dashboard.Title)
		if db != nil {
			*rootDashboards = append(*rootDashboards, dashboard.Title)
		}
	}
}

func buildPublicOAuth2ProvidersList(cfg *config.Config) []*apiv1.OAuth2Provider {
	var publicProviders []*apiv1.OAuth2Provider

	for providerKey, provider := range cfg.AuthOAuth2Providers {
		publicProviders = append(publicProviders, &apiv1.OAuth2Provider{
			Title: provider.Title,
			Icon:  provider.Icon,
			Key:   providerKey,
		})
	}

	sort.Slice(publicProviders, func(i, j int) bool {
		return publicProviders[i].Key < publicProviders[j].Key
	})

	return publicProviders
}

func buildAdditionalLinks(links []*config.NavigationLink) []*apiv1.AdditionalLink {
	var additionalLinks []*apiv1.AdditionalLink

	for _, link := range links {
		additionalLinks = append(additionalLinks, &apiv1.AdditionalLink{
			Title: link.Title,
			Url:   link.Url,
		})
	}

	return additionalLinks
}

func (api *oliveTinAPI) OnOutputChunk(content []byte, executionTrackingId string) {
	toRemove := []*streamingClient{}

	for _, client := range api.copyOfStreamingClients() {
		select {
		case client.channel <- &apiv1.EventStreamResponse{
			Event: &apiv1.EventStreamResponse_OutputChunk{
				OutputChunk: &apiv1.EventOutputChunk{
					Output:              string(content),
					ExecutionTrackingId: executionTrackingId,
				},
			},
		}:
		default:
			log.Warnf("EventStream: client channel is full, removing client")
			toRemove = append(toRemove, client)
		}
	}

	for _, client := range toRemove {
		api.removeClient(client)
	}
}

func (api *oliveTinAPI) GetEntities(ctx ctx.Context, req *connect.Request[apiv1.GetEntitiesRequest]) (*connect.Response[apiv1.GetEntitiesResponse], error) {
	user := auth.UserFromApiCall(ctx, req, api.cfg)

	if err := api.checkDashboardAccess(user); err != nil {
		return nil, err
	}

	entityMap := entities.GetEntities()
	entityNames := make([]string, 0, len(entityMap))
	for name := range entityMap {
		entityNames = append(entityNames, name)
	}
	sort.Strings(entityNames)

	entityDefinitions := make([]*apiv1.EntityDefinition, 0, len(entityNames))
	for _, name := range entityNames {
		def := &apiv1.EntityDefinition{
			Title:            name,
			UsedOnDashboards: findDashboardsForEntity(name, api.cfg.Dashboards),
			Instances:        buildSortedEntityInstances(name, entityMap[name]),
		}
		entityDefinitions = append(entityDefinitions, def)
	}

	res := &apiv1.GetEntitiesResponse{
		EntityDefinitions: entityDefinitions,
	}

	return connect.NewResponse(res), nil
}

func buildSortedEntityInstances(entityType string, entityInstances map[string]*entities.Entity) []*apiv1.Entity {
	instanceKeys := make([]string, 0, len(entityInstances))
	for key := range entityInstances {
		instanceKeys = append(instanceKeys, key)
	}
	sort.Strings(instanceKeys)

	instances := make([]*apiv1.Entity, 0, len(instanceKeys))
	for _, key := range instanceKeys {
		e := entityInstances[key]
		instances = append(instances, &apiv1.Entity{
			Title:     e.Title,
			UniqueKey: e.UniqueKey,
			Type:      entityType,
		})
	}
	return instances
}

func findDashboardsForEntity(entityTitle string, dashboards []*config.DashboardComponent) []string {
	var foundDashboards []string
	seen := make(map[string]bool)

	findEntityInComponents(entityTitle, "", dashboards, &foundDashboards, seen)

	return foundDashboards
}

func findEntityInComponents(entityTitle string, parentTitle string, components []*config.DashboardComponent, foundDashboards *[]string, seen map[string]bool) {
	for _, component := range components {
		if component.Entity == entityTitle {
			addEntityDashboard(component, parentTitle, foundDashboards, seen)
		}

		if len(component.Contents) > 0 {
			findEntityInComponents(entityTitle, component.Title, component.Contents, foundDashboards, seen)
		}
	}
}

func addEntityDashboard(component *config.DashboardComponent, parentTitle string, foundDashboards *[]string, seen map[string]bool) {
	if component.Type == "directory" {
		addEntityDirectory(component, foundDashboards, seen)
	} else {
		addParentDashboard(parentTitle, foundDashboards, seen)
	}
}

func addEntityDirectory(component *config.DashboardComponent, foundDashboards *[]string, seen map[string]bool) {
	dashboardTitle := component.Title + " [Entity Directory]"
	if !seen[dashboardTitle] {
		*foundDashboards = append(*foundDashboards, dashboardTitle)
		seen[dashboardTitle] = true
		seen[component.Title] = true
	}
}

func addParentDashboard(parentTitle string, foundDashboards *[]string, seen map[string]bool) {
	if parentTitle != "" && !seen[parentTitle] {
		*foundDashboards = append(*foundDashboards, parentTitle)
		seen[parentTitle] = true
	}
}

func findDirectoriesInEntityFieldsets(entityType string, dashboards []*config.DashboardComponent) []string {
	var directories []string

	for _, dashboard := range dashboards {
		findDirectoriesInEntityFieldsetsRecursive(entityType, dashboard, &directories)
	}

	return directories
}

func findDirectoriesInEntityFieldsetsRecursive(entityType string, component *config.DashboardComponent, directories *[]string) {
	if component.Entity == entityType {
		collectDirectoriesFromComponent(component, directories)
	}

	if len(component.Contents) > 0 {
		searchSubcomponentsForDirectories(entityType, component.Contents, directories)
	}
}

func collectDirectoriesFromComponent(component *config.DashboardComponent, directories *[]string) {
	for _, subitem := range component.Contents {
		if subitem.Type == "directory" {
			*directories = append(*directories, subitem.Title)
		}
	}
}

func searchSubcomponentsForDirectories(entityType string, contents []*config.DashboardComponent, directories *[]string) {
	for _, subitem := range contents {
		findDirectoriesInEntityFieldsetsRecursive(entityType, subitem, directories)
	}
}

func (api *oliveTinAPI) GetEntity(ctx ctx.Context, req *connect.Request[apiv1.GetEntityRequest]) (*connect.Response[apiv1.Entity], error) {
	user := auth.UserFromApiCall(ctx, req, api.cfg)

	if err := api.checkDashboardAccess(user); err != nil {
		return nil, err
	}

	instances := entities.GetEntityInstances(req.Msg.Type)
	if len(instances) == 0 {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("entity type %s not found", req.Msg.Type))
	}

	entity, ok := instances[req.Msg.UniqueKey]
	if !ok {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("entity with unique key %s not found in type %s", req.Msg.UniqueKey, req.Msg.Type))
	}

	res := buildEntityResponse(entity, req.Msg.Type, api.cfg.Dashboards)
	return connect.NewResponse(res), nil
}

func buildEntityResponse(entity *entities.Entity, entityType string, dashboards []*config.DashboardComponent) *apiv1.Entity {
	res := &apiv1.Entity{
		Title:       entity.Title,
		UniqueKey:   entity.UniqueKey,
		Type:        entityType,
		Directories: findDirectoriesInEntityFieldsets(entityType, dashboards),
		Fields:      serializeEntityFields(entity.Data),
	}
	return res
}

func serializeEntityFields(data any) map[string]string {
	if data == nil {
		return nil
	}

	dataMap, ok := data.(map[string]any)
	if !ok {
		return nil
	}

	fields := make(map[string]string)
	for k, v := range dataMap {
		fields[k] = fmt.Sprintf("%v", v)
	}
	return fields
}

func (api *oliveTinAPI) RestartAction(ctx ctx.Context, req *connect.Request[apiv1.RestartActionRequest]) (*connect.Response[apiv1.StartActionResponse], error) {
	var execReqLogEntry *executor.InternalLogEntry

	execReqLogEntry, found := api.executor.GetLog(req.Msg.ExecutionTrackingId)

	if !found {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("execution not found for tracking ID %s", req.Msg.ExecutionTrackingId))
	}

	if execReqLogEntry.Binding == nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("log entry has no binding for tracking ID %s", req.Msg.ExecutionTrackingId))
	}

	action := execReqLogEntry.Binding.Action

	if action == nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("action not found for tracking ID %s", req.Msg.ExecutionTrackingId))
	}

	log.Warnf("Restarting execution request by tracking ID: %v", req.Msg.ExecutionTrackingId)

	return api.StartAction(ctx, &connect.Request[apiv1.StartActionRequest]{
		Msg: &apiv1.StartActionRequest{
			BindingId:        execReqLogEntry.GetBindingId(),
			UniqueTrackingId: req.Msg.ExecutionTrackingId,
		},
	})
}

func newServer(ex *executor.Executor) *oliveTinAPI {
	server := oliveTinAPI{}
	server.cfg = ex.Cfg
	server.executor = ex
	server.streamingClients = make(map[*streamingClient]struct{})

	ex.AddListener(&server)
	return &server
}

func GetNewHandler(ex *executor.Executor) (string, http.Handler) {
	server := newServer(ex)

	jsonOpt := connectproto.WithJSON(
		protojson.MarshalOptions{
			EmitUnpopulated: true, // https://github.com/OliveTin/OliveTin/issues/674
		},
		protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	)

	return apiv1connect.NewOliveTinApiServiceHandler(server, jsonOpt)
}
