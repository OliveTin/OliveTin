package httpservers

/*
The REST API actually has very few tests, as the "real" API behind OliveTin
is is implemented as a gRPC in /internal/grpc. The REST API therefore only
handles HTTP specific stuff like authentication cookies and JWT parsing.
*/

import (
	"fmt"
	"github.com/OliveTin/OliveTin/internal/cors"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"net"
	"net/http"
	"testing"
)

func setupTestingServer(mux *runtime.ServeMux, t *testing.T) *http.Server {
	lis, err := net.Listen("tcp", ":1337")

	if err != nil || lis == nil {
		t.Errorf("Could not listen %v %v", err, lis)
		return nil
	}

	srv := &http.Server{Handler: cors.AllowCors(mux)}

	go startTestingServer(lis, srv, t)

	return srv
}

func startTestingServer(lis net.Listener, srv *http.Server, t *testing.T) {
	if srv == nil {
		t.Errorf("srv is nil. Could not listen")
		return
	}

	go func() {
		if err := srv.Serve(lis); err != nil {
			fmt.Printf("couldn't start server: %+v", err)
		}
	}()
}

func newReq(path string) (*http.Request, *http.Client) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:1337/%v", path), nil)

	return req, client
}
