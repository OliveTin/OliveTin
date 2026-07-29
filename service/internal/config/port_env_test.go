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
	// parseEnvPort accepts only 1..65535; port 0 and out-of-range values are ignored.
	cases := []struct {
		name string
		port string
	}{
		{name: "non-numeric", port: "not-a-port"},
		{name: "above max", port: "65536"},
		{name: "negative", port: "-1"},
		{name: "zero ignored", port: "0"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("PORT", tc.port)

			cfg := DefaultConfig()
			cfg.ListenAddressSingleHTTPFrontend = "0.0.0.0:1337"
			applyPortEnvironmentOverride(cfg)

			assert.Equal(t, "0.0.0.0:1337", cfg.ListenAddressSingleHTTPFrontend)
		})
	}
}

func TestApplyPortEnvironmentOverrideEmptyListenAddressDefaultsHost(t *testing.T) {
	t.Setenv("PORT", "8080")

	cfg := DefaultConfig()
	cfg.ListenAddressSingleHTTPFrontend = ""
	applyPortEnvironmentOverride(cfg)

	assert.Equal(t, "0.0.0.0:8080", cfg.ListenAddressSingleHTTPFrontend)
}

func TestApplyPortEnvironmentOverrideRejectsMalformedListenAddress(t *testing.T) {
	t.Setenv("PORT", "8080")

	cfg := DefaultConfig()
	cfg.ListenAddressSingleHTTPFrontend = "not-a-valid-address"
	applyPortEnvironmentOverride(cfg)

	assert.Equal(t, "not-a-valid-address", cfg.ListenAddressSingleHTTPFrontend)
}
