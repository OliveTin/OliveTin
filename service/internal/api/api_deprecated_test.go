package api

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiv1 "github.com/OliveTin/OliveTin/gen/olivetin/api/v1"
	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/configissues"
	"github.com/OliveTin/OliveTin/internal/executor"
)

func deprecatedAPICallSources(t *testing.T, client interface {
	GetDiagnostics(context.Context, *connect.Request[apiv1.GetDiagnosticsRequest]) (*connect.Response[apiv1.GetDiagnosticsResponse], error)
}) map[string]bool {
	t.Helper()

	res, err := client.GetDiagnostics(context.Background(), connect.NewRequest(&apiv1.GetDiagnosticsRequest{}))
	require.NoError(t, err)

	sources := map[string]bool{}
	for _, issue := range res.Msg.ConfigIssues {
		if issue.Code == configissues.CodeDeprecatedAPICall {
			assert.Equal(t, configissues.SeverityWarning, issue.Severity)
			sources[issue.Source] = true
		}
	}
	return sources
}

func TestDeprecatedStartActionByGetRaisesDiagnosticsWarning(t *testing.T) {
	configissues.ClearRuntime()
	t.Cleanup(configissues.ClearRuntime)

	cfg := config.DefaultConfig()
	cfg.Actions = []*config.Action{{Title: "Ping", ID: "ping", Shell: "echo ping"}}

	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()
	testServer, client := getNewTestServerAndClientWithExecutor(cfg, ex)
	defer testServer.Close()

	assert.Empty(t, deprecatedAPICallSources(t, client))

	//nolint:staticcheck // Deliberately calling the deprecated RPC.
	_, err := client.StartActionByGet(context.Background(), connect.NewRequest(&apiv1.StartActionByGetRequest{ActionId: "ping"}))
	require.NoError(t, err)

	//nolint:staticcheck // Deliberately calling the deprecated RPC.
	_, err = client.StartActionByGetAndWait(context.Background(), connect.NewRequest(&apiv1.StartActionByGetAndWaitRequest{ActionId: "ping"}))
	require.NoError(t, err)

	sources := deprecatedAPICallSources(t, client)
	assert.True(t, sources["StartActionByGet"])
	assert.True(t, sources["StartActionByGetAndWait"])

	// Config reloads and entity changes rebuild the issue list; the warnings must survive.
	ex.RebuildActionMap()
	sources = deprecatedAPICallSources(t, client)
	assert.True(t, sources["StartActionByGet"])
	assert.True(t, sources["StartActionByGetAndWait"])
}

func TestStartActionDoesNotRaiseDeprecatedWarning(t *testing.T) {
	configissues.ClearRuntime()
	t.Cleanup(configissues.ClearRuntime)

	cfg := config.DefaultConfig()
	cfg.Actions = []*config.Action{{Title: "Ping", ID: "ping", Shell: "echo ping"}}

	testServer, client := getNewTestServerAndClient(cfg)
	defer testServer.Close()

	_, err := client.StartAction(context.Background(), connect.NewRequest(&apiv1.StartActionRequest{BindingId: "ping"}))
	require.NoError(t, err)

	assert.Empty(t, deprecatedAPICallSources(t, client))
}
