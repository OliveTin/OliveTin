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
)

var (
	cfg *config.Config
	ex  = executor.DefaultExecutor()
)

type oliveTinAPI struct {
	pb.UnimplementedOliveTinApiServer
}

func (api *oliveTinAPI) StartAction(ctx ctx.Context, req *pb.StartActionRequest) (*pb.StartActionResponse, error) {
	args := make(map[string]string)

	log.Debugf("SA %v", req)

	for _, arg := range req.Arguments {
		args[arg.Name] = arg.Value
	}

	execReq := executor.ExecutionRequest{
		ActionName: req.ActionName,
		Arguments:  args,
		User:       acl.UserFromContext(ctx),
		Cfg:        cfg,
	}

	return ex.ExecRequest(&execReq), nil
}

func (api *oliveTinAPI) GetDashboardComponents(ctx ctx.Context, req *pb.GetDashboardComponentsRequest) (*pb.GetDashboardComponentsResponse, error) {
	user := acl.UserFromContext(ctx)

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

	for _, logEntry := range ex.Logs {
		ret.Logs = append(ret.Logs, &pb.LogEntry{
			ActionTitle: logEntry.ActionTitle,
			ActionIcon:  logEntry.ActionIcon,
			Datetime:    logEntry.Datetime,
			Stdout:      logEntry.Stdout,
			Stderr:      logEntry.Stderr,
			TimedOut:    logEntry.TimedOut,
			ExitCode:    logEntry.ExitCode,
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

// Start will start the GRPC API.
func Start(globalConfig *config.Config) {
	cfg = globalConfig

	lis, err := net.Listen("tcp", cfg.ListenAddressGrpcActions)

	if err != nil {
		log.Fatalf("Failed to listen - %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterOliveTinApiServer(grpcServer, newServer())

	err = grpcServer.Serve(lis)

	if err != nil {
		log.Fatalf("Could not start gRPC Server - %v", err)
	}
}

func newServer() *oliveTinAPI {
	server := oliveTinAPI{}
	return &server
}
