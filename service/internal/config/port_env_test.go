package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApplyPortEnvironmentOverride(t *testing.T) {
	t.Setenv("PORT", "8080")

	cfg := DefaultConfig()
	cfg.ListenAddressSingleHTTPFrontend = "0.0.0.0:1337"
	applyPortEnvironmentOverride(cfg)

	assert.Equal(t, "0.0.0.0:8080", cfg.ListenAddressSingleHTTPFrontend)
}

func TestApplyPortEnvironmentOverridePreservesHost(t *testing.T) {
	t.Setenv("PORT", "9000")

	cfg := DefaultConfig()
	cfg.ListenAddressSingleHTTPFrontend = "127.0.0.1:1337"
	applyPortEnvironmentOverride(cfg)

	assert.Equal(t, "127.0.0.1:9000", cfg.ListenAddressSingleHTTPFrontend)
}

func TestApplyPortEnvironmentOverrideUnsetLeavesConfig(t *testing.T) {
	t.Setenv("PORT", "")

	cfg := DefaultConfig()
	cfg.ListenAddressSingleHTTPFrontend = "0.0.0.0:2337"
	applyPortEnvironmentOverride(cfg)

	assert.Equal(t, "0.0.0.0:2337", cfg.ListenAddressSingleHTTPFrontend)
}

func TestApplyPortEnvironmentOverrideIgnoresInvalid(t *testing.T) {
	t.Setenv("PORT", "not-a-port")

	cfg := DefaultConfig()
	cfg.ListenAddressSingleHTTPFrontend = "0.0.0.0:1337"
	applyPortEnvironmentOverride(cfg)

	assert.Equal(t, "0.0.0.0:1337", cfg.ListenAddressSingleHTTPFrontend)
}
