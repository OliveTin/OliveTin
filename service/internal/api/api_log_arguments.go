package api

import (
	"fmt"
	"sort"
	"strings"

	"connectrpc.com/connect"

	apiv1 "github.com/OliveTin/OliveTin/gen/olivetin/api/v1"
	"github.com/OliveTin/OliveTin/internal/executor"
)

func logEntryArgumentsToProto(args map[string]string) []*apiv1.StartActionArgument {
	if len(args) == 0 {
		return nil
	}

	names := make([]string, 0, len(args))
	for name := range args {
		names = append(names, name)
	}
	sort.Strings(names)

	out := make([]*apiv1.StartActionArgument, 0, len(names))
	for _, name := range names {
		out = append(out, &apiv1.StartActionArgument{
			Name:  name,
			Value: args[name],
		})
	}

	return out
}

func copyStringMap(source map[string]string) map[string]string {
	copied := make(map[string]string, len(source))
	for key, value := range source {
		copied[key] = value
	}

	return copied
}

func restartArgumentsIncompleteError() error {
	return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("stored arguments are incomplete for restart; use StartAction with the required arguments instead"))
}

func validateRestartLogEntry(entry *executor.InternalLogEntry) error {
	if entry.Binding.Action.Justification && strings.TrimSpace(entry.Justification) == "" {
		return restartRequiresJustificationError()
	}

	if executor.RestartArgumentsIncomplete(entry.Binding.Action, entry.Binding.Entity, entry.Arguments) {
		return restartArgumentsIncompleteError()
	}

	return nil
}
