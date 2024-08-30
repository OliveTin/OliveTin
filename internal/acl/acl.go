package acl

import (
	"context"
	config "github.com/OliveTin/OliveTin/internal/config"
	log "github.com/sirupsen/logrus"

	"golang.org/x/exp/slices"
	"google.golang.org/grpc/metadata"
)

// User respresents a person.
type AuthenticatedUser struct {
	Username  string
	Usergroup string

	acls []string
}

func logAclNotMatched(cfg *config.Config, aclFunction string, user *AuthenticatedUser, action *config.Action) {
	if cfg.LogDebugOptions.AclNotMatched {
		log.WithFields(log.Fields{
			"User":   user.Username,
			"Action": action.Title,
		}).Debugf("%v - No ACLs Matched", aclFunction)
	}
}

func logAclMatched(cfg *config.Config, aclFunction string, user *AuthenticatedUser, action *config.Action, acl *config.AccessControlList) {
	if cfg.LogDebugOptions.AclMatched {
		log.WithFields(log.Fields{
			"User":   user.Username,
			"Action": action.Title,
			"ACL":    acl.Name,
		}).Debugf("%v - Matched ACL", aclFunction)
	}
}

// IsAllowedLogs checks if a AuthenticatedUser is allowed to view an action's logs
func IsAllowedLogs(cfg *config.Config, user *AuthenticatedUser, action *config.Action) bool {
	for _, acl := range getRelevantAcls(cfg, action.Acls, user) {
		if acl.Permissions.Logs {
			logAclMatched(cfg, "isAllowedLogs", user, action, acl)

			return true
		}
	}

	logAclNotMatched(cfg, "isAllowedLogs", user, action)

	return cfg.DefaultPermissions.Logs
}

// IsAllowedExec checks if a AuthenticatedUser is allowed to execute an Action
func IsAllowedExec(cfg *config.Config, user *AuthenticatedUser, action *config.Action) bool {
	for _, acl := range getRelevantAcls(cfg, action.Acls, user) {
		if acl.Permissions.Exec {
			logAclMatched(cfg, "isAllowedExec", user, action, acl)

			return true
		}
	}

	logAclNotMatched(cfg, "isAllowedExec", user, action)

	return cfg.DefaultPermissions.Exec
}

// IsAllowedView checks if a User is allowed to view an Action
func IsAllowedView(cfg *config.Config, user *AuthenticatedUser, action *config.Action) bool {
	if action.Hidden {
		return false
	}

	for _, acl := range getRelevantAcls(cfg, action.Acls, user) {
		if acl.Permissions.View {
			logAclMatched(cfg, "isAllowedView", user, action, acl)

			return true
		}
	}

	logAclNotMatched(cfg, "isAllowedView", user, action)

	return cfg.DefaultPermissions.View
}

func getMetdataKeyOrEmpty(md metadata.MD, key string) string {
	mdValues := md.Get(key)

	if len(mdValues) > 0 {
		return mdValues[0]
	}

	return ""
}

// UserFromContext tries to find a user from a grpc context
func UserFromContext(ctx context.Context, cfg *config.Config) *AuthenticatedUser {
	md, ok := metadata.FromIncomingContext(ctx)

	ret := &AuthenticatedUser{
		Username:  "guest",
		Usergroup: "guest",
	}

	if ok {
		ret.Username = getMetdataKeyOrEmpty(md, "username")
		ret.Usergroup = getMetdataKeyOrEmpty(md, "usergroup")
	}

	buildUserAcls(cfg, ret)

	log.WithFields(log.Fields{
		"username":  ret.Username,
		"usergroup": ret.Usergroup,
	}).Debugf("UserFromContext")

	return ret
}

func UserFromSystem(cfg *config.Config, username string) *AuthenticatedUser {
	ret := &AuthenticatedUser{
		Username:  username,
		Usergroup: "system",
	}

	buildUserAcls(cfg, ret)

	return ret
}

func buildUserAcls(cfg *config.Config, user *AuthenticatedUser) {
	for _, acl := range cfg.AccessControlLists {
		if slices.Contains(acl.MatchUsernames, user.Username) {
			user.acls = append(user.acls, acl.Name)
			continue
		}

		if slices.Contains(acl.MatchUsergroups, user.Usergroup) {
			user.acls = append(user.acls, acl.Name)
			continue

		}
	}
}

func isACLRelevantToAction(cfg *config.Config, actionAcls []string, acl *config.AccessControlList, user *AuthenticatedUser) bool {
	if !slices.Contains(user.acls, acl.Name) {
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
