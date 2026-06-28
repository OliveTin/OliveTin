package executor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecutionResultLabel(t *testing.T) {
	tests := []struct {
		name  string
		entry *InternalLogEntry
		want  string
	}{
		{
			name: "success",
			entry: &InternalLogEntry{
				ExecutionStarted:  true,
				ExecutionFinished: true,
				ExitCode:          0,
			},
			want: executionResultSuccess,
		},
		{
			name: "failed nonzero exit",
			entry: &InternalLogEntry{
				ExecutionStarted:  true,
				ExecutionFinished: true,
				ExitCode:          1,
			},
			want: executionResultFailed,
		},
		{
			name: "blocked",
			entry: &InternalLogEntry{
				Blocked:           true,
				ExecutionFinished: true,
				ExitCode:          0,
			},
			want: executionResultBlocked,
		},
		{
			name: "timeout",
			entry: &InternalLogEntry{
				ExecutionStarted:  true,
				ExecutionFinished: true,
				TimedOut:          true,
				ExitCode:          -1,
			},
			want: executionResultTimeout,
		},
		{
			name: "error before execution",
			entry: &InternalLogEntry{
				ExecutionFinished: true,
				ExitCode:          DefaultExitCodeNotExecuted,
			},
			want: executionResultError,
		},
		{
			name: "error never started",
			entry: &InternalLogEntry{
				ExecutionStarted:  false,
				ExecutionFinished: true,
				ExitCode:          2,
			},
			want: executionResultError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, executionResultLabel(tt.entry))
		})
	}
}
