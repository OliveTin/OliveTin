package api

import (
	"sort"

	apiv1 "github.com/OliveTin/OliveTin/gen/olivetin/api/v1"
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
