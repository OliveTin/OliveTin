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

func logAclNotMatched(cfg *config.Config, aclFunction string, user *authpublic.AuthenticatedUser, action *config.Action, acl *config.AccessControlList) {
	if cfg.LogDebugOptions.AclNotMatched {
		log.WithFields(log.Fields{
			"User":   user.Username,
			"Action": action.Title,
			"ACL":    acl.Name,
		}).Debugf("%v - ACL Not Matched", aclFunction)
	}
}

func logAclMatched(cfg *config.Config, aclFunction string, user *authpublic.AuthenticatedUser, action *config.Action, acl *config.AccessControlList) {
	actionTitle := "N/A"

	if action != nil {
		actionTitle = action.Title
	}

	if cfg.LogDebugOptions.AclMatched {
		log.WithFields(log.Fields{
			"User":   user.Username,
			"Action": actionTitle,
			"ACL":    acl.Name,
		}).Debugf("%v - Matched ACL", aclFunction)
	}
}

func logAclNoneMatched(cfg *config.Config, aclFunction string, user *authpublic.AuthenticatedUser, action *config.Action, defaultPermission bool) {
	if cfg.LogDebugOptions.AclNoneMatched {
		log.WithFields(log.Fields{
			"User":    user.Username,
			"Action":  action.Title,
			"Default": defaultPermission,
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

func aclCheck(requiredPermission PermissionBits, defaultValue bool, cfg *config.Config, aclFunction string, user *authpublic.AuthenticatedUser, action *config.Action) bool {
	relevantAcls := getRelevantAcls(cfg, action.Acls, user)

	if cfg.LogDebugOptions.AclCheckStarted {
		log.WithFields(log.Fields{
			"actionTitle":        action.Title,
			"username":           user.Username,
			"usergroupLine":      user.UsergroupLine,
			"relevantAcls":       len(relevantAcls),
			"requiredPermission": requiredPermission,
		}).Debugf("ACL check - %v", aclFunction)
	}

	for _, acl := range relevantAcls {
		permissionBits := permissionsConfigToBits(acl.Permissions)

		if permissionBits.Has(requiredPermission) {
			logAclMatched(cfg, aclFunction, user, action, acl)

			return true
		} else {
			logAclNotMatched(cfg, aclFunction, user, action, acl)
		}
	}

	logAclNoneMatched(cfg, aclFunction, user, action, cfg.DefaultPermissions.Logs)

	return defaultValue
}

// IsAllowedLogs checks if a AuthenticatedUser is allowed to view an action's logs
func IsAllowedLogs(cfg *config.Config, user *authpublic.AuthenticatedUser, action *config.Action) bool {
	return aclCheck(Logs, cfg.DefaultPermissions.Logs, cfg, "isAllowedLogs", user, action)
}

// IsAllowedExec checks if a AuthenticatedUser is allowed to execute an Action
func IsAllowedExec(cfg *config.Config, user *authpublic.AuthenticatedUser, action *config.Action) bool {
	return aclCheck(Exec, cfg.DefaultPermissions.Exec, cfg, "isAllowedExec", user, action)
}

// IsAllowedView checks if a User is allowed to view an Action
func IsAllowedView(cfg *config.Config, user *authpublic.AuthenticatedUser, action *config.Action) bool {
	if action.Hidden {
		return false
	}

	return aclCheck(View, cfg.DefaultPermissions.View, cfg, "isAllowedView", user, action)
}

func IsAllowedKill(cfg *config.Config, user *authpublic.AuthenticatedUser, action *config.Action) bool {
	return aclCheck(Kill, cfg.DefaultPermissions.Kill, cfg, "isAllowedKill", user, action)
}

func isACLRelevantToAction(cfg *config.Config, actionAcls []string, acl *config.AccessControlList, user *authpublic.AuthenticatedUser) bool {
	if !slices.Contains(user.Acls, acl.Name) {
		// If the user does not have this ACL, then it is not relevant

		return false
	}

	if acl.AddToEveryAction {
		return true
	}

	if slices.Contains(actionAcls, acl.Name) {
		return true
	}

	return false
}

func getRelevantAcls(cfg *config.Config, actionAcls []string, user *authpublic.AuthenticatedUser) []*config.AccessControlList {
	var ret []*config.AccessControlList

	for _, acl := range cfg.AccessControlLists {
		if isACLRelevantToAction(cfg, actionAcls, acl, user) {
			ret = append(ret, acl)
		}
	}

	return ret
}
