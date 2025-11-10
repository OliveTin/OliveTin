package config

import (
	"os"
	"testing"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
)

var stringEnvConfigYaml = `
PageTitle: ${{ INPUT }}
`

var stringEnvInterpolationConfigYaml = `
PageTitle: Olivetin - ${{ INPUT }}
`

var boolEnvConfigYaml = `
CheckForUpdates: ${{ INPUT }}
`

var numericEnvConfigYaml = `
LogHistoryPageSize: ${{ INPUT }}
`

var argsSyntaxConfigYaml = `
actions:
  - title: Ping host
    id: ping_host
    shell: ping {{ host }} -c ${{ INPUT }}
    icon: ping
    timeout: 100
    popupOnStart: execution-dialog-stdout-only
    arguments:
      - name: host
        title: Host
        type: ascii_identifier
        default: example.com
        description: The host that you want to ping
`

func pageTitleSelector(cfg *Config) any {
	return cfg.PageTitle
}

func checkForUpdatesSelector(cfg *Config) any {
	return cfg.CheckForUpdates
}

func logHistoryPageSizeSelector(cfg *Config) any {
	return cfg.LogHistoryPageSize
}

var envConfigTests = []struct {
	yaml     string
	input    string
	output   any
	selector func(*Config) any
}{
	// Test that it works for string type config fields, both standalone and as part of a larger string value.
	{stringEnvConfigYaml, "A Nice Title", "A Nice Title", pageTitleSelector},
	{stringEnvInterpolationConfigYaml, "A Nice Title", "Olivetin - A Nice Title", pageTitleSelector},
	// Test that unset variables turn into empty strings.
	{stringEnvConfigYaml, "", "", pageTitleSelector},
	// Test that it works for bool type config fields for intuitive bool->string conversions.
	{boolEnvConfigYaml, "FALSE", false, checkForUpdatesSelector},
	{boolEnvConfigYaml, "false", false, checkForUpdatesSelector},
	{boolEnvConfigYaml, "False", false, checkForUpdatesSelector},
	{boolEnvConfigYaml, "TRUE", true, checkForUpdatesSelector},
	{boolEnvConfigYaml, "true", true, checkForUpdatesSelector},
	{boolEnvConfigYaml, "True", true, checkForUpdatesSelector},
	{boolEnvConfigYaml, "0", false, checkForUpdatesSelector},
	{boolEnvConfigYaml, "1", true, checkForUpdatesSelector},
	// Test that unset variables turn into false bools.
	{boolEnvConfigYaml, "", false, checkForUpdatesSelector},
	// Test that it works for numeric type config fields.
	{numericEnvConfigYaml, "2048", int64(2048), logHistoryPageSizeSelector},
	// Test that unset variables turn into zero numbers.
	{numericEnvConfigYaml, "", int64(0), logHistoryPageSizeSelector},
	// Test that it doesn't interfere with similar arguments
	{argsSyntaxConfigYaml, "5", "ping {{ host }} -c 5", func(cfg *Config) any {
		if len(cfg.Actions) > 0 {
			return cfg.Actions[0].Shell
		}
		return ""
	}},
}

func TestEnvInConfig(t *testing.T) {
	for _, tt := range envConfigTests {
		cfg := DefaultConfig()
		setIfNotEmpty("INPUT", tt.input)
		processed := processYamlWithEnv(tt.yaml)
		k, err := loadKoanf(processed)
		if err != nil {
			t.Errorf("Error loading YAML: %v", err)
			continue
		}
		if err := k.Unmarshal(".", cfg); err != nil {
			t.Errorf("Error unmarshalling config: %v", err)
			continue
		}
		field := tt.selector(cfg)
		assert.Equal(t, tt.output, field, "Unmarshaled config field doesn't match expected value: env=\"%s\"", tt.input)
		os.Unsetenv("INPUT")
	}
}

func setIfNotEmpty(key, val string) {
	if val != "" {
		os.Setenv(key, val)
	}
}

func processYamlWithEnv(content string) string {
	return envRegex.ReplaceAllStringFunc(content, func(match string) string {
		submatches := envRegex.FindStringSubmatch(match)
		key := submatches[1]
		val, _ := os.LookupEnv(key)
		return val
	})
}

func loadKoanf(processed string) (*koanf.Koanf, error) {
	k := koanf.New(".")
	if err := k.Load(rawbytes.Provider([]byte(processed)), yaml.Parser()); err != nil {
		return nil, err
	}
	return k, nil
}
