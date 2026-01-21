package executor

import (
	"context"
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// timeoutContext is a custom context that kills the process group when cancelled due to timeout.
type timeoutContext struct {
	context.Context
	cancel    context.CancelFunc
	process   *os.Process
	executor  *Executor
	processMu sync.Mutex
}

// newTimeoutContext creates a context that will kill the process group when the timeout expires.
func newTimeoutContext(parent context.Context, timeout time.Duration, executor *Executor) (*timeoutContext, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(parent, timeout)
	tc := &timeoutContext{
		Context:  ctx,
		cancel:   cancel,
		executor: executor,
	}

	// Start a goroutine that kills the process group when the context is cancelled
	go func() {
		<-ctx.Done()
		if ctx.Err() == context.DeadlineExceeded {
			tc.processMu.Lock()
			process := tc.process
			tc.processMu.Unlock()

			if process != nil {
				logEntry := &InternalLogEntry{Process: process}
				if err := executor.Kill(logEntry); err != nil {
					log.WithFields(log.Fields{
						"error": err,
					}).Warnf("Failed to kill process group on timeout")
				}
			}
		}
	}()

	return tc, cancel
}

func (tc *timeoutContext) setProcess(process *os.Process) {
	tc.processMu.Lock()
	tc.process = process
	tc.processMu.Unlock()

	// If deadline already expired before process was set, kill now
	if tc.Context.Err() == context.DeadlineExceeded && process != nil {
		logEntry := &InternalLogEntry{Process: process}
		if err := tc.executor.Kill(logEntry); err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Warnf("Failed to kill process group on timeout (late registration)")
		}
	}
}
