package config

import (
	"testing"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJustificationDecodeHookMigratesLegacyBooleanFalse(t *testing.T) {
	cfg := loadJustificationCompatConfig(t, `
actions:
  - title: Legacy disabled
    shell: echo hi
    justification: false
`)

	require.Len(t, cfg.Actions, 1)
	assert.Empty(t, cfg.Actions[0].Justification)
}

func TestJustificationDecodeHookMigratesLegacyBooleanTrue(t *testing.T) {
	cfg := loadJustificationCompatConfig(t, `
actions:
  - title: Legacy required
    shell: echo hi
    justification: true
`)

	require.Len(t, cfg.Actions, 1)
	assert.Equal(t, JustificationRequiredNoTemplate, cfg.Actions[0].Justification)
}

func TestSanitizeJustificationMigratesWeaklyTypedLegacyStrings(t *testing.T) {
	action := &Action{Justification: "false"}
	action.sanitizeJustification()
	assert.Empty(t, action.Justification)

	action.Justification = "true"
	action.sanitizeJustification()
	assert.Equal(t, JustificationRequiredNoTemplate, action.Justification)
}

func loadJustificationCompatConfig(t *testing.T, yamlBody string) *Config {
	t.Helper()

	k := koanf.New(".")
	require.NoError(t, k.Load(rawbytes.Provider([]byte(yamlBody)), yaml.Parser()))

	cfg := DefaultConfig()
	require.True(t, unmarshalRoot(k, cfg))
	cfg.Sanitize()

	return cfg
}
