package executor

import (
	"fmt"

	"github.com/OliveTin/OliveTin/internal/logfilter"
	"github.com/expr-lang/expr/vm"
)

func filterRecordFromEntry(entry *InternalLogEntry) logfilter.Record {
	return logfilter.Record{
		Status:   logfilter.StatusLabel(entry.ExecutionFinished, entry.Blocked, entry.TimedOut, entry.Queued),
		Action:   entry.ActionTitle,
		User:     entry.Username,
		Tags:     entry.Tags,
		Blocked:  entry.Blocked,
		TimedOut: entry.TimedOut,
		Running:  !entry.ExecutionFinished,
		ExitCode: entry.ExitCode,
		Output:   entry.Output,
	}
}

func applyLogFilter(entries []*InternalLogEntry, program *vm.Program) ([]*InternalLogEntry, error) {
	if program == nil {
		return entries, nil
	}
	return filterEntries(entries, program)
}

func filterEntries(entries []*InternalLogEntry, program *vm.Program) ([]*InternalLogEntry, error) {
	filtered := make([]*InternalLogEntry, 0, len(entries))
	for _, entry := range entries {
		matched, err := entryMatchesFilter(entry, program)
		if err != nil {
			return nil, err
		}
		if matched {
			filtered = append(filtered, entry)
		}
	}
	return filtered, nil
}

func entryMatchesFilter(entry *InternalLogEntry, program *vm.Program) (bool, error) {
	matched, err := logfilter.Matches(program, filterRecordFromEntry(entry))
	if err != nil {
		return false, fmt.Errorf("filter evaluation failed: %w", err)
	}
	return matched, nil
}
