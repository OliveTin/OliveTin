package api

import (
	ctx "context"
	"encoding/json"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/encoding/protojson"

	apiv1 "github.com/OliveTin/OliveTin/gen/olivetin/api/v1"
	apiv1connect "github.com/OliveTin/OliveTin/gen/olivetin/api/v1/apiv1connect"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"fmt"
	"net/http"

	acl "github.com/OliveTin/OliveTin/internal/acl"
	auth "github.com/OliveTin/OliveTin/internal/auth"
	config "github.com/OliveTin/OliveTin/internal/config"
	entities "github.com/OliveTin/OliveTin/internal/entities"
	executor "github.com/OliveTin/OliveTin/internal/executor"
	installationinfo "github.com/OliveTin/OliveTin/internal/installationinfo"
	connectproto "go.akshayshah.org/connectproto"
)

type oliveTinAPI struct {
	executor *executor.Executor
	cfg      *config.Config

	connectedClients []*connectedClients
}

type connectedClients struct {
	channel           chan *apiv1.EventStreamResponse
	AuthenticatedUser *acl.AuthenticatedUser
}

func (api *oliveTinAPI) KillAction(ctx ctx.Context, req *connect.Request[apiv1.KillActionRequest]) (*connect.Response[apiv1.KillActionResponse], error) {
	ret := &apiv1.KillActionResponse{
		ExecutionTrackingId: req.Msg.ExecutionTrackingId,
	}

	var execReqLogEntry *executor.InternalLogEntry

	execReqLogEntry, ret.Found = api.executor.GetLog(req.Msg.ExecutionTrackingId)

	if !ret.Found {
		log.Warnf("Killing execution request not possible - not found by tracking ID: %v", req.Msg.ExecutionTrackingId)
		return connect.NewResponse(ret), nil
	}

	log.Warnf("Killing execution request by tracking ID: %v", req.Msg.ExecutionTrackingId)

	action := execReqLogEntry.Binding.Action

	if action == nil {
		log.Warnf("Killing execution request not possible - action not found: %v", execReqLogEntry.ActionTitle)
		ret.Killed = false
		return connect.NewResponse(ret), nil
	}

	user := acl.UserFromContext(ctx, req, api.cfg)

	api.killActionByTrackingId(user, action, execReqLogEntry, ret)

	return connect.NewResponse(ret), nil
}

func (api *oliveTinAPI) killActionByTrackingId(user *acl.AuthenticatedUser, action *config.Action, execReqLogEntry *executor.InternalLogEntry, ret *apiv1.KillActionResponse) {
	if !acl.IsAllowedKill(api.cfg, user, action) {
		log.Warnf("Killing execution request not possible - user not allowed to kill this action: %v", execReqLogEntry.ExecutionTrackingID)
		ret.Killed = false
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

	authenticatedUser := acl.UserFromContext(ctx, req, api.cfg)

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
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("error creating hash: %w", err))
	}

	ret := &apiv1.PasswordHashResponse{
		Hash: hash,
	}

	return connect.NewResponse(ret), nil
}

func (api *oliveTinAPI) LocalUserLogin(ctx ctx.Context, req *connect.Request[apiv1.LocalUserLoginRequest]) (*connect.Response[apiv1.LocalUserLoginResponse], error) {
	// Check if local user authentication is enabled
	if !api.cfg.AuthLocalUsers.Enabled {
		return connect.NewResponse(&apiv1.LocalUserLoginResponse{
			Success: false,
		}), nil
	}

	match := checkUserPassword(api.cfg, req.Msg.Username, req.Msg.Password)

	response := connect.NewResponse(&apiv1.LocalUserLoginResponse{
		Success: match,
	})

	if match {
		// Set authentication cookie for successful login
		user := api.cfg.FindUserByUsername(req.Msg.Username)
		if user != nil {
			sid := uuid.NewString()
			// Register the session in the session storage
			auth.RegisterUserSession(api.cfg, "local", sid, user.Username)

			log.WithFields(log.Fields{
				"username": user.Username,
			}).Info("LocalUserLogin: Session created and registered")

			// Set the authentication cookie in the response headers
			cookie := &http.Cookie{
				Name:     "olivetin-sid-local",
				Value:    sid,
				MaxAge:   31556952, // 1 year
				HttpOnly: true,
				Path:     "/",
			}
			response.Header().Set("Set-Cookie", cookie.String())
		}

		log.WithFields(log.Fields{
			"username": req.Msg.Username,
		}).Info("LocalUserLogin: User logged in successfully.")
	} else {
		log.WithFields(log.Fields{
			"username": req.Msg.Username,
		}).Warn("LocalUserLogin: User login failed.")
	}

	return response, nil
}

