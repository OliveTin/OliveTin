package executor

import (
	"testing"

	"github.com/OliveTin/OliveTin/internal/logfilter"
	"github.com/stretchr/testify/require"
)

func TestFilterEntriesRejectsNilEntry(t *testing.T) {
	program, err := logfilter.Compile(`Status == Completed`)
	require.NoError(t, err)

	_, err = filterEntries([]*InternalLogEntry{nil}, program)
	require.Error(t, err)
	require.Contains(t, err.Error(), "log entry is nil")
}
