package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/OliveTin/OliveTin/internal/config"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppendSourceStampsActionSourceFiles(t *testing.T) {
	dir := t.TempDir()
	includeDir := filepath.Join(dir, "config.d")
	require.NoError(t, os.Mkdir(includeDir, 0o755))

	basePath := filepath.Join(dir, "config.yaml")
	require.NoError(t, os.WriteFile(basePath, []byte(`
include: config.d
actions:
  - title: From base
    shell: echo base
`), 0o644))

	includePath := filepath.Join(includeDir, "10-extra.yaml")
	require.NoError(t, os.WriteFile(includePath, []byte(`
actions:
  - title: From include
    shell: echo include
entities:
  - name: host
    file: hosts.yaml
`), 0o644))

	k := koanf.New(".")
	require.NoError(t, k.Load(file.Provider(basePath), yaml.Parser()))

	cfg := config.DefaultConfig()
	config.AppendSource(cfg, k, basePath)

	require.Len(t, cfg.Actions, 2)
	assert.Equal(t, basePath, cfg.Actions[0].SourceFile)
	assert.Equal(t, includePath, cfg.Actions[1].SourceFile)

	require.Len(t, cfg.Entities, 1)
	assert.Equal(t, includePath, cfg.Entities[0].SourceFile)
}

func TestAppendSourceIgnoresUserProvidedSourceFile(t *testing.T) {
	dir := t.TempDir()
	basePath := filepath.Join(dir, "config.yaml")
	require.NoError(t, os.WriteFile(basePath, []byte(`
actions:
  - title: Spoofed
    shell: echo hi
    x-olivetin-source-file: /tmp/fake-user-path.yaml
entities:
  - name: host
    file: hosts.yaml
    x-olivetin-source-file: /tmp/fake-entity-path.yaml
`), 0o644))

	k := koanf.New(".")
	require.NoError(t, k.Load(file.Provider(basePath), yaml.Parser()))

	cfg := config.DefaultConfig()
	config.AppendSource(cfg, k, basePath)

	require.Len(t, cfg.Actions, 1)
	assert.Equal(t, basePath, cfg.Actions[0].SourceFile)
	assert.NotEqual(t, "/tmp/fake-user-path.yaml", cfg.Actions[0].SourceFile)

	require.Len(t, cfg.Entities, 1)
	assert.Equal(t, basePath, cfg.Entities[0].SourceFile)
	assert.NotEqual(t, "/tmp/fake-entity-path.yaml", cfg.Entities[0].SourceFile)
}
