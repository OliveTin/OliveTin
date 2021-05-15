package grpcApi;

import (
	"google.golang.org/grpc"
	pb "github.com/jamesread/OliveTin/gen/grpc"
	"net"
	log "github.com/sirupsen/logrus"
	ctx "context"

	config "github.com/jamesread/OliveTin/pkg/config"
	executor "github.com/jamesread/OliveTin/pkg/executor"
)

var (
	cfg *config.Config;
)


type OliveTinApi struct {

}

func (api *OliveTinApi) StartAction(ctx ctx.Context, req *pb.StartActionRequest) (*pb.StartActionResponse, error) {
	return executor.ExecAction(req.ActionName), nil
}

func (api *OliveTinApi) GetButtons(ctx ctx.Context, req *pb.GetButtonsRequest) (*pb.GetButtonsResponse, error) {
	res := &pb.GetButtonsResponse{};

	for _, action := range cfg.ActionButtons {
		btn := pb.ActionButton {
			Title: action.Title,
			Icon: lookupHtmlIcon(action.Icon),
		}

		res.Actions = append(res.Actions, &btn);
	}

	log.Infof("getButtons: %v", res)

	return res, nil
}

func Start(listenAddress string, globalConfig *config.Config) {
	cfg = globalConfig

	lis, err := net.Listen("tcp", listenAddress);

	if err != nil {
		log.Fatalf("Failed to listen - %v", err);
	}

	grpcServer := grpc.NewServer();
	pb.RegisterOliveTinApiServer(grpcServer, newServer());
	grpcServer.Serve(lis)
}

func newServer() (*OliveTinApi) {
	server := OliveTinApi {};
	return &server;
}
