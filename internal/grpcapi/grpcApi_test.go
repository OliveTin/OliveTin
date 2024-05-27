package grpcapi

// Thank you: https://stackoverflow.com/questions/42102496/testing-a-grpc-service

import (
	"context"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"net"
	"testing"

	log "github.com/sirupsen/logrus"

	pb "github.com/OliveTin/OliveTin/gen/grpc"
	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/executor"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func initServer(cfg *config.Config) *executor.Executor {
	ex := executor.DefaultExecutor(cfg)

	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	pb.RegisterOliveTinApiServiceServer(s, newServer(ex))

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()

	return ex
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func getNewTestServerAndClient(t *testing.T, injectedConfig *config.Config) (*grpc.ClientConn, pb.OliveTinApiServiceClient) {
	cfg = injectedConfig

	ctx := context.Background()

	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())

	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}

	client := pb.NewOliveTinApiServiceClient(conn)

	return conn, client
}

func TestGetActionsAndStart(t *testing.T) {
	cfg = config.DefaultConfig()

	ex := initServer(cfg)

	btn1 := &config.Action{}
	btn1.Title = "blat"
	btn1.ID = "blat"
	btn1.Shell = "echo 'test'"
	cfg.Actions = append(cfg.Actions, btn1)

	ex.RebuildActionMap()

	conn, client := getNewTestServerAndClient(t, cfg)

	respGb, err := client.GetDashboardComponents(context.Background(), &pb.GetDashboardComponentsRequest{})

	if err != nil {
		t.Errorf("GetDashboardComponentsRequest: %v", err)
	}

	assert.Equal(t, true, true, "sayHello Failed")

	assert.Equal(t, 1, len(respGb.Actions), "Got 1 action button back")

	log.Printf("Response: %+v", respGb)

	respSa, err := client.StartAction(context.Background(), &pb.StartActionRequest{ActionId: "blat"})

	assert.Nil(t, err, "Empty err after start action")
	assert.NotNil(t, respSa, "Empty err after start action")

	defer conn.Close()
}
