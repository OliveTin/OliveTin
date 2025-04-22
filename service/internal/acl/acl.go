package acl

import (
	"context"
	"strings"

	config "github.com/OliveTin/OliveTin/internal/config"
	log "github.com/sirupsen/logrus"

	"golang.org/x/exp/slices"
	"google.golang.org/grpc/metadata"
)

type PermissionBits int

const (
	View PermissionBits = 1 << iota
	Exec
	Logs
)

func (p PermissionBits) Has(permission PermissionBits) bool {
	return p&permission != 0
}

// User respresents a person.
type AuthenticatedUser struct {
	Username      string
	UsergroupLine string

	Provider string
	SID      string

	Acls []string

	EffectivePolicy *config.ConfigurationPolicy
}

func (u *AuthenticatedUser) IsGuest() bool {
	return u.Username == "guest" && u.Provider == "system"
}

func (u *AuthenticatedUser) parseUsergroupLine(sep string) []string {
	ret := []string{}

	if sep != "" {
		for _, v := range strings.Split(u.UsergroupLine, sep) {
			trimmed := strings.TrimSpace(v)

			if trimmed != "" {
				ret = append(ret, trimmed)
			}
		}
	} else {
		ret = strings.Fields(u.UsergroupLine)
	}
	return ret
}

func (u *AuthenticatedUser) matchesUsergroupAcl(matchUsergroups []string, sep string) bool {
	groupList := u.parseUsergroupLine(sep)

	for _, group := range groupList {
		if slices.Contains(matchUsergroups, group) {
			log.Infof("Group %v found in %+v %v", group, groupList, len(groupList))
			return true
		}
	}

	return false
}

func logAclNotMatched(cfg *config.Config, aclFunction string, user *AuthenticatedUser, action *config.Action, acl *config.AccessControlList) {
	if cfg.LogDebugOptions.AclNotMatched {
		log.WithFields(log.Fields{
			"User":   user.Username,
			"Action": action.Title,
			"ACL":    acl.Name,
		}).Debugf("%v - ACL Not Matched", aclFunction)
	}
}

func logAclMatched(cfg *config.Config, aclFunction string, user *AuthenticatedUser, action *config.Action, acl *config.AccessControlList) {
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

func logAclNoneMatched(cfg *config.Config, aclFunction string, user *AuthenticatedUser, action *config.Action, defaultPermission bool) {
	if cfg.LogDebugOptions.AclNoneMatched {
		log.WithFields(log.Fields{
			"User":    user.Username,
			"Action":  action.Title,
			"Default": defaultPermission,
		}).Debugf("%v - No ACLs Matched, returning default permission", aclFunction)
	}
}

func permissionsConfigToBits(permissions config.PermissionsList) PermissionBits {
	var ret PermissionBits

	if permissions.View {
		ret |= View
	}

	if permissions.Exec {
		ret |= Exec
	}

	if permissions.Logs {
		ret |= Logs
	}

	return ret
}

func aclCheck(requiredPermission PermissionBits, defaultValue bool, cfg *config.Config, aclFunction string, user *AuthenticatedUser, action *config.Action) bool {
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
func IsAllowedLogs(cfg *config.Config, user *AuthenticatedUser, action *config.Action) bool {
	return aclCheck(Logs, cfg.DefaultPermissions.Logs, cfg, "isAllowedLogs", user, action)
}

// IsAllowedExec checks if a AuthenticatedUser is allowed to execute an Action
func IsAllowedExec(cfg *config.Config, user *AuthenticatedUser, action *config.Action) bool {
	return aclCheck(Exec, cfg.DefaultPermissions.Exec, cfg, "isAllowedExec", user, action)
}

// IsAllowedView checks if a User is allowed to view an Action
func IsAllowedView(cfg *config.Config, user *AuthenticatedUser, action *config.Action) bool {
	if action.Hidden {
		return false
	}

	return aclCheck(View, cfg.DefaultPermissions.View, cfg, "isAllowedView", user, action)
}

func getMetadataKeyOrEmpty(md metadata.MD, key string) string {
	mdValues := md.Get(key)

	if len(mdValues) > 0 {
		return mdValues[0]
	}

	return ""
}

// UserFromContext tries to find a user from a grpc context
func UserFromContext(ctx context.Context, cfg *config.Config) *AuthenticatedUser {
	var ret *AuthenticatedUser

	md, ok := metadata.FromIncomingContext(ctx)

	if ok {
		ret = &AuthenticatedUser{}
		ret.Username = getMetadataKeyOrEmpty(md, "username")
		ret.UsergroupLine = getMetadataKeyOrEmpty(md, "usergroup")
		ret.Provider = getMetadataKeyOrEmpty(md, "provider")

		buildUserAcls(cfg, ret)
	}

	if !ok || ret.Username == "" {
		ret = UserGuest(cfg)
	}

	log.WithFields(log.Fields{
		"username":      ret.Username,
		"usergroupLine": ret.UsergroupLine,
		"provider":      ret.Provider,
		"acls":          ret.Acls,
	}).Debugf("UserFromContext")

	return ret
}

func UserGuest(cfg *config.Config) *AuthenticatedUser {
	ret := &AuthenticatedUser{}
	ret.Username = "guest"
	ret.UsergroupLine = "guest"
	ret.Provider = "system"

	buildUserAcls(cfg, ret)

	return ret
}

func UserFromSystem(cfg *config.Config, username string) *AuthenticatedUser {
	ret := &AuthenticatedUser{
		Username:      username,
		UsergroupLine: "system",
		Provider:      "system",
	}

	buildUserAcls(cfg, ret)

	return ret
}

func buildUserAcls(cfg *config.Config, user *AuthenticatedUser) {
	for _, acl := range cfg.AccessControlLists {
		if slices.Contains(acl.MatchUsernames, user.Username) {
			user.Acls = append(user.Acls, acl.Name)
			continue
		}

		if user.matchesUsergroupAcl(acl.MatchUsergroups, cfg.AuthHttpHeaderUserGroupSep) {
			user.Acls = append(user.Acls, acl.Name)
			continue
		}
	}

	user.EffectivePolicy = getEffectivePolicy(cfg, user)
}

func isACLRelevantToAction(cfg *config.Config, actionAcls []string, acl *config.AccessControlList, user *AuthenticatedUser) bool {
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

func getRelevantAcls(cfg *config.Config, actionAcls []string, user *AuthenticatedUser) []*config.AccessControlList {
	var ret []*config.AccessControlList

	for _, acl := range cfg.AccessControlLists {
		if isACLRelevantToAction(cfg, actionAcls, acl, user) {
			ret = append(ret, acl)
		}
	}

	return ret
}

func getEffectivePolicy(cfg *config.Config, user *AuthenticatedUser) *config.ConfigurationPolicy {
	ret := &config.ConfigurationPolicy{
		ShowDiagnostics: cfg.DefaultPolicy.ShowDiagnostics,
		ShowLogList:     cfg.DefaultPolicy.ShowLogList,
	}

	for _, acl := range cfg.AccessControlLists {
		if slices.Contains(user.Acls, acl.Name) {
			logAclMatched(cfg, "GetEffectivePolicy", user, nil, acl)

			ret = buildConfigurationPolicy(ret, acl.Policy)
		}
	}

	return ret
}

func buildConfigurationPolicy(ret *config.ConfigurationPolicy, policy config.ConfigurationPolicy) *config.ConfigurationPolicy {
	if policy.ShowDiagnostics {
		ret.ShowDiagnostics = policy.ShowDiagnostics
	}

	if policy.ShowLogList {
		ret.ShowLogList = policy.ShowLogList
	}

	return ret
}
