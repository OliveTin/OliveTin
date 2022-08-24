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

// IsAllowedExec checks if a AuthenticatedUser is allowed to execute an Action
func IsAllowedExec(cfg *config.Config, user *AuthenticatedUser, action *config.Action) bool {
	for _, acl := range getRelevantAcls(cfg, action.Acls, user) {
		if acl.Permissions.Exec {
			log.WithFields(log.Fields{
				"User":   user.Username,
				"Action": action.Title,
				"ACL":    acl.Name,
			}).Debug("isAllowedExec - Matched ACL")

			return true
		}
	}

	log.WithFields(log.Fields{
		"User":   user.Username,
		"Action": action.Title,
	}).Debug("isAllowedExec - No ACLs matched")

	return cfg.DefaultPermissions.Exec
}

// IsAllowedView checks if a User is allowed to view an Action
func IsAllowedView(cfg *config.Config, user *AuthenticatedUser, action *config.Action) bool {
	for _, acl := range getRelevantAcls(cfg, action.Acls, user) {
		if acl.Permissions.View {
			log.WithFields(log.Fields{
				"User":   user.Username,
				"Action": action.Title,
				"ACL":    acl.Name,
			}).Debug("isAllowedView - Matched ACL")

			return true
		}
	}

	log.WithFields(log.Fields{
		"User":   user.Username,
		"Action": action.Title,
	}).Debug("isAllowedView - No ACLs matched")

	return cfg.DefaultPermissions.View
}

// UserFromContext tries to find a user from a grpc context
func UserFromContext(ctx context.Context, cfg *config.Config) *AuthenticatedUser {
	md, _ := metadata.FromIncomingContext(ctx)

	ret := &AuthenticatedUser{
		Username:  md.Get("username")[0],
		Usergroup: md.Get("usergroup")[0],
	}

	buildUserAcls(cfg, ret)

	log.Infof("UserFromContext: %+v", ret)

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

func getRelevantAcls(cfg *config.Config, actionAcls []string, user *AuthenticatedUser) []*config.AccessControlList {
	var ret []*config.AccessControlList

	for _, acl := range cfg.AccessControlLists {
		if !slices.Contains(user.acls, acl.Name) {
			continue
		}

		if acl.AddToEveryAction {
			ret = append(ret, &acl)
			continue
		}

		if slices.Contains(actionAcls, acl.Name) {
			ret = append(ret, &acl)
			continue
		}
	}

	return ret
}
