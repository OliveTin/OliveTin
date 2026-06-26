package executor

import (
	"strings"

	config "github.com/OliveTin/OliveTin/internal/config"
)

func argumentTypeStorableInLog(argType string) bool {
	switch argType {
	case "password", "very_dangerous_raw_string":
		return false
	default:
		return true
	}
}

func storableArgumentNames(action *config.Action) map[string]struct{} {
	if action == nil {
		return nil
	}

	names := make(map[string]struct{}, len(action.Arguments))
	for i := range action.Arguments {
		arg := &action.Arguments[i]
		if !argumentTypeStorableInLog(arg.Type) {
			continue
		}

		names[arg.Name] = struct{}{}
	}

	return names
}

func storableArgumentNamesFromRequest(req *ExecutionRequest) map[string]struct{} {
	if req == nil || req.Binding == nil || req.Binding.Action == nil {
		return nil
	}

	return storableArgumentNames(req.Binding.Action)
}

func isStorableArgumentName(name string, allowedNames map[string]struct{}) bool {
	if strings.HasPrefix(name, config.ReservedArgumentNamePrefix) {
		return false
	}

	_, ok := allowedNames[name]
	return ok
}

func collectStorableArguments(args map[string]string, allowedNames map[string]struct{}) map[string]string {
	result := make(map[string]string)
	for name, value := range args {
		if isStorableArgumentName(name, allowedNames) {
			result[name] = value
		}
	}

	return result
}

func filterStorableArguments(args map[string]string, allowedNames map[string]struct{}) map[string]string {
	if len(args) == 0 {
		return nil
	}

	result := collectStorableArguments(args, allowedNames)
	if len(result) == 0 {
		return nil
	}

	return result
}

func storableArgumentsFromRequest(req *ExecutionRequest) map[string]string {
	allowedNames := storableArgumentNamesFromRequest(req)
	if len(allowedNames) == 0 {
		return nil
	}

	return filterStorableArguments(req.Arguments, allowedNames)
}

func copyStorableArgumentsToLogEntry(req *ExecutionRequest) {
	args := storableArgumentsFromRequest(req)
	if args == nil || req.logEntry == nil {
		return
	}

	req.mutateLogEntry(func(entry *InternalLogEntry) {
		entry.Arguments = args
	})
}
