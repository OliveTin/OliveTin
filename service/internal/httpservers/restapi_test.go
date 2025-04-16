package httpservers

/*
The REST API actually has very few tests, as the "real" API behind OliveTin
is is implemented as a gRPC in /internal/grpc. The REST API therefore only
handles HTTP specific stuff like authentication cookies and JWT parsing.
*/

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/OliveTin/OliveTin/internal/cors"
)

func setupTestingServer(mux http.Handler) *httptest.Server {
	return httptest.NewServer(cors.AllowCors(mux))
}

func newReq(URL string, path string) (*http.Request, *http.Client) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", fmt.Sprintf("%v/%v", URL, path), nil)

	return req, client
}
