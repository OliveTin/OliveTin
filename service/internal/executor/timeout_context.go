package executor

import (
	"context"
	"os"
	"sync"
	"time"
)

// timeoutContext is a custom context that kills the process group when cancelled due to timeout.
type timeoutContext struct {
	context.Context
	cancel    context.CancelFunc
	logEntry  *InternalLogEntry
	process   *os.Process
	executor  *Executor
	processMu sync.Mutex
}

// newTimeoutContext creates a context that will kill the process group when the timeout expires.
// logEntry is the same InternalLogEntry as the running execution so Kill(logEntry) can use Attributes and Process.
// Pass nil logEntry for subprocesses (e.g. shellAfterCompleted); then setProcess stores the process and it is killed on timeout.
func newTimeoutContext(parent context.Context, timeout time.Duration, executor *Executor, logEntry *InternalLogEntry) (*timeoutContext, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(parent, timeout)
	tc := &timeoutContext{
		Context:  ctx,
		cancel:   cancel,
		logEntry: logEntry,
		executor: executor,
	}

	go func() {
		<-ctx.Done()
		if ctx.Err() != context.DeadlineExceeded {
			return
		}
		tc.processMu.Lock()
		entry := tc.logEntry
		process := tc.process
		tc.processMu.Unlock()
		if entry != nil {
			_ = executor.Kill(entry)
		} else if process != nil {
			_ = executor.Kill(&InternalLogEntry{Process: process})
		}
	}()

	return tc, cancel
}

func (tc *timeoutContext) setProcess(process *os.Process) {
	tc.processMu.Lock()
	if tc.logEntry != nil {
		tc.logEntry.Process = process
	} else {
		tc.process = process
	}
	tc.processMu.Unlock()

	if tc.Context.Err() == context.DeadlineExceeded && process != nil {
		if tc.logEntry != nil {
			_ = tc.executor.Kill(tc.logEntry)
		} else {
			_ = tc.executor.Kill(&InternalLogEntry{Process: process})
		}
	}
}
