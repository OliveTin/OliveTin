package acl

import (
	"testing"

	authpublic "github.com/OliveTin/OliveTin/internal/auth/authpublic"
	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestIsAllowedViewDashboardAbsentAclsUnrestricted(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.DefaultPermissions.View = false

	dashboard := &config.DashboardComponent{
		Title: "Public",
		Contents: []*config.DashboardComponent{
			{Title: "Status", Type: "display"},
		},
	}

	guest := &authpublic.AuthenticatedUser{Username: "guest", Provider: "system"}
	guest.BuildUserAcls(cfg)

	assert.True(t, IsAllowedViewDashboard(cfg, guest, dashboard))
	assert.True(t, IsAllowedViewDashboard(cfg, guest, nil))
}

func TestIsAllowedViewDashboardAllowDenyAndDefaultFallback(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.DefaultPermissions.View = false
	cfg.AccessControlLists = []*config.AccessControlList{
		{
			Name:           "admins",
			MatchUsernames: []string{"admin"},
			Permissions:    config.PermissionsList{View: true, Exec: true},
		},
	}

	dashboard := &config.DashboardComponent{
		Title: "Services",
		Acls:  []string{"admins"},
		Contents: []*config.DashboardComponent{
			{Title: "Status: running", Type: "display"},
		},
	}

	guest := &authpublic.AuthenticatedUser{Username: "guest", Provider: "system"}
	guest.BuildUserAcls(cfg)
	admin := &authpublic.AuthenticatedUser{Username: "admin"}
	admin.BuildUserAcls(cfg)

	assert.False(t, IsAllowedViewDashboard(cfg, guest, dashboard))
	assert.True(t, IsAllowedViewDashboard(cfg, admin, dashboard))

	cfg.DefaultPermissions.View = true
	assert.True(t, IsAllowedViewDashboard(cfg, guest, dashboard),
		"when no relevant ACL matches, fall back to defaultPermissions.view")
}

func TestIsAllowedViewDashboardIgnoresAddToEveryAction(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.DefaultPermissions.View = false
	cfg.AccessControlLists = []*config.AccessControlList{
		{
			Name:             "admins",
			MatchUsernames:   []string{"admin"},
			AddToEveryAction: true,
			Permissions:      config.PermissionsList{View: true, Exec: true},
		},
	}

	dashboard := &config.DashboardComponent{
		Title: "Secret",
		Acls:  []string{"other"},
		Contents: []*config.DashboardComponent{
			{Title: "Hidden status", Type: "display"},
		},
	}

	admin := &authpublic.AuthenticatedUser{Username: "admin"}
	admin.BuildUserAcls(cfg)

	assert.False(t, IsAllowedViewDashboard(cfg, admin, dashboard),
		"AddToEveryAction must not grant dashboard view without listing the ACL on the dashboard")
}
