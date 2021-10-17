package grpcapi

import (
	ctx "context"
	"crypto/md5"
	"fmt"
	pb "github.com/jamesread/OliveTin/gen/grpc"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"

	acl "github.com/jamesread/OliveTin/internal/acl"
	config "github.com/jamesread/OliveTin/internal/config"
	executor "github.com/jamesread/OliveTin/internal/executor"
)

var (
	cfg *config.Config
	ex  = executor.Executor{}
)

type oliveTinAPI struct {
	pb.UnimplementedOliveTinApiServer
}

func (api *oliveTinAPI) StartAction(ctx ctx.Context, req *pb.StartActionRequest) (*pb.StartActionResponse, error) {
	actualAction, err := executor.FindAction(cfg, req.ActionName)

	if err != nil {
		log.Errorf("Error finding action %s, %s", err, req.ActionName)

		return &pb.StartActionResponse{
			LogEntry: nil,
		}, nil
	}

	user := acl.UserFromContext(ctx)

	if !acl.IsAllowedExec(cfg, user, actualAction) {
		return &pb.StartActionResponse{}, nil

	}

	return ex.ExecAction(cfg, acl.UserFromContext(ctx), actualAction), nil
}

func (api *oliveTinAPI) GetButtons(ctx ctx.Context, req *pb.GetButtonsRequest) (*pb.GetButtonsResponse, error) {
	res := &pb.GetButtonsResponse{}

	user := acl.UserFromContext(ctx)

	for _, action := range cfg.ActionButtons {
		if !acl.IsAllowedView(cfg, user, &action) {
			continue
		}

		btn := pb.ActionButton{
			Id:      fmt.Sprintf("%x", md5.Sum([]byte(action.Title))),
			Title:   action.Title,
			Icon:    lookupHTMLIcon(action.Icon),
			CanExec: acl.IsAllowedExec(cfg, user, &action),
		}

		for _, cfgArg := range action.Arguments {
			pbArg := pb.ActionArgument {
				Label: cfgArg.Label,
			}

			btn.Arguments = append(btn.Arguments, &pbArg)
		}

		res.Actions = append(res.Actions, &btn)
	}

	if len(res.Actions) == 0 {
		log.Warn("Zero actions found - check that you have some actions defined, with a view permission")
	}

	log.Debugf("getButtons: %v", res)

	return res, nil
}

func (api *oliveTinAPI) GetLogs(ctx ctx.Context, req *pb.GetLogsRequest) (*pb.GetLogsResponse, error) {
	ret := &pb.GetLogsResponse{}

	// TODO Limit to 10 entries or something to prevent browser lag.

	for _, logEntry := range ex.Logs {
		ret.Logs = append(ret.Logs, &pb.LogEntry{
			ActionTitle: logEntry.ActionTitle,
			Datetime:    logEntry.Datetime,
			Stdout:      logEntry.Stdout,
			Stderr:      logEntry.Stderr,
			TimedOut:    logEntry.TimedOut,
			ExitCode:    logEntry.ExitCode,
		})
	}

	return ret, nil
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
	grpcServer.Serve(lis)
}

func newServer() *oliveTinAPI {
	server := oliveTinAPI{}
	return &server
}
