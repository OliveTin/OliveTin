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
		Justification: true,
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

func TestBuildActionExposesJustificationFlag(t *testing.T) {
	cfg := config.DefaultConfig()
	action := &config.Action{
		Title:         "Audited action",
		ID:            "audited",
		Justification: true,
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
	assert.True(t, pb.Justification)
}

func TestValidateJustificationRequiredAllowsSystemUser(t *testing.T) {
	cfg := config.DefaultConfig()
	action := &config.Action{Title: "Cron job", Justification: true}

	err := validateJustificationRequired(action, "", auth.UserFromSystem(cfg, "cron"))
	require.NoError(t, err)
}
