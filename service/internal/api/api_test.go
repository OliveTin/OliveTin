package api

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"

	log "github.com/sirupsen/logrus"

	apiv1 "github.com/OliveTin/OliveTin/gen/olivetin/api/v1"
	apiv1connect "github.com/OliveTin/OliveTin/gen/olivetin/api/v1/apiv1connect"
	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/executor"

	"net/http"
	"net/http/httptest"
	"path"
)

func getNewTestServerAndClient(t *testing.T, injectedConfig *config.Config) (*httptest.Server, apiv1connect.OliveTinApiServiceClient) {
	ex := executor.DefaultExecutor(injectedConfig)
	ex.RebuildActionMap()

	apiPath, apiHandler := GetNewHandler(ex)

	mux := http.NewServeMux()
	mux.Handle("/api/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Infof("HTTP Request: %s %s", r.Method, r.URL.Path)

		// Translate /api/<service>/<method> to <service>/<method>
		fn := path.Base(r.URL.Path)
		r.URL.Path = apiPath + fn

		apiHandler.ServeHTTP(w, r)
	}))

	log.Infof("API path is %s", apiPath)

	httpclient := &http.Client{}

	ts := httptest.NewServer(mux)

	client := apiv1connect.NewOliveTinApiServiceClient(httpclient, ts.URL+"/api")

	log.Infof("Test server URL is %s", ts.URL+"/api"+apiPath)

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

	respInit, errInit := client.Init(context.Background(), connect.NewRequest(&apiv1.InitRequest{}))
	respGetReady, errReady := client.GetReadyz(context.Background(), connect.NewRequest(&apiv1.GetReadyzRequest{}))

	if errInit != nil {
		t.Errorf("Init request failed: %v", errInit)
		return
	}

	if errReady != nil {
		t.Errorf("GetReadyz request failed: %v", errReady)
		return
	}

	log.Infof("GetReadyz response: %v", respGetReady.Msg)

	assert.Equal(t, true, true, "sayHello Failed")

	//	assert.Equal(t, 1, len(respGb.Msg.Actions), "Got 1 action button back")

	log.Printf("Response: %+v", respInit)

	respSa, err := client.StartAction(context.Background(), connect.NewRequest(&apiv1.StartActionRequest{
		//		ActionId: "blat"
	}))

	assert.NotNil(t, err, "Error 404 after start action")
	assert.Nil(t, respSa, "Nil response for non existing action")

	defer conn.Close()
}
