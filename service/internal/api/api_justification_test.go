package api

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiv1 "github.com/OliveTin/OliveTin/gen/olivetin/api/v1"
	"github.com/OliveTin/OliveTin/internal/auth"
	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/executor"
)

func TestStartActionRequiresJustificationForGuest(t *testing.T) {
	cfg := config.DefaultConfig()
	action := &config.Action{
		Title:         "Send email",
		ID:            "send_email",
		Justification: " ",
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

	time.Sleep(200 * time.Millisecond)

	entry, ok := ex.GetLog(resp.Msg.ExecutionTrackingId)
	require.True(t, ok)
	assert.Equal(t, "New user registration foo@example.com", entry.Justification)
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
		Justification: " ",
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
	assert.Equal(t, " ", pb.Justification)
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

	time.Sleep(200 * time.Millisecond)

	entry, ok := ex.GetLog(resp.Msg.ExecutionTrackingId)
	require.True(t, ok)
	assert.Equal(t, "stuffbox", entry.Justification)
}

func TestValidateJustificationRequiredAllowsSystemUser(t *testing.T) {
	cfg := config.DefaultConfig()
	action := &config.Action{Title: "Cron job", Justification: " "}

	err := validateJustificationRequired(action, "", auth.UserFromSystem(cfg, "cron"))
	require.NoError(t, err)
}
