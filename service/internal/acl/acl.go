package acl

import (
	authpublic "github.com/OliveTin/OliveTin/internal/auth/authpublic"
	config "github.com/OliveTin/OliveTin/internal/config"
	log "github.com/sirupsen/logrus"

	"golang.org/x/exp/slices"
)

type PermissionBits int

const (
	View PermissionBits = 1 << iota
	Exec
	Logs
	Kill
)

func (p PermissionBits) Has(permission PermissionBits) bool {
	return p&permission != 0
}

func logAclNotMatched(cfg *config.Config, aclFunction string, user *authpublic.AuthenticatedUser, resourceTitle string, acl *config.AccessControlList) {
	if cfg.LogDebugOptions.AclNotMatched {
		log.WithFields(log.Fields{
			"User":     user.Username,
			"Resource": resourceTitle,
			"ACL":      acl.Name,
		}).Debugf("%v - ACL Not Matched", aclFunction)
	}
}

func logAclMatched(cfg *config.Config, aclFunction string, user *authpublic.AuthenticatedUser, resourceTitle string, acl *config.AccessControlList) {
	if resourceTitle == "" {
		resourceTitle = "N/A"
	}

	if cfg.LogDebugOptions.AclMatched {
		log.WithFields(log.Fields{
			"User":     user.Username,
			"Resource": resourceTitle,
			"ACL":      acl.Name,
		}).Debugf("%v - Matched ACL", aclFunction)
	}
}

func logAclNoneMatched(cfg *config.Config, aclFunction string, user *authpublic.AuthenticatedUser, resourceTitle string, defaultPermission bool) {
	if cfg.LogDebugOptions.AclNoneMatched {
		log.WithFields(log.Fields{
			"User":     user.Username,
			"Resource": resourceTitle,
			"Default":  defaultPermission,
		}).Debugf("%v - No ACLs Matched, returning default permission", aclFunction)
	}
}

func permissionsConfigToBits(permissions config.PermissionsList) PermissionBits {
	type permPair struct {
		enabled bool
		bit     PermissionBits
	}

	permMap := []permPair{
		{permissions.View, View},
		{permissions.Exec, Exec},
		{permissions.Logs, Logs},
		{permissions.Kill, Kill},
	}

	var ret PermissionBits

	for _, perm := range permMap {
		if perm.enabled {
			ret |= perm.bit
		}
	}

	return ret
}

func aclCheck(requiredPermission PermissionBits, defaultValue bool, cfg *config.Config, aclFunction string, user *authpublic.AuthenticatedUser, resourceTitle string, resourceAcls []string, includeAddToEvery bool) bool {
	relevantAcls := getRelevantAcls(cfg, resourceAcls, user, includeAddToEvery)

	if cfg.LogDebugOptions.AclCheckStarted {
		log.WithFields(log.Fields{
			"resourceTitle":      resourceTitle,
			"username":           user.Username,
			"usergroupLine":      user.UsergroupLine,
			"relevantAcls":       len(relevantAcls),
			"requiredPermission": requiredPermission,
		}).Debugf("ACL check - %v", aclFunction)
	}

	for _, acl := range relevantAcls {
		permissionBits := permissionsConfigToBits(acl.Permissions)

		if permissionBits.Has(requiredPermission) {
			logAclMatched(cfg, aclFunction, user, resourceTitle, acl)

			return true
		}

		logAclNotMatched(cfg, aclFunction, user, resourceTitle, acl)
	}

	logAclNoneMatched(cfg, aclFunction, user, resourceTitle, defaultValue)

	return defaultValue
}

// IsAllowedLogs checks if a AuthenticatedUser is allowed to view an action's logs
func IsAllowedLogs(cfg *config.Config, user *authpublic.AuthenticatedUser, action *config.Action) bool {
	return aclCheck(Logs, cfg.DefaultPermissions.Logs, cfg, "isAllowedLogs", user, action.Title, action.Acls, true)
}

// IsAllowedExec checks if a AuthenticatedUser is allowed to execute an Action
func IsAllowedExec(cfg *config.Config, user *authpublic.AuthenticatedUser, action *config.Action) bool {
	return aclCheck(Exec, cfg.DefaultPermissions.Exec, cfg, "isAllowedExec", user, action.Title, action.Acls, true)
}

// IsAllowedView checks if a User is allowed to view an Action
func IsAllowedView(cfg *config.Config, user *authpublic.AuthenticatedUser, action *config.Action) bool {
	if action.Hidden {
		return false
	}

	return aclCheck(View, cfg.DefaultPermissions.View, cfg, "isAllowedView", user, action.Title, action.Acls, true)
}

func IsAllowedKill(cfg *config.Config, user *authpublic.AuthenticatedUser, action *config.Action) bool {
	return aclCheck(Kill, cfg.DefaultPermissions.Kill, cfg, "isAllowedKill", user, action.Title, action.Acls, true)
}

// IsAllowedViewDashboard checks if a user may see a root dashboard.
// Dashboards with no acls are unrestricted. AddToEveryAction does not apply.
func IsAllowedViewDashboard(cfg *config.Config, user *authpublic.AuthenticatedUser, dashboard *config.DashboardComponent) bool {
	if dashboard == nil || len(dashboard.Acls) == 0 {
		return true
	}

	return aclCheck(View, cfg.DefaultPermissions.View, cfg, "isAllowedViewDashboard", user, dashboard.Title, dashboard.Acls, false)
}

func isACLRelevant(resourceAcls []string, acl *config.AccessControlList, user *authpublic.AuthenticatedUser, includeAddToEvery bool) bool {
	if !slices.Contains(user.Acls, acl.Name) {
		return false
	}

	if includeAddToEvery && acl.AddToEveryAction {
		return true
	}

	return slices.Contains(resourceAcls, acl.Name)
}

func getRelevantAcls(cfg *config.Config, resourceAcls []string, user *authpublic.AuthenticatedUser, includeAddToEvery bool) []*config.AccessControlList {
	var ret []*config.AccessControlList

	for _, acl := range cfg.AccessControlLists {
		if isACLRelevant(resourceAcls, acl, user, includeAddToEvery) {
			ret = append(ret, acl)
		}
	}

	return ret
}
