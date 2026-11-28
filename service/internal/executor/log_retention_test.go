package executor

import (
	"testing"
	"time"

	auth "github.com/OliveTin/OliveTin/internal/auth"
	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newRetentionTestExecutor(t *testing.T, action *config.Action) (*Executor, *config.Config, *ActionBinding) {
	t.Helper()

	cfg := config.DefaultConfig()
	cfg.LogHistoryLimit = config.MinimumLogHistoryLimit
	cfg.Actions = append(cfg.Actions, action)
	cfg.Sanitize()

	e := DefaultExecutor(cfg)
	e.RebuildActionMap()

	binding := e.FindBindingWithNoEntity(action)
	require.NotNil(t, binding)

	return e, cfg, binding
}

func addRetentionTestLog(e *Executor, entry *InternalLogEntry) {
	e.logmutex.Lock()
	e.addLogToBindingMap(entry)
	e.logmutex.Unlock()

	e.SetLog(entry.ExecutionTrackingID, entry)
}

func assertLogStoresConsistent(t *testing.T, e *Executor, bindingID string) {
	t.Helper()

	e.logmutex.RLock()
	defer e.logmutex.RUnlock()

	assert.Len(t, e.logs, len(e.logsTrackingIdsByDate), "map and date index must hold the same entries")
	for _, entry := range e.LogsByBindingId[bindingID] {
		_, found := e.logs[entry.ExecutionTrackingID]
		assert.True(t, found, "per-binding logs must not keep evicted entries")
	}
}

func TestLogHistoryEvictsOldestFinishedEntries(t *testing.T) {
	e, cfg, binding := newRetentionTestExecutor(t, &config.Action{Title: "retention", Shell: "true"})

	entries := make([]*InternalLogEntry, 0, 250)
	for range 250 {
		entry := newQueueTestLogEntry(binding, true)
		entries = append(entries, entry)
		addRetentionTestLog(e, entry)
	}

	e.logmutex.RLock()
	assert.Len(t, e.logsTrackingIdsByDate, cfg.LogHistoryLimit)
	assert.Equal(t, entries[150].ExecutionTrackingID, e.logsTrackingIdsByDate[0], "oldest entries are evicted first")
	assert.Len(t, e.LogsByBindingId[binding.ID], cfg.LogHistoryLimit)
	e.logmutex.RUnlock()

	_, found := e.GetLog(entries[0].ExecutionTrackingID)
	assert.False(t, found)
	_, found = e.GetLog(entries[249].ExecutionTrackingID)
	assert.True(t, found)
	assert.Equal(t, int64(249), entries[249].Index, "indexes keep increasing after eviction")

	assertLogStoresConsistent(t, e, binding.ID)
}

func TestLogHistoryKeepsRunningEntries(t *testing.T) {
	e, cfg, binding := newRetentionTestExecutor(t, &config.Action{Title: "retention", Shell: "true"})

	running := newQueueTestLogEntry(binding, false)
	addRetentionTestLog(e, running)

	for range 250 {
		addRetentionTestLog(e, newQueueTestLogEntry(binding, true))
	}

	_, found := e.GetLog(running.ExecutionTrackingID)
	assert.True(t, found, "running executions must never be evicted")

	e.logmutex.RLock()
	assert.Len(t, e.logsTrackingIdsByDate, cfg.LogHistoryLimit)
	e.logmutex.RUnlock()
	assertLogStoresConsistent(t, e, binding.ID)
}

func TestLogHistoryKeepsEntriesCountingTowardsRateLimit(t *testing.T) {
	limited := &config.Action{
		Title:   "rate limited",
		Shell:   "true",
		MaxRate: []config.RateSpec{{Limit: 3, Duration: "1h"}},
	}
	e, _, limitedBinding := newRetentionTestExecutor(t, limited)

	other := &config.Action{Title: "other", Shell: "true"}
	e.Cfg.Actions = append(e.Cfg.Actions, other)
	e.RebuildActionMap()
	otherBinding := e.FindBindingWithNoEntity(other)
	require.NotNil(t, otherBinding)

	counted := make([]*InternalLogEntry, 0, 3)
	for range 3 {
		entry := newQueueTestLogEntry(limitedBinding, true)
		counted = append(counted, entry)
		addRetentionTestLog(e, entry)
	}

	blocked := newQueueTestLogEntry(limitedBinding, true)
	blocked.Blocked = true
	addRetentionTestLog(e, blocked)

	for range 250 {
		addRetentionTestLog(e, newQueueTestLogEntry(otherBinding, true))
	}

	for _, entry := range counted {
		_, found := e.GetLog(entry.ExecutionTrackingID)
		assert.True(t, found, "entries inside the maxRate window must not be evicted")
	}

	_, found := e.GetLog(blocked.ExecutionTrackingID)
	assert.False(t, found, "blocked entries do not count towards the rate limit and can be evicted")

	req := &ExecutionRequest{Binding: limitedBinding, executor: e, logEntry: &InternalLogEntry{}}
	assert.Equal(t, 3, getExecutionsCount(limited.MaxRate[0], req), "flooding other actions must not reset the rate limit")
	assertLogStoresConsistent(t, e, limitedBinding.ID)
}

func TestLogHistoryBoundsDeniedRequests(t *testing.T) {
	action := &config.Action{Title: "admins only", Shell: "true"}
	e, cfg, binding := newRetentionTestExecutor(t, action)
	cfg.DefaultPermissions.Exec = false
	user := auth.UserGuest(cfg)

	for range 250 {
		wg, _ := e.ExecRequest(&ExecutionRequest{AuthenticatedUser: user, Cfg: cfg, Binding: binding})
		wg.Wait()
	}

	e.logmutex.RLock()
	defer e.logmutex.RUnlock()
	assert.LessOrEqual(t, len(e.logsTrackingIdsByDate), cfg.LogHistoryLimit, "denied requests must not grow the log store without limit")
	assert.LessOrEqual(t, len(e.LogsByBindingId[binding.ID]), cfg.LogHistoryLimit)
}

func TestLogHistoryLimitIsSanitized(t *testing.T) {
	cfg := config.DefaultConfig()
	assert.Equal(t, config.DefaultLogHistoryLimit, cfg.LogHistoryLimit)

	cfg.LogHistoryLimit = 5
	cfg.Sanitize()
	assert.Equal(t, config.MinimumLogHistoryLimit, cfg.LogHistoryLimit)
}

func TestLogHistoryLoadedFromDiskIsBounded(t *testing.T) {
	e, cfg, binding := newRetentionTestExecutor(t, &config.Action{Title: "retention", Shell: "true"})

	loaded := make([]*InternalLogEntry, 0, 250)
	for i := range 250 {
		entry := newQueueTestLogEntry(binding, true)
		entry.DatetimeStarted = time.Now().Add(time.Duration(i-250) * time.Minute)
		loaded = append(loaded, entry)
	}

	e.restoreLogsToExecutor(loaded, 0)

	e.logmutex.RLock()
	assert.Len(t, e.logsTrackingIdsByDate, cfg.LogHistoryLimit)
	assert.Equal(t, loaded[150].ExecutionTrackingID, e.logsTrackingIdsByDate[0], "the newest loaded logs are kept")
	e.logmutex.RUnlock()
	assertLogStoresConsistent(t, e, binding.ID)
}

func TestRequestLogEntrySurvivesHistoryEviction(t *testing.T) {
	e, cfg, binding := newRetentionTestExecutor(t, &config.Action{Title: "retention", Shell: "true"})

	req := &ExecutionRequest{
		Binding:           binding,
		Cfg:               cfg,
		AuthenticatedUser: auth.UserFromSystem(cfg, "test"),
	}
	wg, trackingID := e.ExecRequest(req)
	wg.Wait()

	for range cfg.LogHistoryLimit + 1 {
		addRetentionTestLog(e, newQueueTestLogEntry(binding, true))
	}

	_, found := e.GetLog(trackingID)
	require.False(t, found, "the request's entry is evicted from the history")

	entry := req.LogEntry()
	require.NotNil(t, entry)
	assert.Equal(t, trackingID, entry.ExecutionTrackingID)
	assert.True(t, entry.ExecutionFinished)
	assert.Equal(t, int32(0), entry.ExitCode)
}
