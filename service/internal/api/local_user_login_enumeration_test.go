package api

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiv1 "github.com/OliveTin/OliveTin/gen/olivetin/api/v1"
	apiv1connect "github.com/OliveTin/OliveTin/gen/olivetin/api/v1/apiv1connect"
	config "github.com/OliveTin/OliveTin/internal/config"
)

func newEnumerationTestClient(t *testing.T) apiv1connect.OliveTinApiServiceClient {
	t.Helper()

	hash, err := createHash("correct horse")
	require.NoError(t, err)

	cfg := config.DefaultConfig()
	cfg.SetDir(t.TempDir())
	cfg.AuthLocalUsers.Enabled = true
	cfg.AuthLocalUsers.Users = []*config.LocalUser{
		{Username: "alice", Password: hash},
		{Username: "apikeyonly", ApiKey: "k"},
	}

	ts, client := getNewTestServerAndClient(cfg)
	t.Cleanup(ts.Close)

	return client
}

func loginAttempt(client apiv1connect.OliveTinApiServiceClient, username string, password string) (*connect.Response[apiv1.LocalUserLoginResponse], error) {
	return client.LocalUserLogin(context.Background(), connect.NewRequest(&apiv1.LocalUserLoginRequest{
		Username: username,
		Password: password,
	}))
}

func fillArgon2Slots(t *testing.T) {
	t.Helper()

	for range argon2MaxConcurrent {
		argon2Sem <- struct{}{}
	}

	t.Cleanup(func() {
		for range argon2MaxConcurrent {
			<-argon2Sem
		}
	})
}

// Not parallel: this test holds every Argon2 slot while it runs.
func TestLocalUserLoginIsBusyForEveryUsernameWhenArgon2IsBusy(t *testing.T) {
	client := newEnumerationTestClient(t)
	_, err := getDummyPasswordHash()
	require.NoError(t, err)

	fillArgon2Slots(t)

	for _, username := range []string{"alice", "nosuchuser", "apikeyonly"} {
		_, err := loginAttempt(client, username, "wrong password")
		require.Error(t, err, username)
		assert.Equal(t, connect.CodeResourceExhausted, connect.CodeOf(err), "username %q must get the same busy error as every other username", username)
	}
}

func TestLocalUserLoginFailsTheSameForUnknownAndPasswordlessUsers(t *testing.T) {
	client := newEnumerationTestClient(t)

	for _, username := range []string{"alice", "nosuchuser", "apikeyonly"} {
		resp, err := loginAttempt(client, username, "wrong password")
		require.NoError(t, err, username)
		assert.False(t, resp.Msg.GetSuccess(), username)
		assert.Empty(t, resp.Header().Values("Set-Cookie"), username)
	}
}

func TestLocalUserLoginRejectsDummyPasswordForUnknownAndPasswordlessUsers(t *testing.T) {
	client := newEnumerationTestClient(t)
	_, err := getDummyPasswordHash()
	require.NoError(t, err)
	require.NotEmpty(t, dummyPassword)

	for _, username := range []string{"nosuchuser", "apikeyonly"} {
		resp, err := loginAttempt(client, username, dummyPassword)
		require.NoError(t, err, username)
		assert.False(t, resp.Msg.GetSuccess(), "the dummy password must never log anyone in")
	}
}

func TestLocalUserLoginStillAcceptsTheRightPassword(t *testing.T) {
	client := newEnumerationTestClient(t)

	resp, err := loginAttempt(client, "alice", "correct horse")
	require.NoError(t, err)
	assert.True(t, resp.Msg.GetSuccess())
}
