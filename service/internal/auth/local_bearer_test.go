package auth

import (
	"net/http/httptest"
	"testing"

	authpublic "github.com/OliveTin/OliveTin/internal/auth/authpublic"
	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckUserFromLocalBearerApiKey_Match_LowercaseBearerScheme(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.AuthLocalUsers.Enabled = true
	cfg.AuthLocalUsers.Users = []*config.LocalUser{{
		Username:  "bot",
		Usergroup: "bots",
		ApiKey:    "secret-api-key",
	}}

	req := httptest.NewRequest("POST", "/", nil)
	req.Header.Set("Authorization", "bearer secret-api-key")

	ctx := &authpublic.AuthCheckingContext{Request: req, Config: cfg}
	user := checkUserFromLocalBearerApiKey(ctx)
	require.NotNil(t, user)
	assert.Equal(t, "bot", user.Username)
	assert.Equal(t, "bots", user.UsergroupLine)
	assert.Equal(t, "local", user.Provider)
}

func TestCheckUserFromLocalBearerApiKey_Match(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.AuthLocalUsers.Enabled = true
	cfg.AuthLocalUsers.Users = []*config.LocalUser{{
		Username:  "bot",
		Usergroup: "bots",
		ApiKey:    "secret-api-key",
	}}

	req := httptest.NewRequest("POST", "/", nil)
	req.Header.Set("Authorization", "Bearer secret-api-key")

	ctx := &authpublic.AuthCheckingContext{Request: req, Config: cfg}
	user := checkUserFromLocalBearerApiKey(ctx)
	require.NotNil(t, user)
	assert.Equal(t, "bot", user.Username)
	assert.Equal(t, "bots", user.UsergroupLine)
	assert.Equal(t, "local", user.Provider)
}

func TestCheckUserFromLocalBearerApiKey_WrongKey(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.AuthLocalUsers.Enabled = true
	cfg.AuthLocalUsers.Users = []*config.LocalUser{{
		Username: "bot",
		ApiKey:   "secret-api-key",
	}}

	req := httptest.NewRequest("POST", "/", nil)
	req.Header.Set("Authorization", "Bearer wrong")

	ctx := &authpublic.AuthCheckingContext{Request: req, Config: cfg}
	assert.Nil(t, checkUserFromLocalBearerApiKey(ctx))
}

func TestCheckUserFromLocalBearerApiKey_DisabledLocalUsers(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.AuthLocalUsers.Enabled = false
	cfg.AuthLocalUsers.Users = []*config.LocalUser{{
		Username: "bot",
		ApiKey:   "secret-api-key",
	}}

	req := httptest.NewRequest("POST", "/", nil)
	req.Header.Set("Authorization", "Bearer secret-api-key")

	ctx := &authpublic.AuthCheckingContext{Request: req, Config: cfg}
	assert.Nil(t, checkUserFromLocalBearerApiKey(ctx))
}

func TestCheckUserFromLocalBearerApiKey_NoBearerPrefix(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.AuthLocalUsers.Enabled = true
	cfg.AuthLocalUsers.Users = []*config.LocalUser{{
		Username: "bot",
		ApiKey:   "secret-api-key",
	}}

	req := httptest.NewRequest("POST", "/", nil)
	req.Header.Set("Authorization", "secret-api-key")

	ctx := &authpublic.AuthCheckingContext{Request: req, Config: cfg}
	assert.Nil(t, checkUserFromLocalBearerApiKey(ctx))
}
