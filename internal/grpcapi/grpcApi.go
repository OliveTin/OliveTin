package grpcapi

import (
	ctx "context"
	pb "github.com/OliveTin/OliveTin/gen/grpc"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"net"

	acl "github.com/OliveTin/OliveTin/internal/acl"
	config "github.com/OliveTin/OliveTin/internal/config"
	executor "github.com/OliveTin/OliveTin/internal/executor"
	installationinfo "github.com/OliveTin/OliveTin/internal/installationinfo"
)

var (
	cfg *config.Config
)

type oliveTinAPI struct {
	pb.UnimplementedOliveTinApiServer

	executor *executor.Executor
}

func (api *oliveTinAPI) StartAction(ctx ctx.Context, req *pb.StartActionRequest) (*pb.StartActionResponse, error) {
	args := make(map[string]string)

	log.Debugf("SA %v", req)

	for _, arg := range req.Arguments {
		args[arg.Name] = arg.Value
	}

	execReq := executor.ExecutionRequest{
		ActionName:        req.ActionName,
		Arguments:         args,
		AuthenticatedUser: acl.UserFromContext(ctx, cfg),
		Cfg:               cfg,
	}

	return api.executor.ExecRequest(&execReq), nil
}

func (api *oliveTinAPI) GetDashboardComponents(ctx ctx.Context, req *pb.GetDashboardComponentsRequest) (*pb.GetDashboardComponentsResponse, error) {
	user := acl.UserFromContext(ctx, cfg)

	res := actionsCfgToPb(cfg.Actions, user)

	if len(res.Actions) == 0 {
		log.Warn("Zero actions found - check that you have some actions defined, with a view permission")
	}

	log.Debugf("GetDashboardComponents: %v", res)

	return res, nil
}

func (api *oliveTinAPI) GetLogs(ctx ctx.Context, req *pb.GetLogsRequest) (*pb.GetLogsResponse, error) {
	ret := &pb.GetLogsResponse{}

	// TODO Limit to 10 entries or something to prevent browser lag.

	for _, logEntry := range api.executor.Logs {
		ret.Logs = append(ret.Logs, &pb.LogEntry{
			ActionTitle: logEntry.ActionTitle,
			ActionIcon:  logEntry.ActionIcon,
			Datetime:    logEntry.Datetime,
			Stdout:      logEntry.Stdout,
			Stderr:      logEntry.Stderr,
			TimedOut:    logEntry.TimedOut,
			ExitCode:    logEntry.ExitCode,
			Tags:        logEntry.Tags,
		})
	}

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

func (api *oliveTinAPI) SosReport(ctx ctx.Context, req *pb.SosReportRequest) (*pb.SosReportResponse, error) {
	res := &pb.SosReportResponse{
		Alert: "Your SOS Report has been logged to OliveTin logs.",
	}

	log.Infof("\n" + installationinfo.GetSosReport())

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
	pb.RegisterOliveTinApiServer(grpcServer, newServer(ex))

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
