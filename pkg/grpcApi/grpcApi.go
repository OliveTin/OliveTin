package grpcApi;

import (
	"google.golang.org/grpc"
	pb "github.com/jamesread/OliveTin/gen/grpc"
	"fmt"
	"net"
	log "github.com/sirupsen/logrus"
	ctx "context"
)

type OliveTinApi struct {

}

func (api *OliveTinApi) StartAction(ctx ctx.Context, req *pb.StartActionRequest) (*pb.StartActionResponse, error) {
	res := &pb.StartActionResponse{}

	log.WithFields(log.Fields{
		"actionName": req.ActionName,
	}).Infof("StartAction")

	return res, nil
}

func (api *OliveTinApi) GetButtons(ctx ctx.Context, req *pb.GetButtonsRequest) (*pb.GetButtonsResponse, error) {
	res := &pb.GetButtonsResponse{};

	btn1 := pb.ActionButton {
		Title: "foo",
		Icon: "&#x1F1E6",
	};

	btn2 := pb.ActionButton {
		Title: "bar",
		Icon: "&#x1F1E6",
	};

	res.Actions = append(res.Actions, &btn1);
	res.Actions = append(res.Actions, &btn2);

	return res, nil
}

func Start() {
	port := 1337;
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port));

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
