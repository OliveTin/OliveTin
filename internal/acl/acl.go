package acl

import (
	"context"
	config "github.com/OliveTin/OliveTin/internal/config"
	log "github.com/sirupsen/logrus"
)

// User respresents a person.
type User struct {
	Username string
}

// IsAllowedExec checks if a User is allowed to execute an Action
func IsAllowedExec(cfg *config.Config, user *User, action *config.Action) bool {
	canExec := cfg.DefaultPermissions.Exec

	log.WithFields(log.Fields{
		"User":    user.Username,
		"Action":  action.Title,
		"CanExec": canExec,
	}).Debug("isAllowedExec Permission Default")

	for _, permissionEntry := range action.Permissions {
		if isUserInGroup(user, permissionEntry.Usergroup) {
			log.WithFields(log.Fields{
				"User":    user.Username,
				"Action":  action.Title,
				"CanExec": permissionEntry.Exec,
			}).Debug("isAllowedExec Permission Entry")

			canExec = permissionEntry.Exec
		}
	}

	log.WithFields(log.Fields{
		"User":    user.Username,
		"Action":  action.Title,
		"CanExec": canExec,
	}).Debug("isAllowedExec Final Result")

	return canExec
}

// IsAllowedView checks if a User is allowed to view an Action
func IsAllowedView(cfg *config.Config, user *User, action *config.Action) bool {
	canView := cfg.DefaultPermissions.View

	log.WithFields(log.Fields{
		"User":    user.Username,
		"Action":  action.Title,
		"CanView": canView,
	}).Debug("isAllowedView Permission Default")

	for idx, permissionEntry := range action.Permissions {
		if isUserInGroup(user, permissionEntry.Usergroup) {
			log.WithFields(log.Fields{
				"User":    user.Username,
				"Action":  action.Title,
				"CanView": permissionEntry.View,
				"Index":   idx,
			}).Debug("isAllowedView Permission Entry")

			canView = permissionEntry.View
		}
	}

	log.WithFields(log.Fields{
		"User":    user.Username,
		"Action":  action.Title,
		"CanView": canView,
	}).Debug("isAllowedView Final Result")

	return canView
}

func isUserInGroup(user *User, usergroup string) bool {
	return true
}

// UserFromContext tries to find a user from a grpc context - obviously this is
// a stub at the moment.
func UserFromContext(ctx context.Context) *User {
	return &User{
		Username: "Guest",
	}
}
