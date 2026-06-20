package api

import (
	"fmt"
	"strings"

	"connectrpc.com/connect"

	apiv1 "github.com/OliveTin/OliveTin/gen/olivetin/api/v1"
	authpublic "github.com/OliveTin/OliveTin/internal/auth/authpublic"
	"github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/executor"
)

func validateJustificationRequired(action *config.Action, justification string, user *authpublic.AuthenticatedUser) error {
	if !actionRequiresJustificationConfig(action) || justificationProvided(justification, user) {
		return nil
	}

	return fmt.Errorf("justification is required for this action")
}

func actionRequiresJustificationConfig(action *config.Action) bool {
	return action != nil && action.Justification
}

func justificationProvided(justification string, user *authpublic.AuthenticatedUser) bool {
	return strings.TrimSpace(justification) != "" || executor.IsSystemExecution(user)
}

func connectInvalidJustification(err error) error {
	return connect.NewError(connect.CodeInvalidArgument, err)
}

func startActionArgumentsFromProto(args []*apiv1.StartActionArgument) map[string]string {
	result := make(map[string]string, len(args))
	for _, arg := range args {
		result[arg.Name] = arg.Value
	}
	return result
}

func restartRequiresJustificationError() error {
	return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("justification is required for this action; use StartAction with a justification instead"))
}
