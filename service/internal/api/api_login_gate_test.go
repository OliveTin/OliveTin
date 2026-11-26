package api

import (
	"context"
	"net/http"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/emptypb"

	apiv1 "github.com/OliveTin/OliveTin/gen/olivetin/api/v1"
	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/executor"
)

// The procedures that unauthenticated visitors may call when login is required.
// This list is deliberately kept separate from loginNotRequiredProcedures, so
// that making a procedure public needs both to be changed on purpose.
var expectedPublicProcedures = []string{
	"Init",
	"GetReadyz",
	"LocalUserLogin",
	"Logout",
	"PasswordHash",
}

// loginRequiredGuestAclConfig is the configuration from GHSA-9q82-wjpr-32jm:
// login is required, and an ACL that matches the "guest" usergroup grants
// every permission on the action.
func loginRequiredGuestAclConfig() *config.Config {
	cfg := config.DefaultConfig()
	cfg.AuthRequireGuestsToLogin = true
	cfg.AccessControlLists = []*config.AccessControlList{
		{
			Name:            "basic-users",
			MatchUsergroups: []string{"guest"},
			Permissions:     config.PermissionsList{View: true, Exec: true, Logs: true, Kill: true},
		},
	}
	cfg.Actions = append(cfg.Actions, &config.Action{
		ID:    "restart_web_server",
		Title: "Restart Web Server",
		Shell: "echo restarted",
		Acls:  []string{"basic-users"},
	})
	cfg.Sanitize()

	return cfg
}

func oliveTinServiceDescriptor() protoreflect.ServiceDescriptor {
	return apiv1.File_olivetin_api_v1_olivetin_proto.Services().ByName("OliveTinApiService")
}

func procedureForMethod(method protoreflect.MethodDescriptor) string {
	return "/" + string(method.Parent().FullName()) + "/" + string(method.Name())
}

// callAsGuest calls any procedure with an empty message and no credentials.
// An empty protobuf message is valid for every request type.
func callAsGuest(baseURL string, method protoreflect.MethodDescriptor) error {
	client := connect.NewClient[emptypb.Empty, emptypb.Empty](http.DefaultClient, baseURL+"/api/"+string(method.Name()))

	if !method.IsStreamingServer() {
		_, err := client.CallUnary(context.Background(), connect.NewRequest(&emptypb.Empty{}))
		return err
	}

	stream, err := client.CallServerStream(context.Background(), connect.NewRequest(&emptypb.Empty{}))
	if err != nil {
		return err
	}
	received := stream.Receive()
	closeErr := stream.Close()

	if received {
		return closeErr
	}

	return stream.Err()
}

func TestPublicProcedureListsMatch(t *testing.T) {
	expected := make(map[string]bool)
	for _, name := range expectedPublicProcedures {
		method := oliveTinServiceDescriptor().Methods().ByName(protoreflect.Name(name))
		require.NotNil(t, method, "public procedure %s does not exist in the proto service", name)
		expected[procedureForMethod(method)] = true
	}

	assert.Equal(t, expected, loginNotRequiredProcedures)
}

// TestEveryProcedureRequiresLoginUnlessPublic (GHSA-9q82-wjpr-32jm) calls every
// procedure in the proto service as an unauthenticated visitor, so a procedure
// added later is covered without anyone having to remember to add a test.
func TestEveryProcedureRequiresLoginUnlessPublic(t *testing.T) {
	cfg := loginRequiredGuestAclConfig()
	ts, _ := getNewTestServerAndClient(cfg)
	defer ts.Close()

	methods := oliveTinServiceDescriptor().Methods()
	require.Positive(t, methods.Len())

	for i := 0; i < methods.Len(); i++ {
		method := methods.Get(i)

		t.Run(string(method.Name()), func(t *testing.T) {
			err := callAsGuest(ts.URL, method)

			if loginNotRequiredProcedures[procedureForMethod(method)] {
				assert.NotEqual(t, connect.CodePermissionDenied, connect.CodeOf(err), "public procedure must be callable without logging in")
				return
			}

			require.Error(t, err)
			assert.Equal(t, connect.CodePermissionDenied, connect.CodeOf(err), "procedure must refuse unauthenticated visitors when login is required")
		})
	}
}

func TestGuestCannotStartActionsWhenLoginRequired(t *testing.T) {
	cfg := loginRequiredGuestAclConfig()
	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()
	ts, client := getNewTestServerAndClientWithExecutor(cfg, ex)
	defer ts.Close()

	_, err := client.StartAction(context.Background(), connect.NewRequest(&apiv1.StartActionRequest{BindingId: "restart_web_server"}))
	assert.Equal(t, connect.CodePermissionDenied, connect.CodeOf(err))

	_, err = client.StartActionAndWait(context.Background(), connect.NewRequest(&apiv1.StartActionAndWaitRequest{ActionId: "restart_web_server"}))
	assert.Equal(t, connect.CodePermissionDenied, connect.CodeOf(err))

	binding := ex.FindBindingByID("restart_web_server")
	require.NotNil(t, binding)
	assert.Empty(t, ex.GetLogsByBindingId(binding.ID), "the action must not have been requested at all")
}

// Real and made-up action IDs must get the same answer, so that unauthenticated
// visitors cannot find out which actions exist.
func TestGuestCannotTellRealActionIdsFromFakeOnes(t *testing.T) {
	cfg := loginRequiredGuestAclConfig()
	ts, client := getNewTestServerAndClient(cfg)
	defer ts.Close()

	_, realErr := client.StartAction(context.Background(), connect.NewRequest(&apiv1.StartActionRequest{BindingId: "restart_web_server"}))
	_, fakeErr := client.StartAction(context.Background(), connect.NewRequest(&apiv1.StartActionRequest{BindingId: "does_not_exist_1234"}))

	require.Error(t, realErr)
	require.Error(t, fakeErr)
	assert.Equal(t, connect.CodeOf(realErr), connect.CodeOf(fakeErr))
	assert.Equal(t, realErr.Error(), fakeErr.Error())
}

func TestGuestCanStillStartActionsWhenLoginNotRequired(t *testing.T) {
	cfg := loginRequiredGuestAclConfig()
	cfg.AuthRequireGuestsToLogin = false
	ts, client := getNewTestServerAndClient(cfg)
	defer ts.Close()

	res, err := client.StartActionAndWait(context.Background(), connect.NewRequest(&apiv1.StartActionAndWaitRequest{ActionId: "restart_web_server"}))
	require.NoError(t, err)
	assert.Contains(t, res.Msg.LogEntry.Output, "restarted")
}
