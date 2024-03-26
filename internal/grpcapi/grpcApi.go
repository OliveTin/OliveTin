package grpcapi

import (
	ctx "context"
	pb "github.com/OliveTin/OliveTin/gen/grpc"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc"

	"errors"
	"net"
	"sort"

	acl "github.com/OliveTin/OliveTin/internal/acl"
	config "github.com/OliveTin/OliveTin/internal/config"
	executor "github.com/OliveTin/OliveTin/internal/executor"
	installationinfo "github.com/OliveTin/OliveTin/internal/installationinfo"
	sv "github.com/OliveTin/OliveTin/internal/stringvariables"
)

var (
	cfg *config.Config
)

type oliveTinAPI struct {
	// Uncomment this if you want to allow undefined methods during dev.
	//	pb.UnimplementedOliveTinApiServiceServer

	executor *executor.Executor
}

func (api *oliveTinAPI) StartAction(ctx ctx.Context, req *pb.StartActionRequest) (*pb.StartActionResponse, error) {
	args := make(map[string]string)

	for _, arg := range req.Arguments {
		args[arg.Name] = arg.Value
	}

	pair := publicActionIdToActionMap[req.ActionId]

	execReq := executor.ExecutionRequest{
		Action:            pair.Action,
		EntityPrefix:      pair.EntityPrefix,
		TrackingID:        req.UniqueTrackingId,
		Arguments:         args,
		AuthenticatedUser: acl.UserFromContext(ctx, cfg),
		Cfg:               cfg,
	}

	api.executor.ExecRequest(&execReq)

	return &pb.StartActionResponse{
		ExecutionTrackingId: execReq.TrackingID,
	}, nil
}

func (api *oliveTinAPI) StartActionAndWait(ctx ctx.Context, req *pb.StartActionAndWaitRequest) (*pb.StartActionAndWaitResponse, error) {
	args := make(map[string]string)

	execReq := executor.ExecutionRequest{
		Action:            findActionByPublicID(req.ActionId),
		TrackingID:        uuid.NewString(),
		Arguments:         args,
		AuthenticatedUser: acl.UserFromContext(ctx, cfg),
		Cfg:               cfg,
	}

	wg, _ := api.executor.ExecRequest(&execReq)
	wg.Wait()

	internalLogEntry, ok := api.executor.Logs[execReq.TrackingID]

	if ok {
		return &pb.StartActionAndWaitResponse{
			LogEntry: internalLogEntryToPb(internalLogEntry),
		}, nil
	} else {
		return nil, errors.New("Execution not found!")
	}
}

func (api *oliveTinAPI) StartActionByGet(ctx ctx.Context, req *pb.StartActionByGetRequest) (*pb.StartActionByGetResponse, error) {
	args := make(map[string]string)

	execReq := executor.ExecutionRequest{
		Action:            findActionByPublicID(req.ActionId),
		TrackingID:        uuid.NewString(),
		Arguments:         args,
		AuthenticatedUser: acl.UserFromContext(ctx, cfg),
		Cfg:               cfg,
	}

	_, uniqueTrackingId := api.executor.ExecRequest(&execReq)

	return &pb.StartActionByGetResponse{
		ExecutionTrackingId: uniqueTrackingId,
	}, nil
}

func (api *oliveTinAPI) StartActionByGetAndWait(ctx ctx.Context, req *pb.StartActionByGetAndWaitRequest) (*pb.StartActionByGetAndWaitResponse, error) {
	args := make(map[string]string)

	execReq := executor.ExecutionRequest{
		Action:            findActionByPublicID(req.ActionId),
		TrackingID:        uuid.NewString(),
		Arguments:         args,
		AuthenticatedUser: acl.UserFromContext(ctx, cfg),
		Cfg:               cfg,
	}

	wg, _ := api.executor.ExecRequest(&execReq)
	wg.Wait()

	internalLogEntry, ok := api.executor.Logs[execReq.TrackingID]

	if ok {
		return &pb.StartActionByGetAndWaitResponse{
			LogEntry: internalLogEntryToPb(internalLogEntry),
		}, nil
	} else {
		return nil, errors.New("Execution not found!")
	}
}

