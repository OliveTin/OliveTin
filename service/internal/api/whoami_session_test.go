package api

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiv1 "github.com/OliveTin/OliveTin/gen/olivetin/api/v1"
	apiv1connect "github.com/OliveTin/OliveTin/gen/olivetin/api/v1/apiv1connect"
	config "github.com/OliveTin/OliveTin/internal/config"
)

func loginLocalUserForSessionCookie(t *testing.T, client apiv1connect.OliveTinApiServiceClient) *http.Cookie {
	t.Helper()

	resp, err := client.LocalUserLogin(context.Background(), connect.NewRequest(&apiv1.LocalUserLoginRequest{
		Username: "alice",
		Password: "correct horse",
	}))
	require.NoError(t, err)
	require.True(t, resp.Msg.GetSuccess())

	header := http.Header{"Set-Cookie": resp.Header().Values("Set-Cookie")}
	for _, cookie := range (&http.Response{Header: header}).Cookies() {
		if cookie.Name == "olivetin-sid-local" && cookie.Value != "" {
			return cookie
		}
	}

	t.Fatal("login did not set the olivetin-sid-local cookie")
	return nil
}

func postWhoAmI(t *testing.T, baseURL string, cookie *http.Cookie) string {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, baseURL+"/api/WhoAmI", strings.NewReader("{}"))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { assert.NoError(t, resp.Body.Close()) }()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, string(body))

	return string(body)
}

func TestWhoAmINeverReturnsSessionID(t *testing.T) {
	hash, err := createHash("correct horse")
	require.NoError(t, err)

	cfg := config.DefaultConfig()
	cfg.SetDir(t.TempDir())
	cfg.AuthLocalUsers.Enabled = true
	cfg.AuthLocalUsers.Users = []*config.LocalUser{{Username: "alice", Password: hash}}

	ts, client := getNewTestServerAndClient(cfg)
	defer ts.Close()

	cookie := loginLocalUserForSessionCookie(t, client)
	body := postWhoAmI(t, ts.URL, cookie)

	assert.Contains(t, body, `"authenticatedUser":"alice"`)
	assert.NotContains(t, body, cookie.Value, "WhoAmI must not return the session cookie value")
	assert.NotContains(t, body, `"sid"`)
	assert.Contains(t, body, `"sessionFingerprint":"`)
}