func (api *oliveTinAPI) StartActionAndWait(ctx ctx.Context, req *connect.Request[apiv1.StartActionAndWaitRequest]) (*connect.Response[apiv1.StartActionAndWaitResponse], error) {
	args := make(map[string]string)

	for _, arg := range req.Msg.Arguments {
		args[arg.Name] = arg.Value
	}

	user := acl.UserFromContext(ctx, req, api.cfg)

	execReq := executor.ExecutionRequest{
		Binding:           api.executor.FindBindingByID(req.Msg.ActionId),
		TrackingID:        uuid.NewString(),
		Arguments:         args,
		AuthenticatedUser: user,
		Cfg:               api.cfg,
	}

	wg, _ := api.executor.ExecRequest(&execReq)
	wg.Wait()

	internalLogEntry, ok := api.executor.GetLog(execReq.TrackingID)

	if ok {
		return connect.NewResponse(&apiv1.StartActionAndWaitResponse{
			LogEntry: api.internalLogEntryToPb(internalLogEntry, user),
		}), nil
	} else {
		return nil, fmt.Errorf("execution not found")
	}
}

func (api *oliveTinAPI) StartActionByGet(ctx ctx.Context, req *connect.Request[apiv1.StartActionByGetRequest]) (*connect.Response[apiv1.StartActionByGetResponse], error) {
	args := make(map[string]string)

	execReq := executor.ExecutionRequest{
		Binding:           api.executor.FindBindingByID(req.Msg.ActionId),
		TrackingID:        uuid.NewString(),
		Arguments:         args,
		AuthenticatedUser: acl.UserFromContext(ctx, req, api.cfg),
		Cfg:               api.cfg,
	}

	_, uniqueTrackingId := api.executor.ExecRequest(&execReq)

	return connect.NewResponse(&apiv1.StartActionByGetResponse{
		ExecutionTrackingId: uniqueTrackingId,
	}), nil
}

func (api *oliveTinAPI) StartActionByGetAndWait(ctx ctx.Context, req *connect.Request[apiv1.StartActionByGetAndWaitRequest]) (*connect.Response[apiv1.StartActionByGetAndWaitResponse], error) {
	args := make(map[string]string)

	user := acl.UserFromContext(ctx, req, api.cfg)

	execReq := executor.ExecutionRequest{
		Binding:           api.executor.FindBindingByID(req.Msg.ActionId),
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
	} else {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("execution not found"))
	}
}

