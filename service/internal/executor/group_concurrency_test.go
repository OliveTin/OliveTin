package executor

import (
	"sync"
	"testing"
	"time"

	"github.com/OliveTin/OliveTin/internal/auth"
	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testGroupExecutor(actions []*config.Action, groups map[string]*config.ActionGroup) (*Executor, *config.Config) {
	cfg := config.DefaultConfig()
	cfg.ActionGroups = groups
	cfg.Actions = actions
	cfg.Sanitize()

	e := DefaultExecutor(cfg)
	e.RebuildActionMap()

	return e, cfg
}

func TestGroupConcurrencyQueuesSecondAction(t *testing.T) {
	t.Parallel()

	slowAction := &config.Action{
		Title:  "Unity Job 1",
		Shell:  "sleep 2",
		Groups: []string{"unity"},
	}
	fastAction := &config.Action{
		Title:  "Unity Job 2",
		Shell:  "echo queued-run",
		Groups: []string{"unity"},
	}

	e, cfg := testGroupExecutor(
		[]*config.Action{slowAction, fastAction},
		map[string]*config.ActionGroup{
			"unity": {MaxConcurrent: 1},
		},
	)

	binding1 := e.FindBindingWithNoEntity(slowAction)
	binding2 := e.FindBindingWithNoEntity(fastAction)
	require.NotNil(t, binding1)
	require.NotNil(t, binding2)

	wg1, tracking1 := e.ExecRequest(&ExecutionRequest{
		Binding:           binding1,
		Cfg:               cfg,
		AuthenticatedUser: auth.UserFromSystem(cfg, "testuser"),
	})

	waitUntilExecutionStarted(t, e, tracking1)

	wg2, tracking2 := e.ExecRequest(&ExecutionRequest{
		Binding:           binding2,
		Cfg:               cfg,
		AuthenticatedUser: auth.UserFromSystem(cfg, "testuser"),
	})

	require.Eventually(t, func() bool {
		logEntry, ok := e.GetLog(tracking2)
		return ok && logEntry.Queued
	}, time.Second, 10*time.Millisecond)

	wg1.Wait()
	wg2.Wait()

	logEntry, ok := e.GetLog(tracking2)
	require.True(t, ok)
	assert.False(t, logEntry.Queued)
	assert.False(t, logEntry.Blocked)
	assert.Equal(t, int32(0), logEntry.ExitCode)
	assert.Contains(t, logEntry.Output, "queued-run")
}

func TestDifferentGroupsRunConcurrently(t *testing.T) {
	t.Parallel()

	actionA := &config.Action{
		Title:  "Group A Job",
		Shell:  "sleep 1",
		Groups: []string{"groupA"},
	}
	actionB := &config.Action{
		Title:  "Group B Job",
		Shell:  "echo group-b",
		Groups: []string{"groupB"},
	}

	e, cfg := testGroupExecutor(
		[]*config.Action{actionA, actionB},
		map[string]*config.ActionGroup{
			"groupA": {MaxConcurrent: 1},
			"groupB": {MaxConcurrent: 1},
		},
	)

	wg1, _ := e.ExecRequest(&ExecutionRequest{
		Binding:           e.FindBindingWithNoEntity(actionA),
		Cfg:               cfg,
		AuthenticatedUser: auth.UserFromSystem(cfg, "testuser"),
	})

	time.Sleep(50 * time.Millisecond)

	wg2, tracking2 := e.ExecRequest(&ExecutionRequest{
		Binding:           e.FindBindingWithNoEntity(actionB),
		Cfg:               cfg,
		AuthenticatedUser: auth.UserFromSystem(cfg, "testuser"),
	})

	require.Eventually(t, func() bool {
		logEntry, ok := e.GetLog(tracking2)
		return ok && logEntry.ExecutionFinished && !logEntry.Queued
	}, 2*time.Second, 20*time.Millisecond)

	wg1.Wait()
	wg2.Wait()

	logEntry, ok := e.GetLog(tracking2)
	require.True(t, ok)
	assert.Contains(t, logEntry.Output, "group-b")
}

func TestPerActionConcurrencyStillBlocksWithoutQueue(t *testing.T) {
	t.Parallel()

	action := &config.Action{
		Title:         "Single binding",
		Shell:         "sleep 1",
		MaxConcurrent: 1,
	}

	e, cfg := testGroupExecutor([]*config.Action{action}, nil)
	binding := e.FindBindingWithNoEntity(action)

	wg1, _ := e.ExecRequest(&ExecutionRequest{
		Binding:           binding,
		Cfg:               cfg,
		AuthenticatedUser: auth.UserFromSystem(cfg, "testuser"),
	})

	time.Sleep(50 * time.Millisecond)

	wg2, tracking2 := e.ExecRequest(&ExecutionRequest{
		Binding:           binding,
		Cfg:               cfg,
		AuthenticatedUser: auth.UserFromSystem(cfg, "testuser"),
	})

	wg1.Wait()
	wg2.Wait()

	logEntry, ok := e.GetLog(tracking2)
	require.True(t, ok)
	assert.True(t, logEntry.Blocked)
	assert.False(t, logEntry.Queued)
}

func waitUntilExecutionStarted(t *testing.T, e *Executor, trackingID string) {
	t.Helper()

	require.Eventually(t, func() bool {
		logEntry, ok := e.GetLog(trackingID)
		return ok && logEntry.ExecutionStarted
	}, 2*time.Second, 10*time.Millisecond)
}

func assertWaitGroupPending(t *testing.T, wg *sync.WaitGroup) {
	t.Helper()

	done := make(chan struct{})

	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		t.Fatal("wait group completed before queued execution finished")
	case <-time.After(100 * time.Millisecond):
	}
}

