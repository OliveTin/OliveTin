package acl

import (
	config "github.com/jamesread/OliveTin/internal/config"
	log "github.com/sirupsen/logrus"
	"context"
)

type User struct {
	Username string;
}

func IsAllowedExec(cfg *config.Config, user *User, action *config.ActionButton) bool {
	canExec := cfg.DefaultPermissions.Exec

	log.WithFields(log.Fields{
				"User": user.Username,
				"Action": action.Title,
				"CanExec": canExec,
	}).Debug("isAllowedExec Permission Default")

	for _, permissionEntry := range action.Permissions {
		if isUserInGroup(user, permissionEntry.Usergroup) {
			log.WithFields(log.Fields{
				"User": user.Username,
				"Action": action.Title,
				"CanExec": permissionEntry.Exec,
			}).Debug("isAllowedExec Permission Entry")

			canExec = permissionEntry.Exec
		}
	}

	log.WithFields(log.Fields{
		"User": user.Username,
		"Action": action.Title,
		"CanExec": canExec,
	}).Debug("isAllowedExec Final Result")

	return canExec;
}

func IsAllowedView(cfg *config.Config, user *User, action *config.ActionButton) bool {
	canView := cfg.DefaultPermissions.View

	log.WithFields(log.Fields{
				"User": user.Username,
				"Action": action.Title,
				"CanView": canView,
	}).Debug("isAllowedView Permission Default")

	for idx, permissionEntry := range action.Permissions {
		if isUserInGroup(user, permissionEntry.Usergroup) {
			log.WithFields(log.Fields{
				"User": user.Username,
				"Action": action.Title,
				"CanView": permissionEntry.View,
				"Index": idx,
			}).Debug("isAllowedView Permission Entry")

			canView = permissionEntry.View
		}
	}

	log.WithFields(log.Fields{
		"User": user.Username,
		"Action": action.Title,
		"CanView": canView,
	}).Debug("isAllowedView Final Result")

	return canView;
}



func isUserInGroup(user *User, usergroup string) bool {
	return true;
}

func UserFromContext(ctx context.Context) *User {
	return &User {
		Username: "Guest",
	}
}