func (api *oliveTinAPI) internalLogEntryToPb(logEntry *executor.InternalLogEntry, authenticatedUser *acl.AuthenticatedUser) *apiv1.LogEntry {
	pble := &apiv1.LogEntry{
		ActionTitle:         logEntry.ActionTitle,
		ActionIcon:          logEntry.ActionIcon,
		ActionId:            logEntry.ActionId,
		DatetimeStarted:     logEntry.DatetimeStarted.Format("2006-01-02 15:04:05"),
		DatetimeFinished:    logEntry.DatetimeFinished.Format("2006-01-02 15:04:05"),
		DatetimeIndex:       logEntry.Index,
		Output:              logEntry.Output,
		TimedOut:            logEntry.TimedOut,
		Blocked:             logEntry.Blocked,
		ExitCode:            logEntry.ExitCode,
		Tags:                logEntry.Tags,
		ExecutionTrackingId: logEntry.ExecutionTrackingID,
		ExecutionStarted:    logEntry.ExecutionStarted,
		ExecutionFinished:   logEntry.ExecutionFinished,
		User:                logEntry.Username,
	}

	if !pble.ExecutionFinished {
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

func getMostRecentExecutionStatusById(api *oliveTinAPI, actionId string) *executor.InternalLogEntry {
	var ile *executor.InternalLogEntry

	logs := api.executor.GetLogsByActionId(actionId)

	if len(logs) == 0 {
		return nil
	} else {
		// Get last log entry
		ile = logs[len(logs)-1]
	}

	return ile
}

func (api *oliveTinAPI) ExecutionStatus(ctx ctx.Context, req *connect.Request[apiv1.ExecutionStatusRequest]) (*connect.Response[apiv1.ExecutionStatusResponse], error) {
	res := &apiv1.ExecutionStatusResponse{}

	user := acl.UserFromContext(ctx, req, api.cfg)

	if err := api.checkDashboardAccess(user); err != nil {
		return nil, err
	}

	var ile *executor.InternalLogEntry

	if req.Msg.ExecutionTrackingId != "" {
		ile = getExecutionStatusByTrackingID(api, req.Msg.ExecutionTrackingId)

	} else {
		ile = getMostRecentExecutionStatusById(api, req.Msg.ActionId)
	}

	if ile == nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("execution not found for tracking ID %s or action ID %s", req.Msg.ExecutionTrackingId, req.Msg.ActionId))
	} else {
		res.LogEntry = api.internalLogEntryToPb(ile, user)
	}

	return connect.NewResponse(res), nil
}

func (api *oliveTinAPI) Logout(ctx ctx.Context, req *connect.Request[apiv1.LogoutRequest]) (*connect.Response[apiv1.LogoutResponse], error) {
	user := acl.UserFromContext(ctx, req, api.cfg)

	log.WithFields(log.Fields{
		"username": user.Username,
		"provider": user.Provider,
	}).Info("Logout: User logged out")

	response := connect.NewResponse(&apiv1.LogoutResponse{})

	// Clear the authentication cookie by setting it to expire
	cookie := &http.Cookie{
		Name:     "olivetin-sid-local",
		Value:    "",
		MaxAge:   -1, // This tells the browser to delete the cookie
		HttpOnly: true,
		Path:     "/",
	}
	response.Header().Set("Set-Cookie", cookie.String())

	return response, nil
}

func (api *oliveTinAPI) GetActionBinding(ctx ctx.Context, req *connect.Request[apiv1.GetActionBindingRequest]) (*connect.Response[apiv1.GetActionBindingResponse], error) {
	user := acl.UserFromContext(ctx, req, api.cfg)

	if err := api.checkDashboardAccess(user); err != nil {
		return nil, err
	}

	binding := api.executor.FindBindingByID(req.Msg.BindingId)

	return connect.NewResponse(&apiv1.GetActionBindingResponse{
		Action: buildAction(binding, &DashboardRenderRequest{
			cfg:               api.cfg,
			AuthenticatedUser: user,
			ex:                api.executor,
		}),
	}), nil
}

func (api *oliveTinAPI) GetDashboard(ctx ctx.Context, req *connect.Request[apiv1.GetDashboardRequest]) (*connect.Response[apiv1.GetDashboardResponse], error) {
	user := acl.UserFromContext(ctx, req, api.cfg)

	if err := api.checkDashboardAccess(user); err != nil {
		return nil, err
	}

	dashboardRenderRequest := api.createDashboardRenderRequest(user)

	if api.isDefaultDashboard(req.Msg.Title) {
		return api.buildDefaultDashboardResponse(dashboardRenderRequest)
	}

	return api.buildCustomDashboardResponse(dashboardRenderRequest, req.Msg.Title)
}

func (api *oliveTinAPI) checkDashboardAccess(user *acl.AuthenticatedUser) error {
	if user.IsGuest() && api.cfg.AuthRequireGuestsToLogin {
		return connect.NewError(connect.CodePermissionDenied, fmt.Errorf("guests are not allowed to access the dashboard"))
	}
	return nil
}

