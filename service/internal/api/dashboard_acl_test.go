package api

import (
	"testing"

	authpublic "github.com/OliveTin/OliveTin/internal/auth/authpublic"
	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/executor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildDashboardAclTestConfig() *config.Config {
	cfg := config.DefaultConfig()
	cfg.DefaultPermissions.View = false
	cfg.DefaultPermissions.Exec = false
	cfg.AccessControlLists = []*config.AccessControlList{
		{
			Name:           "admins",
			MatchUsernames: []string{"admin"},
			Permissions:    config.PermissionsList{View: true, Exec: true},
		},
	}
	cfg.Dashboards = []*config.DashboardComponent{
		{
			Title: "Public tools",
			Contents: []*config.DashboardComponent{
				{Title: "Welcome", Type: "display"},
			},
		},
		{
			Title: "Services",
			Acls:  []string{"admins"},
			Contents: []*config.DashboardComponent{
				{Title: "Status: running", Type: "display"},
			},
		},
	}

	return cfg
}

func TestDashboardAclsRootNavAndGetDashboard(t *testing.T) {
	cfg := buildDashboardAclTestConfig()
	ex := executor.DefaultExecutor(cfg)
	api := newServer(ex)

	guest := &authpublic.AuthenticatedUser{Username: "guest", Provider: "system"}
	guest.BuildUserAcls(cfg)
	admin := &authpublic.AuthenticatedUser{Username: "admin"}
	admin.BuildUserAcls(cfg)

	guestRoots := api.buildRootDashboards(guest, cfg.Dashboards)
	assert.Contains(t, guestRoots, "Public tools")
	assert.NotContains(t, guestRoots, "Services")

	adminRoots := api.buildRootDashboards(admin, cfg.Dashboards)
	assert.Contains(t, adminRoots, "Public tools")
	assert.Contains(t, adminRoots, "Services")

	guestRR := api.createDashboardRenderRequest(guest, "", "")
	assert.Nil(t, renderDashboard(guestRR, "Services"),
		"GetDashboard must not leak ACL-restricted dashboard content via deep link")

	adminRR := api.createDashboardRenderRequest(admin, "", "")
	db := renderDashboard(adminRR, "Services")
	require.NotNil(t, db)
	assert.Equal(t, "Services", db.Title)
}

func TestDashboardAclsNestedDirectoryDeepLink(t *testing.T) {
	cfg := buildDashboardAclTestConfig()
	cfg.Dashboards = []*config.DashboardComponent{
		{
			Title: "Public tools",
			Contents: []*config.DashboardComponent{
				{Title: "Welcome", Type: "display"},
			},
		},
		{
			Title: "Services",
			Acls:  []string{"admins"},
			Contents: []*config.DashboardComponent{
				{
					Title: "Infrastructure",
					Contents: []*config.DashboardComponent{
						{Title: "Status: running", Type: "display"},
					},
				},
			},
		},
	}

	ex := executor.DefaultExecutor(cfg)
	api := newServer(ex)

	guest := &authpublic.AuthenticatedUser{Username: "guest", Provider: "system"}
	guest.BuildUserAcls(cfg)
	admin := &authpublic.AuthenticatedUser{Username: "admin"}
	admin.BuildUserAcls(cfg)

	guestRR := api.createDashboardRenderRequest(guest, "", "")
	assert.Nil(t, renderDashboard(guestRR, "Infrastructure"),
		"nested directory under ACL-restricted root must not leak via deep link")

	adminRR := api.createDashboardRenderRequest(admin, "", "")
	db := renderDashboard(adminRR, "Infrastructure")
	require.NotNil(t, db)
	assert.Equal(t, "Infrastructure", db.Title)
}