func internalLogEntryToPb(logEntry *executor.InternalLogEntry) *pb.LogEntry {
	return &pb.LogEntry{
		ActionTitle:         logEntry.ActionTitle,
		ActionIcon:          logEntry.ActionIcon,
		ActionId:            logEntry.ActionId,
		DatetimeStarted:     logEntry.DatetimeStarted,
		DatetimeFinished:    logEntry.DatetimeFinished,
		Stdout:              logEntry.Stdout,
		Stderr:              logEntry.Stderr,
		TimedOut:            logEntry.TimedOut,
		Blocked:             logEntry.Blocked,
		ExitCode:            logEntry.ExitCode,
		Tags:                logEntry.Tags,
		ExecutionTrackingId: logEntry.ExecutionTrackingID,
		ExecutionStarted:    logEntry.ExecutionStarted,
		ExecutionFinished:   logEntry.ExecutionFinished,
	}
}

func getExecutionStatusByTrackingID(api *oliveTinAPI, executionTrackingId string) *executor.InternalLogEntry {
	logEntry, ok := api.executor.Logs[executionTrackingId]

	if !ok {
		return nil
	}

	return logEntry
}

func getMostRecentExecutionStatusById(api *oliveTinAPI, actionId string) *executor.InternalLogEntry {
	var ile *executor.InternalLogEntry

	for _, candidateLe := range api.executor.Logs {
		if actionId == candidateLe.ActionId {
			ile = candidateLe
		}
	}

	return ile
}

func (api *oliveTinAPI) ExecutionStatus(ctx ctx.Context, req *pb.ExecutionStatusRequest) (*pb.ExecutionStatusResponse, error) {
	res := &pb.ExecutionStatusResponse{}

	var ile *executor.InternalLogEntry

	if req.ExecutionTrackingId != "" {
		ile = getExecutionStatusByTrackingID(api, req.ExecutionTrackingId)
	} else {
		ile = getMostRecentExecutionStatusById(api, req.ActionId)
	}

	res.LogEntry = internalLogEntryToPb(ile)

	return res, nil
}

/**
func (api *oliveTinAPI) WatchExecution(req *pb.WatchExecutionRequest, srv pb.OliveTinApi_WatchExecutionServer) error {
	log.Infof("Watch")

	if logEntry, ok := api.executor.Logs[req.ExecutionUuid]; !ok {
		log.Errorf("Execution not found: %v", req.ExecutionUuid)

		return nil
	} else {
		if logEntry.ExecutionStarted {
			for !logEntry.ExecutionCompleted {
				tmp := make([]byte, 256)

				red, err := io.ReadAtLeast(logEntry.StdoutBuffer, tmp, 1)

				log.Infof("%v %v", red, err)

				srv.Send(&pb.WatchExecutionUpdate{
					Update: string(tmp),
				})
			}
		}

		return nil
	}
}
*/

func (api *oliveTinAPI) GetDashboardComponents(ctx ctx.Context, req *pb.GetDashboardComponentsRequest) (*pb.GetDashboardComponentsResponse, error) {
	user := acl.UserFromContext(ctx, cfg)

	res := actionsCfgToPb(cfg.Actions, user)

	if len(res.Actions) == 0 {
		log.Warn("Zero actions found - check that you have some actions defined, with a view permission")
	}

	log.Tracef("GetDashboardComponents: %v", res)

	dashboardCfgToPb(res, cfg.Dashboards)

	return res, nil
}

func (api *oliveTinAPI) GetLogs(ctx ctx.Context, req *pb.GetLogsRequest) (*pb.GetLogsResponse, error) {
	ret := &pb.GetLogsResponse{}

	// TODO Limit to 10 entries or something to prevent browser lag.

	for trackingId, logEntry := range api.executor.Logs {
		pbLogEntry := internalLogEntryToPb(logEntry)
		pbLogEntry.ExecutionTrackingId = trackingId

		ret.Logs = append(ret.Logs, pbLogEntry)
	}

	sorter := func(i, j int) bool {
		return ret.Logs[i].DatetimeStarted < ret.Logs[j].DatetimeStarted
	}

	sort.Slice(ret.Logs, sorter)

	return ret, nil
}

