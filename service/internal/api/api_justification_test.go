package api

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiv1 "github.com/OliveTin/OliveTin/gen/olivetin/api/v1"
	apiv1connect "github.com/OliveTin/OliveTin/gen/olivetin/api/v1/apiv1connect"
	"github.com/OliveTin/OliveTin/internal/auth"
	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/executor"
)

func TestStartActionRequiresJustificationForGuest(t *testing.T) {
	cfg := config.DefaultConfig()
	action := &config.Action{
		Title:         "Send email",
		ID:            "send_email",
		Justification: config.JustificationRequiredNoTemplate,
		Shell:         "echo done",
	}
	cfg.Actions = append(cfg.Actions, action)

	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()
	binding := ex.FindBindingWithNoEntity(action)
	require.NotNil(t, binding)

	ts, client := getNewTestServerAndClientWithExecutor(cfg, ex)
	defer ts.Close()

	_, err := client.StartAction(context.Background(), connect.NewRequest(&apiv1.StartActionRequest{
		BindingId:        binding.ID,
		UniqueTrackingId: uuid.NewString(),
	}))
	require.Error(t, err)
	assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))

	resp, err := client.StartAction(context.Background(), connect.NewRequest(&apiv1.StartActionRequest{
		BindingId:        binding.ID,
		UniqueTrackingId: uuid.NewString(),
		Justification:    "New user registration foo@example.com",
	}))
	require.NoError(t, err)
	require.NotEmpty(t, resp.Msg.ExecutionTrackingId)

	waitForLogJustification(t, ex, resp.Msg.ExecutionTrackingId, "New user registration foo@example.com")
}

func TestBuildActionExposesJustificationTemplate(t *testing.T) {
	cfg := config.DefaultConfig()
	action := &config.Action{
		Title:         "Audited action",
		ID:            "audited",
		Justification: "{{ target }}",
		Shell:         "echo hi",
	}
	cfg.Actions = append(cfg.Actions, action)

	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()
	binding := ex.FindBindingWithNoEntity(action)
	require.NotNil(t, binding)

	pb := buildAction(binding, &DashboardRenderRequest{
		cfg: cfg,
		ex:  ex,
	})

	require.NotNil(t, pb)
	assert.Equal(t, "{{ target }}", pb.Justification)
}

func TestBuildActionExposesBlankRequiredJustification(t *testing.T) {
	cfg := config.DefaultConfig()
	action := &config.Action{
		Title:         "Audited action",
		ID:            "audited",
		Justification: config.JustificationRequiredNoTemplate,
		Shell:         "echo hi",
	}
	cfg.Actions = append(cfg.Actions, action)

	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()
	binding := ex.FindBindingWithNoEntity(action)
	require.NotNil(t, binding)

	pb := buildAction(binding, &DashboardRenderRequest{
		cfg: cfg,
		ex:  ex,
	})

	require.NotNil(t, pb)
	assert.Equal(t, config.JustificationRequiredNoTemplate, pb.Justification)
}

func TestResolveStartJustificationUsesTemplateWhenClientValueEmpty(t *testing.T) {
	action := &config.Action{
		Justification: "{{ ansible_host }}",
	}
	binding := &executor.ActionBinding{}

	got := resolveStartJustification(action, binding, "", map[string]string{
		"ansible_host": "192.168.66.8",
	})
	assert.Equal(t, "192.168.66.8", got)
}

func TestResolveStartJustificationPrefersClientValue(t *testing.T) {
	action := &config.Action{
		Justification: "{{ ansible_host }}",
	}
	binding := &executor.ActionBinding{}

	got := resolveStartJustification(action, binding, "manual reason", map[string]string{
		"ansible_host": "192.168.66.8",
	})
	assert.Equal(t, "manual reason", got)
}

