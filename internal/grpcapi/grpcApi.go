package grpcapi

import (
	ctx "context"
	pb "github.com/jamesread/OliveTin/gen/grpc"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"

	config "github.com/jamesread/OliveTin/internal/config"
	executor "github.com/jamesread/OliveTin/internal/executor"
)

var (
	cfg *config.Config
)

type oliveTinAPI struct {
}

func (api *oliveTinAPI) StartAction(ctx ctx.Context, req *pb.StartActionRequest) (*pb.StartActionResponse, error) {
	return executor.ExecAction(cfg, req.ActionName), nil
}

func (api *oliveTinAPI) GetButtons(ctx ctx.Context, req *pb.GetButtonsRequest) (*pb.GetButtonsResponse, error) {
	res := &pb.GetButtonsResponse{}

	for _, action := range cfg.ActionButtons {
		btn := pb.ActionButton{
			Title: action.Title,
			Icon:  lookupHTMLIcon(action.Icon),
		}

		res.Actions = append(res.Actions, &btn)
	}

	log.Infof("getButtons: %v", res)

	return res, nil
}

func Start(listenAddress string, globalConfig *config.Config) {
	cfg = globalConfig

	lis, err := net.Listen("tcp", listenAddress)

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
