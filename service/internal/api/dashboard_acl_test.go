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

func TestRootDashboardEntriesIncludeCategory(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Dashboards = []*config.DashboardComponent{
		{
			Title:    "Misc Tools",
			Contents: []*config.DashboardComponent{{Title: "Hello", Type: "display"}},
		},
		{
			Title:    "My Servers",
			Category: "Infrastructure",
			Contents: []*config.DashboardComponent{{Title: "Ping", Type: "display"}},
		},
		{
			Title:    "Status Board",
			Category: "Monitoring",
			Contents: []*config.DashboardComponent{{Title: "Uptime", Type: "display"}},
		},
		{
			Title:    "My Containers",
			Category: "Infrastructure",
			Contents: []*config.DashboardComponent{{Title: "Restart", Type: "display"}},
		},
	}

	ex := executor.DefaultExecutor(cfg)
	api := newServer(ex)
	user := &authpublic.AuthenticatedUser{Username: "guest", Provider: "system"}
	user.BuildUserAcls(cfg)

	entries := api.buildRootDashboardEntries(user, cfg.Dashboards)
	require.Len(t, entries, 4)
	assert.Equal(t, []string{"Misc Tools", "My Servers", "Status Board", "My Containers"}, rootDashboardTitles(entries))
	assert.Equal(t, "", entries[0].Category)
	assert.Equal(t, "Infrastructure", entries[1].Category)
	assert.Equal(t, "Monitoring", entries[2].Category)
	assert.Equal(t, "Infrastructure", entries[3].Category)
}

func TestRootDashboardEntriesOmitAclHiddenCategories(t *testing.T) {
	cfg := buildDashboardAclTestConfig()
	cfg.Dashboards[0].Category = "Public"
	cfg.Dashboards[1].Category = "Admin only"

	ex := executor.DefaultExecutor(cfg)
	api := newServer(ex)

	guest := &authpublic.AuthenticatedUser{Username: "guest", Provider: "system"}
	guest.BuildUserAcls(cfg)

	entries := api.buildRootDashboardEntries(guest, cfg.Dashboards)
	require.Len(t, entries, 1)
	assert.Equal(t, "Public tools", entries[0].Title)
	assert.Equal(t, "Public", entries[0].Category)
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
