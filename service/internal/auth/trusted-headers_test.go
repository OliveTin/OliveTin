package auth

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	types "github.com/OliveTin/OliveTin/internal/auth/authpublic"
	"github.com/OliveTin/OliveTin/internal/config"
)

func TestCheckUserFromHeadersIgnoresClientProviderHeader(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.AuthHttpHeaderUsername = "X-Username"

	header := http.Header{}
	header.Set("X-Username", "alice")
	header.Set("provider", "system")

	user := checkUserFromHeaders(&types.AuthCheckingContext{
		Request: &http.Request{Header: header},
		Config:  cfg,
	})

	assert.Equal(t, "alice", user.Username)
	assert.Equal(t, "trusted-header", user.Provider)
}
