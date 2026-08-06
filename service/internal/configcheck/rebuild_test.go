package configcheck_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/configcheck"
	"github.com/OliveTin/OliveTin/internal/configissues"
	"github.com/OliveTin/OliveTin/internal/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRebuildTemplateParseDefault(t *testing.T) {
	configissues.BeginConfigLoad()
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
		},
	}

	configcheck.Rebuild(cfg)

	issues := configissues.List()
	require.NotEmpty(t, issues)
	assert.True(t, hasCode(issues, configissues.CodeTemplateParse))
	assert.True(t, hasCode(issues, configissues.CodeEntityEmpty))
	assert.False(t, hasCode(issues, configissues.CodeCronEntityBinding))
}

func TestRebuildValidTemplateDefault(t *testing.T) {
	configissues.BeginConfigLoad()
	cfg := config.DefaultConfig()
	cfg.Actions = []*config.Action{
		{
			Title: "Sleep",
			ID:    "sleep",
			Arguments: []config.ActionArgument{
				{
					Name:    "ssh_host",
					Type:    "ascii_identifier",
					Default: `{{ host.hostname }}`,
				},
			},
		},
	}

	configcheck.Rebuild(cfg)

	assert.False(t, hasCode(configissues.List(), configissues.CodeTemplateParse))
}

func TestRebuildCronEntityBinding(t *testing.T) {
	configissues.BeginConfigLoad()
	cfg := config.DefaultConfig()
	cfg.Actions = []*config.Action{
		{
			Title:      "Nightly",
			ID:         "nightly",
			Entity:     "host",
			ExecOnCron: []string{"0 1 * * *"},
		},
	}

	configcheck.Rebuild(cfg)

	assert.True(t, hasCode(configissues.List(), configissues.CodeCronEntityBinding))
}

func TestRebuildInvalidCron(t *testing.T) {
	configissues.BeginConfigLoad()
	cfg := config.DefaultConfig()
	cfg.Actions = []*config.Action{
		{
			Title:      "Broken cron",
			ID:         "broken",
			ExecOnCron: []string{"not a cron line"},
		},
	}

	configcheck.Rebuild(cfg)

	assert.True(t, hasCode(configissues.List(), configissues.CodeCronInvalid))
}

func TestRebuildMissingEntityFile(t *testing.T) {
	configissues.BeginConfigLoad()
	cfg := config.DefaultConfig()
	dir := t.TempDir()
	cfg.SetDir(dir)
	cfg.Entities = []*config.EntityFile{
		{Name: "host", File: "missing-hosts.yaml"},
	}

	configcheck.Rebuild(cfg)

	assert.True(t, hasCode(configissues.List(), configissues.CodeEntityFile))
}

func TestRebuildEntityFileLoadsAndClearsEmpty(t *testing.T) {
	configissues.BeginConfigLoad()
	cfg := config.DefaultConfig()
	dir := t.TempDir()
	cfg.SetDir(dir)

	entityPath := filepath.Join(dir, "hosts.yaml")
	require.NoError(t, os.WriteFile(entityPath, []byte("- name: cradle\n"), 0o644))

	cfg.Entities = []*config.EntityFile{
		{Name: "host", File: "hosts.yaml"},
	}
	cfg.Actions = []*config.Action{
		{Title: "Wake", ID: "wake", Entity: "host"},
	}

	entities.AddEntity("host", "0", map[string]any{"name": "cradle"})
	t.Cleanup(func() { entities.ClearEntitiesOfType("host") })

	configcheck.Rebuild(cfg)

	issues := configissues.List()
	assert.False(t, hasCode(issues, configissues.CodeEntityFile))
	assert.False(t, hasCode(issues, configissues.CodeEntityEmpty))
}

