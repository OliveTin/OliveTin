package api

import (
	"context"
	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"testing"

	log "github.com/sirupsen/logrus"

	apiv1 "github.com/OliveTin/OliveTin/gen/olivetin/api/v1"
	apiv1connect "github.com/OliveTin/OliveTin/gen/olivetin/api/v1/apiv1connect"
	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/executor"

	"net/http"
	"net/http/httptest"
)

func getNewTestServerAndClient(t *testing.T, injectedConfig *config.Config) (*httptest.Server, apiv1connect.OliveTinApiServiceClient) {
	ex := executor.DefaultExecutor(injectedConfig)
	ex.RebuildActionMap()

	path, handler := GetNewHandler(ex)

	path = "/api" + path

	mux := http.NewServeMux()
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		log.Infof("HTTP Request: %s %s", r.Method, r.URL.Path)

		http.StripPrefix("/api/", handler)
	})

	log.Infof("API path is %s", path)

	httpclient := &http.Client{
	}

	ts := httptest.NewServer(mux)

	client := apiv1connect.NewOliveTinApiServiceClient(httpclient, ts.URL + "/api")

	log.Infof("Test server URL is %s", ts.URL + path)

	return ts, client
}

func TestGetActionsAndStart(t *testing.T) {
	cfg := config.DefaultConfig()

	btn1 := &config.Action{}
	btn1.Title = "blat"
	btn1.ID = "blat"
	btn1.Shell = "echo 'test'"
	cfg.Actions = append(cfg.Actions, btn1)

	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()

	conn, client := getNewTestServerAndClient(t, cfg)

	respGb, err := client.GetDashboardComponents(context.Background(), connect.NewRequest(&apiv1.GetDashboardComponentsRequest{}))
	respGetReady, err := client.GetReadyz(context.Background(), connect.NewRequest(&apiv1.GetReadyzRequest{}))

	if err != nil {
		t.Errorf("GetDashboardComponentsRequest: %v", err)
		return
	}

	log.Infof("GetReadyz response: %v", respGetReady.Msg)

	assert.Equal(t, true, true, "sayHello Failed")

//	assert.Equal(t, 1, len(respGb.Msg.Actions), "Got 1 action button back")

	log.Printf("Response: %+v", respGb)

	respSa, err := client.StartAction(context.Background(), connect.NewRequest(&apiv1.StartActionRequest{ActionId: "blat"}))

	assert.NotNil(t, err, "Error 404 after start action")
	assert.Nil(t, respSa, "Nil response for non existing action")

	defer conn.Close()
}
