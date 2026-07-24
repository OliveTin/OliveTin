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
)

func TestGetDiagnosticsReturnsConfigIssues(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Actions = []*config.Action{
		{
			Title:  "Deploy",
			ID:     "deploy",
			Entity: "host",
			Arguments: []config.ActionArgument{
				{
					Name:    "ansible_host",
					Type:    "ascii_identifier",
					Default: `{{ host.name | default "All" }}`,
				},
			},
			ExecOnCron: []string{"0 1 * * *"},
		},
	}

	testServer, client := getNewTestServerAndClient(cfg)
	defer testServer.Close()

	res, err := client.GetDiagnostics(context.Background(), connect.NewRequest(&apiv1.GetDiagnosticsRequest{}))
	require.NoError(t, err)
	require.NotNil(t, res.Msg)
	require.NotEmpty(t, res.Msg.ConfigIssues)

	codes := map[string]bool{}
	for _, issue := range res.Msg.ConfigIssues {
		codes[issue.Code] = true
	}
	assert.True(t, codes[configissues.CodeTemplateParse])
	assert.True(t, codes[configissues.CodeCronEntityBinding])

	initRes, err := client.Init(context.Background(), connect.NewRequest(&apiv1.InitRequest{}))
	require.NoError(t, err)
	assert.Equal(t, int32(len(res.Msg.ConfigIssues)), initRes.Msg.ConfigIssueCount)
}

func TestInitHidesConfigIssueCountWithoutDiagnostics(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.DefaultPolicy.ShowDiagnostics = false
	cfg.Actions = []*config.Action{
		{
			Title:  "Deploy",
			ID:     "deploy",
			Entity: "host",
			Arguments: []config.ActionArgument{
				{
					Name:    "ansible_host",
					Type:    "ascii_identifier",
					Default: `{{ host.name | default "All" }}`,
				},
			},
		},
	}

	testServer, client := getNewTestServerAndClient(cfg)
	defer testServer.Close()

	initRes, err := client.Init(context.Background(), connect.NewRequest(&apiv1.InitRequest{}))
	require.NoError(t, err)
	assert.Equal(t, int32(0), initRes.Msg.ConfigIssueCount)

	_, err = client.GetDiagnostics(context.Background(), connect.NewRequest(&apiv1.GetDiagnosticsRequest{}))
	require.Error(t, err)
	assert.Equal(t, connect.CodePermissionDenied, connect.CodeOf(err))
}

func attachBrokenTemplateArgToSecretAction(cfg *config.Config) {
	for i := range cfg.Actions {
		if cfg.Actions[i].ID != "secret_action" {
			continue
		}
		cfg.Actions[i].Arguments = []config.ActionArgument{
			{
				Name:    "target",
				Type:    "ascii_identifier",
				Default: `{{ host.name | default "All" }}`,
			},
		}
		return
	}
}

func assertNoActionScopedConfigIssues(t *testing.T, issues []*apiv1.ConfigIssue) {
	t.Helper()
	for _, issue := range issues {
		assert.Empty(t, issue.ActionId,
			"user with view:false must not see action-scoped config issues")
	}
}

func configIssuesContainActionID(issues []*apiv1.ConfigIssue, actionID string) bool {
	for _, issue := range issues {
		if issue.ActionId == actionID {
			return true
		}
	}
	return false
}

func TestConfigIssuesHideActionsWithoutViewPermission(t *testing.T) {
	cfg, _, _ := buildViewPermissionTestConfig(t)
	cfg.AuthHttpHeaderUsername = "X-Ot-User"
	attachBrokenTemplateArgToSecretAction(cfg)

	testServer, client := getNewTestServerAndClient(cfg)
	defer testServer.Close()

	configissues.Report(configissues.Issue{
		Severity: configissues.SeverityWarning,
		Code:     configissues.CodeEnvUnset,
		Message:  "TEST_GLOBAL_ENV is not set",
		Source:   "TEST_GLOBAL_ENV",
	})

	lowReq := connect.NewRequest(&apiv1.GetDiagnosticsRequest{})
	lowReq.Header().Set("X-Ot-User", "low")
	lowRes, err := client.GetDiagnostics(context.Background(), lowReq)
	require.NoError(t, err)

	assertNoActionScopedConfigIssues(t, lowRes.Msg.ConfigIssues)
	require.NotEmpty(t, lowRes.Msg.ConfigIssues, "low user should still see global config issues")

	lowInit := connect.NewRequest(&apiv1.InitRequest{})
	lowInit.Header().Set("X-Ot-User", "low")
	lowInitRes, err := client.Init(context.Background(), lowInit)
	require.NoError(t, err)
	assert.Equal(t, int32(len(lowRes.Msg.ConfigIssues)), lowInitRes.Msg.ConfigIssueCount)

	adminReq := connect.NewRequest(&apiv1.GetDiagnosticsRequest{})
	adminReq.Header().Set("X-Ot-User", "admin")
	adminRes, err := client.GetDiagnostics(context.Background(), adminReq)
	require.NoError(t, err)

	assert.True(t, configIssuesContainActionID(adminRes.Msg.ConfigIssues, "secret_action"),
		"admin must still see config issues for secret_action")
	assert.Greater(t, len(adminRes.Msg.ConfigIssues), len(lowRes.Msg.ConfigIssues))
}

func TestConfigIssuesHideRuntimeWatcherFailuresWithoutViewPermission(t *testing.T) {
	cfg, _, _ := buildViewPermissionTestConfig(t)
	cfg.AuthHttpHeaderUsername = "X-Ot-User"

	testServer, client := getNewTestServerAndClient(cfg)
	defer testServer.Close()

	configissues.Report(configissues.Issue{
		Severity:    configissues.SeverityError,
		Code:        configissues.CodeWatcherPath,
		Message:     `Could not create watcher for "/secret/path": permission denied`,
		ActionID:    "secret_action",
		ActionTitle: "Secret Action",
		Source:      "/secret/path",
		ConfigFile:  "/etc/OliveTin/secret.yaml",
	})

	lowReq := connect.NewRequest(&apiv1.GetDiagnosticsRequest{})
	lowReq.Header().Set("X-Ot-User", "low")
	lowRes, err := client.GetDiagnostics(context.Background(), lowReq)
	require.NoError(t, err)

	for _, issue := range lowRes.Msg.ConfigIssues {
		assert.NotEqual(t, configissues.CodeWatcherPath, issue.Code,
			"user with view:false must not see action-scoped watcher_path issues")
		assert.NotContains(t, issue.Source, "/secret/path")
	}

	adminReq := connect.NewRequest(&apiv1.GetDiagnosticsRequest{})
	adminReq.Header().Set("X-Ot-User", "admin")
	adminRes, err := client.GetDiagnostics(context.Background(), adminReq)
	require.NoError(t, err)

	foundWatcher := false
	for _, issue := range adminRes.Msg.ConfigIssues {
		if issue.Code == configissues.CodeWatcherPath && issue.ActionId == "secret_action" {
			foundWatcher = true
			break
		}
	}
	assert.True(t, foundWatcher, "admin must see action-scoped watcher_path issues")
}
