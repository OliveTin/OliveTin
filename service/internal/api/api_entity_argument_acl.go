package api

import (
	"fmt"
	"strings"

	"connectrpc.com/connect"

	authpublic "github.com/OliveTin/OliveTin/internal/auth/authpublic"
	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/entities"
	"github.com/OliveTin/OliveTin/internal/executor"
	"github.com/OliveTin/OliveTin/internal/tpl"
)

// errUnlessStartEntityAccessAllowed enforces entity-type view ACL on the binding
// and rejects entity-backed argument values the user may not use.
func (api *oliveTinAPI) errUnlessStartEntityAccessAllowed(user *authpublic.AuthenticatedUser, binding *executor.ActionBinding, args map[string]string) error {
	if err := api.errUnlessBindingEntityTypeAllowed(user, binding); err != nil {
		return err
	}

	if binding == nil {
		return nil
	}

	return api.errUnlessEntityArgumentsAllowed(user, binding.Action, args)
}

// errUnlessEntityArgumentsAllowed rejects starts that use entity-backed arguments
// the user may not view, or guessed values that are not in the allowed choice set.
func (api *oliveTinAPI) errUnlessEntityArgumentsAllowed(user *authpublic.AuthenticatedUser, action *config.Action, args map[string]string) error {
	if action == nil {
		return nil
	}

	for i := range action.Arguments {
		arg := &action.Arguments[i]
		if arg.Entity == "" {
			continue
		}

		if err := api.errUnlessEntityArgumentAllowed(user, arg, args[arg.Name]); err != nil {
			return err
		}
	}

	return nil
}

func isEntityBackedArgument(arg *config.ActionArgument) bool {
	return arg != nil && arg.Entity != "" && len(arg.Choices) == 1
}

func isMalformedEntityArgument(arg *config.ActionArgument) bool {
	return arg != nil && arg.Entity != "" && len(arg.Choices) != 1
}

func (api *oliveTinAPI) errUnlessEntityArgumentAllowed(user *authpublic.AuthenticatedUser, arg *config.ActionArgument, value string) error {
	if err := errUnlessEntityArgumentShapeAllowed(arg); err != nil {
		return err
	}

	if !isEntityBackedArgument(arg) {
		return nil
	}

	if !api.userCanViewEntityType(user, arg.Entity) {
		return connect.NewError(connect.CodePermissionDenied, fmt.Errorf("permission denied"))
	}

	if err := errUnlessEntityArgumentValueAllowed(arg, value); err != nil {
		return connect.NewError(connect.CodeInvalidArgument, err)
	}

	return nil
}

func errUnlessEntityArgumentShapeAllowed(arg *config.ActionArgument) error {
	if !isMalformedEntityArgument(arg) {
		return nil
	}

	return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("argument %q with entity must define exactly one choice template", arg.Name))
}

func errUnlessEntityArgumentValueAllowed(arg *config.ActionArgument, value string) error {
	if isMalformedEntityArgument(arg) {
		return fmt.Errorf("argument %q with entity must define exactly one choice template", arg.Name)
	}

	if !isEntityBackedArgument(arg) {
		return nil
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	if !entityArgumentValueAllowed(arg, value) {
		return fmt.Errorf("argument %q is not a permitted entity value", arg.Name)
	}

	return nil
}

func entityArgumentValueAllowed(arg *config.ActionArgument, value string) bool {
	allowed := entityArgumentAllowedValues(arg)
	if strings.EqualFold(arg.Type, "checklist") {
		return checklistEntityValuesAllowed(value, allowed)
	}

	_, ok := allowed[value]
	return ok
}

func entityArgumentAllowedValues(arg *config.ActionArgument) map[string]struct{} {
	allowed := make(map[string]struct{})
	if arg == nil || len(arg.Choices) != 1 {
		return allowed
	}

	for _, ent := range entities.GetEntityInstancesOrdered(arg.Entity) {
		resolved := tpl.ParseTemplateOfActionBeforeExec(arg.Choices[0].Value, ent)
		if resolved == "" {
			continue
		}
		allowed[resolved] = struct{}{}
	}

	return allowed
}

func checklistEntityValuesAllowed(value string, allowed map[string]struct{}) bool {
	parts := strings.Split(value, ",")
	sawItem := false

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		sawItem = true
		if _, ok := allowed[part]; !ok {
			return false
		}
	}

	return sawItem
}
