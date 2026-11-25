package executor

import (
	"testing"
	"time"

	auth "github.com/OliveTin/OliveTin/internal/auth"
	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetActiveExecutionsACLFiltersFinishedAndACL(t *testing.T) {
	e, cfg := testingExecutor()

	allowedAction := &config.Action{
		Title: "allowed",
		Shell: "sleep 1",
		Acls:  []string{"view-logs"},
	}
	secretAction := &config.Action{
		Title: "secret",
		Shell: "sleep 1",
	}
	cfg.Actions = append(cfg.Actions, allowedAction, secretAction)
	cfg.DefaultPermissions.Logs = false
	cfg.AccessControlLists = []*config.AccessControlList{
		{
			Name:           "view-logs",
			MatchUsernames: []string{"guest"},
			Permissions: config.PermissionsList{
				Logs: true,
			},
		},
	}
	cfg.Sanitize()
	e.RebuildActionMap()

	allowedBinding := e.FindBindingWithNoEntity(allowedAction)
	secretBinding := e.FindBindingWithNoEntity(secretAction)
	require.NotNil(t, allowedBinding)
	require.NotNil(t, secretBinding)

	activeAllowed := newQueueTestLogEntry(allowedBinding, false)
	finishedAllowed := newQueueTestLogEntry(allowedBinding, true)
	activeSecret := newQueueTestLogEntry(secretBinding, false)

	e.SetLog(activeAllowed.ExecutionTrackingID, activeAllowed)
	e.SetLog(finishedAllowed.ExecutionTrackingID, finishedAllowed)
	e.SetLog(activeSecret.ExecutionTrackingID, activeSecret)

	user := auth.UserGuest(cfg)
	active := e.GetActiveExecutionsACL(cfg, user)

	require.Len(t, active, 1)
	assert.Equal(t, activeAllowed.ExecutionTrackingID, active[0].ExecutionTrackingID)
}

func newQueueTestLogEntry(binding *ActionBinding, finished bool) *InternalLogEntry {
	entry := &InternalLogEntry{
		Binding:             binding,
		DatetimeStarted:     time.Now(),
		ExecutionTrackingID: uuid.NewString(),
		ActionTitle:         binding.Action.Title,
		ExecutionFinished:   finished,
	}
	if finished {
		entry.DatetimeFinished = time.Now()
	}
	return entry
}

func TestGetActiveExecutionsACLReturnsCopies(t *testing.T) {
	e, cfg := testingExecutor()
	action := &config.Action{Title: "running", Shell: "sleep 1"}
	cfg.Actions = append(cfg.Actions, action)
	cfg.Sanitize()
	e.RebuildActionMap()

	binding := e.FindBindingWithNoEntity(action)
	require.NotNil(t, binding)

	running := newQueueTestLogEntry(binding, false)
	e.SetLog(running.ExecutionTrackingID, running)

	stop := make(chan struct{})
	writerDone := make(chan struct{})
	go func() {
		defer close(writerDone)
		for {
			select {
			case <-stop:
				return
			default:
				e.logmutex.Lock()
				running.Output += "x"
				running.ExitCode++
				e.logmutex.Unlock()
			}
		}
	}()

	user := auth.UserGuest(cfg)
	for range 200 {
		for _, entry := range e.GetActiveExecutionsACL(cfg, user) {
			assert.NotSame(t, running, entry, "callers must get a copy of a running entry")
			_ = len(entry.Output) + int(entry.ExitCode)
		}
	}

	close(stop)
	<-writerDone
}
