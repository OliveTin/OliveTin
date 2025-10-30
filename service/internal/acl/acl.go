package acl

import (
	"context"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	"github.com/OliveTin/OliveTin/internal/auth"
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

	log.Debugf("parseUsergroupLine: %v, %v, sep:%v", u.UsergroupLine, ret, sep)

	return ret
}

func (u *AuthenticatedUser) matchesUsergroupAcl(matchUsergroups []string, sep string) bool {
	groupList := u.parseUsergroupLine(sep)

	for _, group := range groupList {
		if slices.Contains(matchUsergroups, group) {
			log.Debugf("Usergroup %v found in %+v (len: %v)", group, groupList, len(groupList))
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

func IsAllowedKill(cfg *config.Config, user *AuthenticatedUser, action *config.Action) bool {
	return aclCheck(Kill, cfg.DefaultPermissions.Kill, cfg, "isAllowedKill", user, action)
}

func getHeaderKeyOrEmpty(headers http.Header, key string) string {
	values := headers.Values(key)
	if len(values) > 0 {
		return values[0]
	}
	return ""
}

// UserFromContext tries to find a user from a Connect RPC context
func UserFromContext[T any](ctx context.Context, req *connect.Request[T], cfg *config.Config) *AuthenticatedUser {
	var ret *AuthenticatedUser

	if req != nil {
		ret = &AuthenticatedUser{}
		// Only trust headers if explicitly configured
		if cfg.AuthHttpHeaderUsername != "" {
			ret.Username = getHeaderKeyOrEmpty(req.Header(), cfg.AuthHttpHeaderUsername)
		}

		if cfg.AuthHttpHeaderUserGroup != "" {
			ret.UsergroupLine = getHeaderKeyOrEmpty(req.Header(), cfg.AuthHttpHeaderUserGroup)
		}
		// Optional provider header; otherwise infer below
		prov := getHeaderKeyOrEmpty(req.Header(), "provider")
		if prov != "" {
			ret.Provider = prov
		}

		// If no username from headers, fall back to local session cookie
		if ret.Username == "" {
			// Build a minimal http.Request to parse cookies from headers
			dummy := &http.Request{Header: req.Header()}
			if c, err := dummy.Cookie("olivetin-sid-local"); err == nil && c != nil && c.Value != "" {
				if sess := auth.GetUserSession("local", c.Value); sess != nil {
					if u := cfg.FindUserByUsername(sess.Username); u != nil {
						ret.Username = u.Username
						ret.UsergroupLine = u.Usergroup
						ret.Provider = "local"
						ret.SID = c.Value
					} else {
						log.WithFields(log.Fields{"username": sess.Username}).Warn("UserFromContext: local session user not in config")
					}
				} else {
					log.WithFields(log.Fields{"sid": c.Value, "provider": "local"}).Warn("UserFromContext: stale local session")
				}
			}
		}

		if ret.Username != "" {
			buildUserAcls(cfg, ret)
		}
	}

	if ret == nil || ret.Username == "" {
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