func TestRebuildIncludesActionConfigFile(t *testing.T) {
	configissues.BeginConfigLoad()
	cfg := config.DefaultConfig()
	cfg.Actions = []*config.Action{
		{
			Title:      "Nightly",
			ID:         "nightly",
			Entity:     "host",
			SourceFile: "/etc/OliveTin/config.d/16-gitops-sr-soe.yaml",
			ExecOnCron: []string{"0 1 * * *"},
		},
	}

	configcheck.Rebuild(cfg)

	var found bool
	for _, issue := range configissues.List() {
		if issue.Code == configissues.CodeCronEntityBinding {
			assert.Equal(t, "/etc/OliveTin/config.d/16-gitops-sr-soe.yaml", issue.ConfigFile)
			found = true
		}
	}
	assert.True(t, found)
}

func TestRebuildSkipsEntityEmptyBeforeLoadAttempt(t *testing.T) {
	entities.ResetEntityLoadAttempts()
	configissues.BeginConfigLoad()
	cfg := config.DefaultConfig()
	cfg.Entities = []*config.EntityFile{{Name: "container", File: "containers.json"}}
	cfg.Actions = []*config.Action{
		{Title: "Start container", ID: "start", Entity: "container"},
	}

	configcheck.Rebuild(cfg)
	assert.False(t, hasCode(configissues.List(), configissues.CodeEntityEmpty))

	entities.MarkEntityLoadAttempted("container")
	configcheck.Rebuild(cfg)
	assert.True(t, hasCode(configissues.List(), configissues.CodeEntityEmpty))
}

func TestStickyEnvUnsetSurvivesRebuild(t *testing.T) {
	configissues.BeginConfigLoad()
	configissues.ReportSticky(configissues.Issue{
		Severity: configissues.SeverityWarning,
		Code:     configissues.CodeEnvUnset,
		Message:  "unset",
		Source:   "MISSING_VAR",
	})

	configcheck.Rebuild(config.DefaultConfig())

	assert.True(t, hasCode(configissues.List(), configissues.CodeEnvUnset))
}

func TestRebuildWarnsForEntityTypeWithoutConfigEntry(t *testing.T) {
	entities.ClearEntitiesOfType("orphan_type")
	t.Cleanup(func() {
		entities.ClearEntitiesOfType("orphan_type")
	})

	entities.AddEntity("orphan_type", "0", map[string]any{"name": "lonely"})

	configissues.BeginConfigLoad()
	cfg := config.DefaultConfig()
	cfg.Entities = nil
	configcheck.Rebuild(cfg)

	issues := configissues.List()
	require.True(t, hasCode(issues, configissues.CodeEntityTypeUnconfigured))
	found := false
	for _, issue := range issues {
		if issue.Code == configissues.CodeEntityTypeUnconfigured {
			assert.Equal(t, configissues.SeverityWarning, issue.Severity)
			assert.Equal(t, "orphan_type", issue.Source)
			found = true
		}
	}
	assert.True(t, found)
}

func TestRebuildEntityArgumentChoices(t *testing.T) {
	configissues.BeginConfigLoad()
	cfg := config.DefaultConfig()
	cfg.Actions = []*config.Action{
		{
			Title: "Reboot",
			ID:    "reboot",
			Arguments: []config.ActionArgument{
				{
					Name:   "target",
					Entity: "servers",
					Choices: []config.ActionArgumentChoice{
						{Value: "{{ servers.name }}"},
						{Value: "web01"},
					},
				},
			},
		},
	}

	configcheck.Rebuild(cfg)

	issues := configissues.List()
	require.True(t, hasCode(issues, configissues.CodeEntityArgumentChoices))
	for _, issue := range issues {
		if issue.Code == configissues.CodeEntityArgumentChoices {
			assert.Equal(t, configissues.SeverityError, issue.Severity)
			assert.Equal(t, "target", issue.ArgumentName)
		}
	}
}

func hasCode(issues []configissues.Issue, code string) bool {
	for _, issue := range issues {
		if issue.Code == code {
			return true
		}
	}
	return false
}