func assertWaitGroupCompletes(t *testing.T, wg *sync.WaitGroup) {
	t.Helper()

	done := make(chan struct{})

	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("wait group did not complete after queue drained")
	}
}

func TestStartActionAndWaitWaitsForQueuedExecution(t *testing.T) {
	t.Parallel()

	first := &config.Action{
		Title:  "Hold group",
		Shell:  "sleep 1",
		Groups: []string{"unity"},
	}
	second := &config.Action{
		Title:  "Wait in queue",
		Shell:  "echo waited",
		Groups: []string{"unity"},
	}

	e, cfg := testGroupExecutor(
		[]*config.Action{first, second},
		map[string]*config.ActionGroup{
			"unity": {MaxConcurrent: 1},
		},
	)

	wg1, tracking1 := e.ExecRequest(&ExecutionRequest{
		Binding:           e.FindBindingWithNoEntity(first),
		Cfg:               cfg,
		AuthenticatedUser: auth.UserFromSystem(cfg, "testuser"),
	})

	waitUntilExecutionStarted(t, e, tracking1)

	wg2, tracking2 := e.ExecRequest(&ExecutionRequest{
		Binding:           e.FindBindingWithNoEntity(second),
		Cfg:               cfg,
		AuthenticatedUser: auth.UserFromSystem(cfg, "testuser"),
	})

	assertWaitGroupPending(t, wg2)

	wg1.Wait()

	assertWaitGroupCompletes(t, wg2)

	logEntry, ok := e.GetLog(tracking2)
	require.True(t, ok)
	assert.Contains(t, logEntry.Output, "waited")
}

func TestUnknownActionGroupReferenceWarnsAndSkipsLimit(t *testing.T) {
	t.Parallel()

	action := &config.Action{
		Title:  "Unknown group action",
		Shell:  "echo ok",
		Groups: []string{"missing"},
	}

	e, cfg := testGroupExecutor([]*config.Action{action}, map[string]*config.ActionGroup{})
	wg, tracking := e.ExecRequest(&ExecutionRequest{
		Binding:           e.FindBindingWithNoEntity(action),
		Cfg:               cfg,
		AuthenticatedUser: auth.UserFromSystem(cfg, "testuser"),
	})

	wg.Wait()

	logEntry, ok := e.GetLog(tracking)
	require.True(t, ok)
	assert.False(t, logEntry.Queued)
	assert.Equal(t, int32(0), logEntry.ExitCode)
}
