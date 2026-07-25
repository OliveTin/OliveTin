package configcheck_test

import (
	"path/filepath"
	"testing"

	"github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/configcheck"
	"github.com/OliveTin/OliveTin/internal/configissues"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRebuildWatcherPathMissingDir(t *testing.T) {
	configissues.BeginConfigLoad()
	configissues.Replace(nil)

	missing := filepath.Join(t.TempDir(), "does-not-exist")
	cfg := config.DefaultConfig()
	cfg.Actions = []*config.Action{
		{
			Title:                  "On create",
			ID:                     "on_create",
			SourceFile:             "/etc/OliveTin/config.yaml",
			ExecOnFileCreatedInDir: []string{missing},
		},
	}

	configcheck.Rebuild(cfg)

	issues := configissues.List()
	require.True(t, hasCode(issues, configissues.CodeWatcherPath))

	var found *configissues.Issue
	for i := range issues {
		if issues[i].Code == configissues.CodeWatcherPath {
			found = &issues[i]
			break
		}
	}
	require.NotNil(t, found)
	assert.Equal(t, missing, found.Source)
	assert.Equal(t, "/etc/OliveTin/config.yaml", found.ConfigFile)
	assert.Contains(t, found.Message, missing)
	assert.Contains(t, found.Message, "Could not create watcher")
}
