package api

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiv1 "github.com/OliveTin/OliveTin/gen/olivetin/api/v1"
	config "github.com/OliveTin/OliveTin/internal/config"
)

func TestLocalUserLoginRejectsUserWithNoPassword(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.AuthLocalUsers.Enabled = true
	cfg.AuthLocalUsers.Users = []*config.LocalUser{{
		Username: "onlykey",
		ApiKey:   "k",
		Password: "",
	}}

	ts, client := getNewTestServerAndClient(cfg)
	defer ts.Close()

	resp, err := client.LocalUserLogin(context.Background(), connect.NewRequest(&apiv1.LocalUserLoginRequest{
		Username: "onlykey",
		Password: "anything",
	}))
	require.NoError(t, err)
	assert.False(t, resp.Msg.GetSuccess())
}
