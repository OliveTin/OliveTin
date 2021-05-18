package grpcapi

// Thank you: https://stackoverflow.com/questions/42102496/testing-a-grpc-service

import (
	"net"
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/grpc"

	log "github.com/sirupsen/logrus"

	config "github.com/jamesread/OliveTin/internal/config"
	pb "github.com/jamesread/OliveTin/gen/grpc"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func init() {
    lis = bufconn.Listen(bufSize)
    s := grpc.NewServer()
    pb.RegisterOliveTinApiServer(s, newServer())

    go func() {
        if err := s.Serve(lis); err != nil {
            log.Fatalf("Server exited with error: %v", err)
        }
    }()
}

func bufDialer(context.Context, string) (net.Conn, error) {
    return lis.Dial()
}

func getNewTestServerAndClient(t *testing.T, injectedConfig *config.Config) (*grpc.ClientConn, pb.OliveTinApiClient) {
	cfg = injectedConfig

    ctx := context.Background()

    conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())

    if err != nil {
        t.Fatalf("Failed to dial bufnet: %v", err)
    }

    client := pb.NewOliveTinApiClient(conn)

	return conn, client
}

func TestGetButtonsAndStart(t *testing.T) {
	cfg = config.DefaultConfig();
	btn1 := config.ActionButton{}
	btn1.Title = "blat"
	btn1.Shell = "echo 'test'"
	cfg.ActionButtons = append(cfg.ActionButtons, btn1);

	conn, client := getNewTestServerAndClient(t, cfg)

    respGb, err := client.GetButtons(context.Background(), &pb.GetButtonsRequest{})

	if err != nil {
		t.Errorf("GetButtons: %v", err)
	}

	assert.Equal(t, true, true, "sayHello Failed")

	assert.Equal(t, 1, len(respGb.Actions), "Got 1 action button back")

    log.Printf("Response: %+v", respGb)

	respSa, err := client.StartAction(context.Background(), &pb.StartActionRequest{ActionName: "blat"})

	assert.Nil(t, err, "Empty err after start action")
	assert.NotNil(t, respSa, "Empty err after start action")

    defer conn.Close()
}
