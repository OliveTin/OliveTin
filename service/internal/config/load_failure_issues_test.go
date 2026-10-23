package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/configissues"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func appendSourceFromFile(t *testing.T, basePath string) {
	t.Helper()
	k := koanf.New(".")
	require.NoError(t, k.Load(file.Provider(basePath), yaml.Parser()))
	config.AppendSource(config.DefaultConfig(), k, basePath)
}

func findStickyIssue(code string) *configissues.Issue {
	for _, issue := range configissues.CopySticky() {
		if issue.Code == code {
			return &issue
		}
	}
	return nil
}

func TestAppendSourceReportsUnparseableIncludeFile(t *testing.T) {
	dir := t.TempDir()
	includeDir := filepath.Join(dir, "config.d")
	require.NoError(t, os.Mkdir(includeDir, 0o755))

	basePath := filepath.Join(dir, "config.yaml")
	require.NoError(t, os.WriteFile(basePath, []byte("include: config.d\n"), 0o644))

	includePath := filepath.Join(includeDir, "10-broken.yaml")
	require.NoError(t, os.WriteFile(includePath, []byte("actions: [\n  - title: broken\n"), 0o644))

	appendSourceFromFile(t, basePath)

	issue := findStickyIssue(configissues.CodeIncludeParse)
	require.NotNil(t, issue)
	assert.Equal(t, configissues.SeverityError, issue.Severity)
	assert.Equal(t, includePath, issue.ConfigFile)
}

func TestAppendSourceReportsDecodeFailure(t *testing.T) {
	basePath := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(basePath, []byte("actions: not-a-list-of-actions\n"), 0o644))

	appendSourceFromFile(t, basePath)

	issue := findStickyIssue(configissues.CodeConfigDecode)
	require.NotNil(t, issue)
	assert.Equal(t, configissues.SeverityError, issue.Severity)
}

func TestAppendSourceClearsLoadFailuresOnSuccessfulLoad(t *testing.T) {
	basePath := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(basePath, []byte("actions: not-a-list-of-actions\n"), 0o644))
	appendSourceFromFile(t, basePath)
	require.NotNil(t, findStickyIssue(configissues.CodeConfigDecode))

	require.NoError(t, os.WriteFile(basePath, []byte("pageTitle: Fixed\n"), 0o644))
	appendSourceFromFile(t, basePath)

	assert.Nil(t, findStickyIssue(configissues.CodeConfigDecode))
}
