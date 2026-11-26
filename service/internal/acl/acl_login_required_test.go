package acl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	authpublic "github.com/OliveTin/OliveTin/internal/auth/authpublic"
	config "github.com/OliveTin/OliveTin/internal/config"
)

func guestMatchingAclConfig(requireLogin bool) (*config.Config, *config.Action) {
	cfg := config.DefaultConfig()
	cfg.AuthRequireGuestsToLogin = requireLogin
	cfg.AccessControlLists = []*config.AccessControlList{
		{
			Name:            "basic-users",
			MatchUsergroups: []string{"guest"},
			MatchUsernames:  []string{"guest"},
			Permissions:     config.PermissionsList{View: true, Exec: true, Logs: true, Kill: true},
		},
	}
	cfg.Sanitize()

	action := &config.Action{ID: "restart_web_server", Title: "Restart Web Server", Acls: []string{"basic-users"}}

	return cfg, action
}

func newGuest(cfg *config.Config) *authpublic.AuthenticatedUser {
	user := &authpublic.AuthenticatedUser{Username: "guest", UsergroupLine: "guest", Provider: "system"}
	user.BuildUserAcls(cfg)
	return user
}

func assertAllActionPermissions(t *testing.T, cfg *config.Config, user *authpublic.AuthenticatedUser, action *config.Action, expected bool) {
	t.Helper()
	assert.Equal(t, expected, IsAllowedView(cfg, user, action), "view")
	assert.Equal(t, expected, IsAllowedExec(cfg, user, action), "exec")
	assert.Equal(t, expected, IsAllowedLogs(cfg, user, action), "logs")
	assert.Equal(t, expected, IsAllowedKill(cfg, user, action), "kill")
}

// GHSA-9q82-wjpr-32jm: an ACL that names the guest user or usergroup must not
// give unauthenticated visitors any permission when login is required.
func TestGuestMatchingAclGrantsNothingWhenLoginRequired(t *testing.T) {
	cfg, action := guestMatchingAclConfig(true)
	guest := newGuest(cfg)

	assert.Empty(t, guest.Acls)
	assertAllActionPermissions(t, cfg, guest, action, false)
}

func TestGuestDeniedByAclCheckEvenWithAclNamesSet(t *testing.T) {
	cfg, action := guestMatchingAclConfig(true)
	guest := &authpublic.AuthenticatedUser{Username: "guest", UsergroupLine: "guest", Provider: "system", Acls: []string{"basic-users"}}

	assertAllActionPermissions(t, cfg, guest, action, false)
}

func TestGuestMatchingAclStillAppliesWhenLoginNotRequired(t *testing.T) {
	cfg, action := guestMatchingAclConfig(false)
	guest := newGuest(cfg)

	assert.Equal(t, []string{"basic-users"}, guest.Acls)
	assertAllActionPermissions(t, cfg, guest, action, true)
}

// The documented OAuth2 pattern gives logged-in users with no group claim the
// usergroup "guest". Those users are not unauthenticated visitors, so the ACL
// must keep working for them.
func TestLoggedInUserInGuestUsergroupKeepsAclWhenLoginRequired(t *testing.T) {
	cfg, action := guestMatchingAclConfig(true)
	user := &authpublic.AuthenticatedUser{Username: "alice", UsergroupLine: "guest", Provider: "oauth2"}
	user.BuildUserAcls(cfg)

	assert.Equal(t, []string{"basic-users"}, user.Acls)
	assertAllActionPermissions(t, cfg, user, action, true)
}