func (api *oliveTinAPI) createDashboardRenderRequest(user *acl.AuthenticatedUser) *DashboardRenderRequest {
	return &DashboardRenderRequest{
		AuthenticatedUser: user,
		cfg:               api.cfg,
		ex:                api.executor,
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

func (api *oliveTinAPI) GetLogs(ctx ctx.Context, req *connect.Request[apiv1.GetLogsRequest]) (*connect.Response[apiv1.GetLogsResponse], error) {
	user := acl.UserFromContext(ctx, req, api.cfg)

	if err := api.checkDashboardAccess(user); err != nil {
		return nil, err
	}

	ret := &apiv1.GetLogsResponse{}

	logEntries, pagingResult := api.executor.GetLogTrackingIds(req.Msg.StartOffset, api.cfg.LogHistoryPageSize)

	for _, logEntry := range logEntries {
		action := logEntry.Binding.Action

		if action == nil || acl.IsAllowedLogs(api.cfg, user, action) {
			pbLogEntry := api.internalLogEntryToPb(logEntry, user)

			ret.Logs = append(ret.Logs, pbLogEntry)
		}
	}

	ret.CountRemaining = pagingResult.CountRemaining
	ret.PageSize = pagingResult.PageSize
	ret.TotalCount = pagingResult.TotalCount
	ret.StartOffset = pagingResult.StartOffset

	return connect.NewResponse(ret), nil
}

/*
This function is ONLY a helper for the UI - the arguments are validated properly
on the StartAction -> Executor chain. This is here basically to provide helpful
error messages more quickly before starting the action.
*/
func (api *oliveTinAPI) ValidateArgumentType(ctx ctx.Context, req *connect.Request[apiv1.ValidateArgumentTypeRequest]) (*connect.Response[apiv1.ValidateArgumentTypeResponse], error) {
	err := executor.TypeSafetyCheck("", req.Msg.Value, req.Msg.Type)
	desc := ""

	if err != nil {
		desc = err.Error()
	}

	return connect.NewResponse(&apiv1.ValidateArgumentTypeResponse{
		Valid:       err == nil,
		Description: desc,
	}), nil
}

func (api *oliveTinAPI) WhoAmI(ctx ctx.Context, req *connect.Request[apiv1.WhoAmIRequest]) (*connect.Response[apiv1.WhoAmIResponse], error) {
	user := acl.UserFromContext(ctx, req, api.cfg)

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
	sos := installationinfo.GetSosReport()

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

	jsonstring, _ := json.MarshalIndent(entities.GetAll(), "", "  ")
	fmt.Printf("%s", &jsonstring)

	res.Alert = "Dumping variables has been enabled in the configuration. Please set InsecureAllowDumpVars = false again after you don't need it anymore"

	return connect.NewResponse(res), nil
}

func (api *oliveTinAPI) DumpPublicIdActionMap(ctx ctx.Context, req *connect.Request[apiv1.DumpPublicIdActionMapRequest]) (*connect.Response[apiv1.DumpPublicIdActionMapResponse], error) {
	res := &apiv1.DumpPublicIdActionMapResponse{}
	res.Contents = make(map[string]*apiv1.ActionEntityPair)

	if !api.cfg.InsecureAllowDumpActionMap {
		res.Alert = "Dumping Public IDs is disallowed."

		return connect.NewResponse(res), nil
	}

	api.executor.MapActionIdToBindingLock.RLock()

	for k, v := range api.executor.MapActionIdToBinding {
		res.Contents[k] = &apiv1.ActionEntityPair{
			ActionTitle:  v.Action.Title,
			EntityPrefix: "?",
		}
	}

	api.executor.MapActionIdToBindingLock.RUnlock()

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

	user := acl.UserFromContext(ctx, req, api.cfg)

	if err := api.checkDashboardAccess(user); err != nil {
		return err
	}

	client := &connectedClients{
		channel:           make(chan *apiv1.EventStreamResponse, 10), // Buffered channel to hold Events
		AuthenticatedUser: user,
	}

	log.Infof("EventStream: client connected: %v", client.AuthenticatedUser.Username)

	api.connectedClients = append(api.connectedClients, client)

	// loop over client channel and send events to connectedClient
	for msg := range client.channel {
		log.Debugf("Sending event to client: %v", msg)
		if err := srv.Send(msg); err != nil {
			log.Errorf("Error sending event to client: %v", err)
		}
	}

	log.Infof("EventStream: client disconnected")

	return nil
}

func (api *oliveTinAPI) OnActionMapRebuilt() {
	for _, client := range api.connectedClients {
		select {
		case client.channel <- &apiv1.EventStreamResponse{
			Event: &apiv1.EventStreamResponse_ConfigChanged{
				ConfigChanged: &apiv1.EventConfigChanged{},
			},
		}:
		default:
			log.Warnf("EventStream: client channel is full, dropping message")
		}
	}
}

func (api *oliveTinAPI) OnExecutionStarted(ex *executor.InternalLogEntry) {
	for _, client := range api.connectedClients {
		select {
		case client.channel <- &apiv1.EventStreamResponse{
			Event: &apiv1.EventStreamResponse_ExecutionStarted{
				ExecutionStarted: &apiv1.EventExecutionStarted{
					LogEntry: api.internalLogEntryToPb(ex, client.AuthenticatedUser),
				},
			},
		}:
		default:
			log.Warnf("EventStream: client channel is full, dropping message")
		}
	}
}

func (api *oliveTinAPI) OnExecutionFinished(ex *executor.InternalLogEntry) {
	for _, client := range api.connectedClients {
		select {
		case client.channel <- &apiv1.EventStreamResponse{
			Event: &apiv1.EventStreamResponse_ExecutionFinished{
				ExecutionFinished: &apiv1.EventExecutionFinished{
					LogEntry: api.internalLogEntryToPb(ex, client.AuthenticatedUser),
				},
			},
		}:
		default:
			log.Warnf("EventStream: client channel is full, dropping message")
		}
	}
}

func (api *oliveTinAPI) GetDiagnostics(ctx ctx.Context, req *connect.Request[apiv1.GetDiagnosticsRequest]) (*connect.Response[apiv1.GetDiagnosticsResponse], error) {
	res := &apiv1.GetDiagnosticsResponse{
		SshFoundKey:    installationinfo.Runtime.SshFoundKey,
		SshFoundConfig: installationinfo.Runtime.SshFoundConfig,
	}

	return connect.NewResponse(res), nil
}

func (api *oliveTinAPI) Init(ctx ctx.Context, req *connect.Request[apiv1.InitRequest]) (*connect.Response[apiv1.InitResponse], error) {
	user := acl.UserFromContext(ctx, req, api.cfg)

	loginRequired := user.IsGuest() && api.cfg.AuthRequireGuestsToLogin

	res := &apiv1.InitResponse{
		ShowFooter:                api.cfg.ShowFooter,
		ShowNavigation:            api.cfg.ShowNavigation,
		ShowNewVersions:           api.cfg.ShowNewVersions,
		AvailableVersion:          installationinfo.Runtime.AvailableVersion,
		CurrentVersion:            installationinfo.Build.Version,
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
	}

	return connect.NewResponse(res), nil
}

func (api *oliveTinAPI) buildRootDashboards(user *acl.AuthenticatedUser, dashboards []*config.DashboardComponent) []string {
	var rootDashboards []string
	dashboardRenderRequest := api.createDashboardRenderRequest(user)

	api.addDefaultDashboardIfNeeded(&rootDashboards, dashboardRenderRequest)
	api.addCustomDashboards(&rootDashboards, dashboards, dashboardRenderRequest)

	return rootDashboards
}

func (api *oliveTinAPI) addDefaultDashboardIfNeeded(rootDashboards *[]string, rr *DashboardRenderRequest) {
	defaultDashboard := buildDefaultDashboard(rr)
	if defaultDashboard != nil && len(defaultDashboard.Contents) > 0 {
		log.Infof("defaultDashboard: %+v", defaultDashboard.Contents)
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

	for _, provider := range cfg.AuthOAuth2Providers {
		publicProviders = append(publicProviders, &apiv1.OAuth2Provider{
			Title: provider.Title,
			Url:   provider.AuthUrl,
			Icon:  provider.Icon,
		})
	}

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
	for _, client := range api.connectedClients {
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
			log.Warnf("EventStream: client channel is full, dropping message")
		}
	}
}

func (api *oliveTinAPI) GetEntities(ctx ctx.Context, req *connect.Request[apiv1.GetEntitiesRequest]) (*connect.Response[apiv1.GetEntitiesResponse], error) {
	user := acl.UserFromContext(ctx, req, api.cfg)

	if err := api.checkDashboardAccess(user); err != nil {
		return nil, err
	}

	res := &apiv1.GetEntitiesResponse{
		EntityDefinitions: make([]*apiv1.EntityDefinition, 0),
	}

	for name, entityInstances := range entities.GetEntities() {
		def := &apiv1.EntityDefinition{
			Title:            name,
			UsedOnDashboards: findDashboardsForEntity(name, api.cfg.Dashboards),
		}

		for _, e := range entityInstances {
			entity := &apiv1.Entity{
				Title:     e.Title,
				UniqueKey: e.UniqueKey,
				Type:      name,
			}

			def.Instances = append(def.Instances, entity)
		}

		res.EntityDefinitions = append(res.EntityDefinitions, def)
	}

	return connect.NewResponse(res), nil
}

func findDashboardsForEntity(entityTitle string, dashboards []*config.DashboardComponent) []string {
	var foundDashboards []string

	findEntityInComponents(entityTitle, "", dashboards, &foundDashboards)

	return foundDashboards
}

func findEntityInComponents(entityTitle string, parentTitle string, components []*config.DashboardComponent, foundDashboards *[]string) {
	for _, component := range components {
		if component.Entity == entityTitle {
			*foundDashboards = append(*foundDashboards, parentTitle)
		}

		if len(component.Contents) > 0 {
			findEntityInComponents(entityTitle, component.Title, component.Contents, foundDashboards)
		}
	}
}

func (api *oliveTinAPI) GetEntity(ctx ctx.Context, req *connect.Request[apiv1.GetEntityRequest]) (*connect.Response[apiv1.Entity], error) {
	user := acl.UserFromContext(ctx, req, api.cfg)

	if err := api.checkDashboardAccess(user); err != nil {
		return nil, err
	}

	res := &apiv1.Entity{}

	instances := entities.GetEntityInstances(req.Msg.Type)

	log.Infof("msg: %+v", req.Msg)

	if len(instances) == 0 {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("entity type %s not found", req.Msg.Type))
	}

	if entity, ok := instances[req.Msg.UniqueKey]; !ok {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("entity with unique key %s not found in type %s", req.Msg.UniqueKey, req.Msg.Type))
	} else {
		res.Title = entity.Title

		return connect.NewResponse(res), nil
	}
}