/*
This function is ONLY a helper for the UI - the arguments are validated properly
on the StartAction -> Executor chain. This is here basically to provide helpful
error messages more quickly before starting the action.
*/
func (api *oliveTinAPI) ValidateArgumentType(ctx ctx.Context, req *pb.ValidateArgumentTypeRequest) (*pb.ValidateArgumentTypeResponse, error) {
	err := executor.TypeSafetyCheck("", req.Value, req.Type)
	desc := ""

	if err != nil {
		desc = err.Error()
	}

	return &pb.ValidateArgumentTypeResponse{
		Valid:       err == nil,
		Description: desc,
	}, nil
}

func (api *oliveTinAPI) WhoAmI(ctx ctx.Context, req *pb.WhoAmIRequest) (*pb.WhoAmIResponse, error) {
	user := acl.UserFromContext(ctx, cfg)

	res := &pb.WhoAmIResponse{
		AuthenticatedUser: user.Username,
	}

	log.Warnf("usergroup: %v", user.Usergroup)

	return res, nil
}

func (api *oliveTinAPI) SosReport(ctx ctx.Context, req *pb.SosReportRequest) (*httpbody.HttpBody, error) {
	sos := installationinfo.GetSosReport()

	if !cfg.InsecureAllowDumpSos {
		log.Info(sos)
		sos = "Your SOS Report has been logged to OliveTin logs.\n\nIf you are in a safe network, you can temporarily set `insecureAllowDumpSos: true` in your config.yaml, restart OliveTin, and refresh this page - it will put the output directly in the browser."
	}

	ret := &httpbody.HttpBody{
		ContentType: "text/plain",
		Data:        []byte(sos),
	}

	return ret, nil
}

func (api *oliveTinAPI) DumpVars(ctx ctx.Context, req *pb.DumpVarsRequest) (*pb.DumpVarsResponse, error) {
	res := &pb.DumpVarsResponse{}

	if !cfg.InsecureAllowDumpVars {
		res.Alert = "Dumping variables is not allowed by default because it is insecure."

		return res, nil
	}

	res.Alert = "Dumping variables has been enabled in the configuration. Please set InsecureAllowDumpVars = false again after you don't need it anymore"
	res.Contents = sv.GetAll()

	return res, nil
}

func (api *oliveTinAPI) DumpPublicIdActionMap(ctx ctx.Context, req *pb.DumpPublicIdActionMapRequest) (*pb.DumpPublicIdActionMapResponse, error) {
	res := &pb.DumpPublicIdActionMapResponse{}
	res.Contents = make(map[string]*pb.ActionEntityPair)

	if !cfg.InsecureAllowDumpActionMap {
		res.Alert = "Dumping Public IDs is disallowed."

		return res, nil
	}

	for k, v := range publicActionIdToActionMap {
		res.Contents[k] = &pb.ActionEntityPair{
			ActionTitle:  v.Action.Title,
			EntityPrefix: v.EntityPrefix,
		}
	}

	res.Alert = "Dumping variables has been enabled in the configuration. Please set InsecureAllowDumpActionMap = false again after you don't need it anymore"

	return res, nil
}

func (api *oliveTinAPI) GetReadyz(ctx ctx.Context, req *pb.GetReadyzRequest) (*pb.GetReadyzResponse, error) {
	res := &pb.GetReadyzResponse{
		Status: "OK",
	}

	return res, nil
}

// Start will start the GRPC API.
func Start(globalConfig *config.Config, ex *executor.Executor) {
	cfg = globalConfig

	lis, err := net.Listen("tcp", cfg.ListenAddressGrpcActions)

	if err != nil {
		log.Fatalf("Failed to listen - %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterOliveTinApiServiceServer(grpcServer, newServer(ex))

	err = grpcServer.Serve(lis)

	if err != nil {
		log.Fatalf("Could not start gRPC Server - %v", err)
	}
}

func newServer(ex *executor.Executor) *oliveTinAPI {
	server := oliveTinAPI{}
	server.executor = ex
	return &server
}
