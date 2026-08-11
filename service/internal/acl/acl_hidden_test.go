package acl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	authpublic "github.com/OliveTin/OliveTin/internal/auth/authpublic"
	config "github.com/OliveTin/OliveTin/internal/config"
)

func TestIsAllowedViewIgnoresHiddenFlag(t *testing.T) {
	cfg := config.DefaultConfig()
	guest := &authpublic.AuthenticatedUser{Username: "guest", Provider: "system"}
	guest.BuildUserAcls(cfg)

	visible := &config.Action{Title: "Visible", Shell: "echo"}
	hidden := &config.Action{Title: "Hidden", Shell: "echo", Hidden: true}

	assert.True(t, IsAllowedView(cfg, guest, visible))
	assert.True(t, IsAllowedView(cfg, guest, hidden),
		"hidden must not deny view; use ACLs to restrict access")
}

func TestIsAllowedViewHiddenStillRespectsAcl(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.DefaultPermissions.View = false
	cfg.AccessControlLists = []*config.AccessControlList{
		{
			Name:             "admins",
			MatchUsernames:   []string{"admin"},
			AddToEveryAction: true,
			Permissions:      config.PermissionsList{View: true},
		},
	}

	admin := &authpublic.AuthenticatedUser{Username: "admin"}
	admin.BuildUserAcls(cfg)
	guest := &authpublic.AuthenticatedUser{Username: "guest", Provider: "system"}
	guest.BuildUserAcls(cfg)

	hidden := &config.Action{Title: "Webhook", Shell: "echo", Hidden: true}

	assert.True(t, IsAllowedView(cfg, admin, hidden))
	assert.False(t, IsAllowedView(cfg, guest, hidden))
}