func TestStartActionResolvesJustificationTemplateForGuest(t *testing.T) {
	cfg := config.DefaultConfig()
	action := &config.Action{
		Title:         "Run playbook",
		ID:            "run_playbook",
		Justification: "{{ ansible_host }}",
		Shell:         "echo done",
		Arguments: []config.ActionArgument{
			{Name: "ansible_host", Title: "Host"},
		},
	}
	cfg.Actions = append(cfg.Actions, action)

	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()
	binding := ex.FindBindingWithNoEntity(action)
	require.NotNil(t, binding)

	ts, client := getNewTestServerAndClientWithExecutor(cfg, ex)
	defer ts.Close()

	resp, err := client.StartAction(context.Background(), connect.NewRequest(&apiv1.StartActionRequest{
		BindingId:        binding.ID,
		UniqueTrackingId: uuid.NewString(),
		Arguments: []*apiv1.StartActionArgument{
			{Name: "ansible_host", Value: "stuffbox"},
		},
	}))
	require.NoError(t, err)
	require.NotEmpty(t, resp.Msg.ExecutionTrackingId)

	waitForLogJustification(t, ex, resp.Msg.ExecutionTrackingId, "stuffbox")
}

func TestValidateJustificationRequiredAllowsSystemUser(t *testing.T) {
	cfg := config.DefaultConfig()
	action := &config.Action{Title: "Cron job", Justification: config.JustificationRequiredNoTemplate}

	err := validateJustificationRequired(action, "", auth.UserFromSystem(cfg, "cron"))
	require.NoError(t, err)
}

func newJustificationRequiredTestServer(t *testing.T, cfg *config.Config) (apiv1connect.OliveTinApiServiceClient, func()) {
	t.Helper()

	cfg.Actions = append(cfg.Actions, &config.Action{
		Title:         "Approve change",
		ID:            "approve",
		Justification: config.JustificationRequiredNoTemplate,
		Shell:         "echo approved",
	})

	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()

	ts, client := getNewTestServerAndClientWithExecutor(cfg, ex)
	return client, ts.Close
}

func TestStartActionByGetRequiresJustification(t *testing.T) {
	client, closeServer := newJustificationRequiredTestServer(t, config.DefaultConfig())
	defer closeServer()

	//nolint:staticcheck // Deliberately calling the deprecated RPC.
	_, err := client.StartActionByGet(context.Background(), connect.NewRequest(&apiv1.StartActionByGetRequest{ActionId: "approve"}))
	require.Error(t, err)
	assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
}

func TestStartActionByGetAndWaitRequiresJustification(t *testing.T) {
	client, closeServer := newJustificationRequiredTestServer(t, config.DefaultConfig())
	defer closeServer()

	//nolint:staticcheck // Deliberately calling the deprecated RPC.
	_, err := client.StartActionByGetAndWait(context.Background(), connect.NewRequest(&apiv1.StartActionByGetAndWaitRequest{ActionId: "approve"}))
	require.Error(t, err)
	assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
}

func TestStartActionByGetUsesJustificationTemplate(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Actions = append(cfg.Actions, &config.Action{
		Title:         "Templated",
		ID:            "templated",
		Justification: "Scheduled maintenance",
		Shell:         "echo hi",
	})

	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()
	ts, client := getNewTestServerAndClientWithExecutor(cfg, ex)
	defer ts.Close()

	//nolint:staticcheck // Deliberately calling the deprecated RPC.
	resp, err := client.StartActionByGet(context.Background(), connect.NewRequest(&apiv1.StartActionByGetRequest{ActionId: "templated"}))
	require.NoError(t, err)

	waitForLogJustification(t, ex, resp.Msg.ExecutionTrackingId, "Scheduled maintenance")
}

func TestTrustedHeaderProviderSystemDoesNotSkipJustification(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.AuthHttpHeaderUsername = "X-Username"
	cfg.AuthHttpHeaderUserGroup = "X-Usergroup"

	client, closeServer := newJustificationRequiredTestServer(t, cfg)
	defer closeServer()

	req := connect.NewRequest(&apiv1.StartActionAndWaitRequest{ActionId: "approve"})
	req.Header().Set("X-Username", "alice")
	req.Header().Set("X-Usergroup", "operators")
	req.Header().Set("provider", "system")

	_, err := client.StartActionAndWait(context.Background(), req)
	require.Error(t, err)
	assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
}