func (api *oliveTinAPI) RestartAction(ctx ctx.Context, req *connect.Request[apiv1.RestartActionRequest]) (*connect.Response[apiv1.StartActionResponse], error) {
	ret := &apiv1.StartActionResponse{
		ExecutionTrackingId: req.Msg.ExecutionTrackingId,
	}

	var execReqLogEntry *executor.InternalLogEntry

	execReqLogEntry, found := api.executor.GetLog(req.Msg.ExecutionTrackingId)

	if !found {
		log.Warnf("Restarting execution request not possible - not found by tracking ID: %v", req.Msg.ExecutionTrackingId)
		return connect.NewResponse(ret), nil
	}

	log.Warnf("Restarting execution request by tracking ID: %v", req.Msg.ExecutionTrackingId)

	action := execReqLogEntry.Binding.Action

	if action == nil {
		log.Warnf("Restarting execution request not possible - action not found: %v", execReqLogEntry.ActionTitle)
		return connect.NewResponse(ret), nil
	}

	return api.StartAction(ctx, &connect.Request[apiv1.StartActionRequest]{
		Msg: &apiv1.StartActionRequest{
			// FIXME
			UniqueTrackingId: req.Msg.ExecutionTrackingId,
		},
	})
}

func newServer(ex *executor.Executor) *oliveTinAPI {
	server := oliveTinAPI{}
	server.cfg = ex.Cfg
	server.executor = ex

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
