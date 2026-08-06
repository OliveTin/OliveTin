package acl

import (
	"testing"

	authpublic "github.com/OliveTin/OliveTin/internal/auth/authpublic"
	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestIsAllowedViewEntityTypeAbsentAclsUnrestricted(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.DefaultPermissions.View = false

	entityFile := &config.EntityFile{
		Name: "printers",
		File: "printers.yaml",
	}

	guest := &authpublic.AuthenticatedUser{Username: "guest", Provider: "system"}
	guest.BuildUserAcls(cfg)

	assert.True(t, IsAllowedViewEntityType(cfg, guest, entityFile))
	assert.True(t, IsAllowedViewEntityType(cfg, guest, nil))
}

func TestIsAllowedViewEntityTypeAllowDenyAndDefaultFallback(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.DefaultPermissions.View = false
	cfg.AccessControlLists = []*config.AccessControlList{
		{
			Name:           "ops",
			MatchUsernames: []string{"admin"},
			Permissions:    config.PermissionsList{View: true, Exec: true},
		},
	}

	entityFile := &config.EntityFile{
		Name: "servers",
		File: "servers.yaml",
		Acls: []string{"ops"},
	}

	guest := &authpublic.AuthenticatedUser{Username: "guest", Provider: "system"}
	guest.BuildUserAcls(cfg)
	admin := &authpublic.AuthenticatedUser{Username: "admin"}
	admin.BuildUserAcls(cfg)

	assert.False(t, IsAllowedViewEntityType(cfg, guest, entityFile))
	assert.True(t, IsAllowedViewEntityType(cfg, admin, entityFile))

	cfg.DefaultPermissions.View = true
	assert.True(t, IsAllowedViewEntityType(cfg, guest, entityFile),
		"when no relevant ACL matches, fall back to defaultPermissions.view")
}

func TestIsAllowedViewEntityTypeIgnoresAddToEveryAction(t *testing.T) {
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

	entityFile := &config.EntityFile{
		Name: "secret",
		File: "secret.yaml",
		Acls: []string{"other"},
	}

	admin := &authpublic.AuthenticatedUser{Username: "admin"}
	admin.BuildUserAcls(cfg)

	assert.False(t, IsAllowedViewEntityType(cfg, admin, entityFile),
		"AddToEveryAction must not grant entity view without listing the ACL on the entity")
}
