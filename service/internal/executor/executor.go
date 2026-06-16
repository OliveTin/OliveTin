package executor

import (
	acl "github.com/OliveTin/OliveTin/internal/acl"
	"github.com/OliveTin/OliveTin/internal/auth"
	authpublic "github.com/OliveTin/OliveTin/internal/auth/authpublic"
	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/entities"
	"github.com/OliveTin/OliveTin/internal/tpl"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gopkg.in/yaml.v3"

	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	DefaultExitCodeNotExecuted = -1337
	MaxTriggerDepth            = 10
)

var validTrackingIDPattern = regexp.MustCompile(`^[a-fA-F0-9\-]+$`)

func isValidTrackingID(id string) bool {
	const MaxTrackingIDLength = 36

	return id != "" && len(id) <= MaxTrackingIDLength && validTrackingIDPattern.MatchString(id)
}

var (
	metricActionsRequested = promauto.NewCounter(prometheus.CounterOpts{
		Name: "olivetin_actions_requested_count",
		Help: "The actions requested count",
	})
)

type ActionBinding struct {
	ID            string
	Action        *config.Action
	Entity        *entities.Entity
	ConfigOrder   int
	IsOnDashboard bool
}

// Executor represents a helper class for executing commands. It's main method
// is ExecRequest
type Executor struct {
	logs                  map[string]*InternalLogEntry
	logsTrackingIdsByDate []string
	LogsByBindingId       map[string][]*InternalLogEntry

	logmutex sync.RWMutex

	MapActionBindings     map[string]*ActionBinding
	MapActionBindingsLock sync.RWMutex

	Cfg *config.Config

	listeners []listener

	chainOfCommand []executorStepFunc

	groupQueue   []*queuedExecution
	groupQueueMu sync.Mutex
}

// ExecutionRequest is a request to execute an action. It's passed to an
// Executor. They're created from the api.
type ExecutionRequest struct {
	Binding           *ActionBinding
	Arguments         map[string]string
	TrackingID        string
	Tags              []string
	Cfg               *config.Config
	AuthenticatedUser *authpublic.AuthenticatedUser
	TriggerDepth      int

	logEntry                *InternalLogEntry
	finalParsedCommand      string
	execArgs                []string
	useDirectExec           bool
	executor                *Executor
	skipRequestRegistration bool
}

func (req *ExecutionRequest) mutateLogEntry(mutator func(*InternalLogEntry)) {
	if req.executor == nil {
		mutator(req.logEntry)
		return
	}

	req.executor.logmutex.Lock()
	defer req.executor.logmutex.Unlock()

	mutator(req.logEntry)
}

// LogEntrySnapshot is a copy of selected log entry fields for race-safe reads.
type LogEntrySnapshot struct {
	Queued            bool
	Blocked           bool
	ExecutionStarted  bool
	ExecutionFinished bool
	ExitCode          int32
	Output            string
}

// SnapshotLog returns a copy of selected log entry fields under read lock.
func (e *Executor) SnapshotLog(trackingID string) (LogEntrySnapshot, bool) {
	e.logmutex.RLock()
	defer e.logmutex.RUnlock()

	entry, found := e.logs[trackingID]
	if !found {
		return LogEntrySnapshot{}, false
	}

	return LogEntrySnapshot{
		Queued:            entry.Queued,
		Blocked:           entry.Blocked,
		ExecutionStarted:  entry.ExecutionStarted,
		ExecutionFinished: entry.ExecutionFinished,
		ExitCode:          entry.ExitCode,
		Output:            entry.Output,
	}, true
}

// InternalLogEntry objects are created by an Executor, and represent the final
// state of execution (even if the command is not executed). It's designed to be
// easily serializable.
type InternalLogEntry struct {
	Binding             *ActionBinding
	DatetimeStarted     time.Time
	DatetimeFinished    time.Time
	Output              string
	TimedOut            bool
	Blocked             bool
	Queued              bool
	QueuedForGroup      string
	ExitCode            int32
	Tags                []string
	ExecutionStarted    bool
	ExecutionFinished   bool
	ExecutionTrackingID string
	Process             *os.Process
	Username            string
	Index               int64
	EntityPrefix        string
	ActionConfigTitle   string // This is the title of the action as defined in the config, not the final parsed title.

	/*
		The following 3 properties are obviously on Action normally, but it's useful
		that logs are lightweight (so we don't need to have an action associated to
		logs, etc. Therefore, we duplicate those values here.
	*/
	ActionTitle string
	ActionIcon  string
}

// .Binding can be nil, so we need to handle that.
func (e *InternalLogEntry) GetBindingId() string {
	if e.Binding == nil {
		return ""
	}

	return e.Binding.ID
}

type executorStepFunc func(*ExecutionRequest) bool

// DefaultExecutor returns an Executor, with a sensible "chain of command" for
// executing actions.
func DefaultExecutor(cfg *config.Config) *Executor {
	e := Executor{}
	e.Cfg = cfg
	e.logs = make(map[string]*InternalLogEntry)
	e.logsTrackingIdsByDate = make([]string, 0)
	e.LogsByBindingId = make(map[string][]*InternalLogEntry)
	e.MapActionBindings = make(map[string]*ActionBinding)

	e.chainOfCommand = []executorStepFunc{
		stepRequestAction,
		stepConcurrencyCheck,
		stepRateCheck,
		stepACLCheck,
		stepParseArgs,
		stepLogStart,
		stepExec,
		stepExecAfter,
		stepLogFinish,
		stepSaveLog,
		stepTrigger,
	}

	return &e
}

type listener interface {
	OnExecutionStarted(logEntry *InternalLogEntry)
	OnExecutionFinished(logEntry *InternalLogEntry)
	OnOutputChunk(o []byte, executionTrackingId string)
	OnActionMapRebuilt()
}

func (e *Executor) AddListener(m listener) {
	e.listeners = append(e.listeners, m)
}

// getPagingStartIndex calculates the starting index for log pagination.
// Parameters:
//
//	startOffset: The offset from the most recent log (0 means start from the most recent)
//	totalLogCount: Total number of logs available
//	count: Number of logs to retrieve
//
// Returns: The calculated starting index for pagination
func getPagingStartIndex(startOffset int64, totalLogCount int64) int64 {
	var startIndex int64

	if startOffset <= 0 {
		startIndex = totalLogCount
	} else {
		startIndex = (totalLogCount - startOffset)

		if startIndex < 0 {
			startIndex = 1
		}
	}

	return startIndex - 1
}

type PagingResult struct {
	CountRemaining int64
	PageSize       int64
	TotalCount     int64
	StartOffset    int64
}

func (e *Executor) GetLogTrackingIds(startOffset int64, pageCount int64) ([]*InternalLogEntry, *PagingResult) {
	pagingResult := &PagingResult{
		CountRemaining: 0,
		PageSize:       pageCount,
		TotalCount:     0,
		StartOffset:    startOffset,
	}

	e.logmutex.RLock()

	totalLogCount := int64(len(e.logsTrackingIdsByDate))
	pagingResult.TotalCount = totalLogCount

	startIndex := getPagingStartIndex(startOffset, totalLogCount)

	pageCount = min(totalLogCount, pageCount)

	endIndex := max(0, (startIndex-pageCount)+1)

	log.WithFields(log.Fields{
		"startOffset": startOffset,
		"pageCount":   pageCount,
		"total":       totalLogCount,
		"startIndex":  startIndex,
		"endIndex":    endIndex,
	}).Tracef("GetLogTrackingIds")

	trackingIds := make([]*InternalLogEntry, 0, pageCount)

	if totalLogCount > 0 {
		for i := endIndex; i <= startIndex; i++ {
			trackingIds = append(trackingIds, e.logs[e.logsTrackingIdsByDate[i]])
		}
	}

	e.logmutex.RUnlock()

	pagingResult.CountRemaining = endIndex

	return trackingIds, pagingResult
}

func isValidLogEntryForACL(entry *InternalLogEntry) bool {
	return entry != nil && entry.Binding != nil && entry.Binding.Action != nil
}

func isLogEntryAllowedByACL(cfg *config.Config, user *authpublic.AuthenticatedUser, entry *InternalLogEntry) bool {
	return acl.IsAllowedLogs(cfg, user, entry.Binding.Action)
}

func (e *Executor) filterLogsByACL(cfg *config.Config, user *authpublic.AuthenticatedUser, dateFilter string) []*InternalLogEntry {
	e.logmutex.RLock()
	defer e.logmutex.RUnlock()

	filtered := make([]*InternalLogEntry, 0, len(e.logsTrackingIdsByDate))
	filterDate, hasDateFilter := parseDateFilter(dateFilter)

	for _, trackingId := range e.logsTrackingIdsByDate {
		entry := e.logs[trackingId]

		if shouldIncludeLogEntry(cfg, user, entry, filterDate, hasDateFilter) {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}

// parseDateFilter parses the date filter string and returns filter information.
func parseDateFilter(dateFilter string) (filterDate time.Time, hasDateFilter bool) {
	if dateFilter == "" {
		return time.Time{}, false
	}

	parsedDate, err := time.Parse("2006-01-02", dateFilter)
	if err != nil {
		log.WithFields(log.Fields{
			"dateFilter": dateFilter,
			"error":      err,
		}).Errorf("Failed to parse date filter, expected format YYYY-MM-DD")
		return time.Time{}, false
	}

	return parsedDate, true
}

// shouldIncludeLogEntry determines if a log entry should be included based on ACL and date filter.
func shouldIncludeLogEntry(cfg *config.Config, user *authpublic.AuthenticatedUser, entry *InternalLogEntry, filterDate time.Time, hasDateFilter bool) bool {
	if !isValidLogEntryForACL(entry) {
		return false
	}

	if !isLogEntryAllowedByACL(cfg, user, entry) {
		return false
	}

	return matchesDateFilter(entry, filterDate, hasDateFilter)
}

// matchesDateFilter checks if the log entry matches the date filter.
func matchesDateFilter(entry *InternalLogEntry, filterDate time.Time, hasDateFilter bool) bool {
	if !hasDateFilter {
		return true
	}

	entryDate := entry.DatetimeStarted.UTC().Truncate(24 * time.Hour)
	filterDateUTC := filterDate.UTC().Truncate(24 * time.Hour)
	return entryDate.Equal(filterDateUTC)
}

// paginateFilteredLogs applies pagination to a filtered list of logs and returns
// the paginated results along with pagination metadata.
func paginateFilteredLogs(filtered []*InternalLogEntry, startOffset int64, pageCount int64) ([]*InternalLogEntry, *PagingResult) {
	total := int64(len(filtered))
	paging := &PagingResult{PageSize: pageCount, TotalCount: total, StartOffset: startOffset}

	if total == 0 {
		paging.CountRemaining = 0
		return []*InternalLogEntry{}, paging
	}

	startIndex := getPagingStartIndex(startOffset, total)
	pageCount = min(total, pageCount)
	endIndex := max(0, (startIndex-pageCount)+1)

	out := make([]*InternalLogEntry, 0, pageCount)
	for i := endIndex; i <= startIndex && i < int64(len(filtered)); i++ {
		out = append(out, filtered[i])
	}

	paging.CountRemaining = endIndex
	return out, paging
}

// GetLogTrackingIdsACL returns logs filtered by ACL visibility for the user and
// paginated correctly based on the filtered set.
// dateFilter is optional and should be in YYYY-MM-DD format. If empty, no date filtering is applied.
func (e *Executor) GetLogTrackingIdsACL(cfg *config.Config, user *authpublic.AuthenticatedUser, startOffset int64, pageCount int64, dateFilter string) ([]*InternalLogEntry, *PagingResult) {
	filtered := e.filterLogsByACL(cfg, user, dateFilter)
	return paginateFilteredLogs(filtered, startOffset, pageCount)
}

func (e *Executor) GetLog(trackingID string) (*InternalLogEntry, bool) {
	e.logmutex.RLock()

	entry, found := e.logs[trackingID]

	e.logmutex.RUnlock()

	return entry, found
}

func (e *Executor) GetLogsByBindingId(bindingId string) []*InternalLogEntry {
	e.logmutex.RLock()

	logs, found := e.LogsByBindingId[bindingId]

	e.logmutex.RUnlock()

	if !found {
		return make([]*InternalLogEntry, 0)
	}

	return logs
}

// shouldCountExecution checks if a log entry should be counted for rate limiting.
func shouldCountExecution(logEntry *InternalLogEntry, windowStart time.Time) bool {
	return !logEntry.Blocked && !logEntry.Queued && logEntry.DatetimeStarted.After(windowStart)
}

// updateOldestExecution updates the oldest execution time if this entry is older.
func updateOldestExecution(oldestExecutionTime **time.Time, logEntry *InternalLogEntry) {
	if *oldestExecutionTime == nil {
		*oldestExecutionTime = &logEntry.DatetimeStarted
	} else if logEntry.DatetimeStarted.Before(**oldestExecutionTime) {
		*oldestExecutionTime = &logEntry.DatetimeStarted
	}
}

// findOldestExecutionInWindow finds the oldest execution within the time window and counts executions.
// Returns the count of executions and the oldest execution time, or nil if none found.
func findOldestExecutionInWindow(logs []*InternalLogEntry, windowStart time.Time) (int, *time.Time) {
	executions := 0
	var oldestExecutionTime *time.Time

	for _, logEntry := range logs {
		if !shouldCountExecution(logEntry, windowStart) {
			continue
		}

		executions++
		updateOldestExecution(&oldestExecutionTime, logEntry)
	}

	return executions, oldestExecutionTime
}

// calculateExpiryTime calculates when the oldest execution will fall outside the rate limit window.
func calculateExpiryTime(oldestExecutionTime time.Time, duration time.Duration, now time.Time) time.Time {
	expiryTime := oldestExecutionTime.Add(duration)
	if !expiryTime.After(now) {
		return time.Time{}
	}
	return expiryTime
}

// updateMaxExpiryTime updates maxExpiryTime if expiryTime is later.
func updateMaxExpiryTime(maxExpiryTime *time.Time, expiryTime time.Time) {
	if expiryTime.IsZero() {
		return
	}

	if maxExpiryTime.IsZero() || expiryTime.After(*maxExpiryTime) {
		*maxExpiryTime = expiryTime
	}
}

// calculateExpiryForRate calculates the expiry time for a single rate limit rule.
// Returns the expiry time if the rate limit is exceeded, or zero time if not.
func calculateExpiryForRate(rate config.RateSpec, logs []*InternalLogEntry, now time.Time) time.Time {
	duration := parseDuration(rate)
	if duration <= 0 {
		return time.Time{}
	}

	windowStart := now.Add(-duration)
	executions, oldestExecutionTime := findOldestExecutionInWindow(logs, windowStart)

	if executions < rate.Limit || oldestExecutionTime == nil {
		return time.Time{}
	}

	return calculateExpiryTime(*oldestExecutionTime, duration, now)
}

// getLogsForBinding retrieves logs for a binding ID.
func (e *Executor) getLogsForBinding(bindingId string) []*InternalLogEntry {
	e.logmutex.RLock()
	logs, found := e.LogsByBindingId[bindingId]
	e.logmutex.RUnlock()

	if !found || len(logs) == 0 {
		return nil
	}

	return logs
}

// calculateMaxExpiryTimeFromRates calculates the maximum expiry time across all rate limit rules.
func calculateMaxExpiryTimeFromRates(rates []config.RateSpec, logs []*InternalLogEntry, now time.Time) time.Time {
	var maxExpiryTime time.Time

	for _, rate := range rates {
		expiryTime := calculateExpiryForRate(rate, logs, now)
		updateMaxExpiryTime(&maxExpiryTime, expiryTime)
	}

	return maxExpiryTime
}

// GetTimeUntilAvailable calculates when an action will be available again based on rate limits.
// Returns the Unix timestamp in seconds when the rate limit expires, or 0 if the action is available now.
func (e *Executor) GetTimeUntilAvailable(binding *ActionBinding) int64 {
	if len(binding.Action.MaxRate) == 0 {
		return 0
	}

	logs := e.getLogsForBinding(binding.ID)
	if logs == nil {
		return 0
	}

	maxExpiryTime := calculateMaxExpiryTimeFromRates(binding.Action.MaxRate, logs, time.Now())

	if maxExpiryTime.IsZero() {
		return 0
	}

	return maxExpiryTime.Unix()
}

func (e *Executor) SetLog(trackingID string, entry *InternalLogEntry) string {
	e.logmutex.Lock()
	defer e.logmutex.Unlock()

	if _, found := e.logs[trackingID]; found || !isValidTrackingID(trackingID) {
		trackingID = uuid.NewString()
		entry.ExecutionTrackingID = trackingID
	}

	entry.Index = int64(len(e.logsTrackingIdsByDate))

	e.logs[trackingID] = entry
	e.logsTrackingIdsByDate = append(e.logsTrackingIdsByDate, trackingID)

	return trackingID
}

// ExecRequest processes an ExecutionRequest
func (e *Executor) ExecRequest(req *ExecutionRequest) (*sync.WaitGroup, string) {
	e.initializeExecRequest(req)

	log.Tracef("executor.ExecRequest(): trackingID=%s bindingID=%s", req.TrackingID, bindingIDForTrace(req))

	req.TrackingID = e.SetLog(req.TrackingID, req.logEntry)

	wg := new(sync.WaitGroup)
	wg.Add(1)

	go func() {
		queued := e.execChain(req, wg)
		if !queued {
			wg.Done()
		}
	}()

	return wg, req.TrackingID
}

func (e *Executor) initializeExecRequest(req *ExecutionRequest) {
	if req.AuthenticatedUser == nil {
		req.AuthenticatedUser = auth.UserGuest(req.Cfg)
	}

	req.executor = e
	req.logEntry = &InternalLogEntry{
		Binding:             req.Binding,
		DatetimeStarted:     time.Now(),
		ExecutionTrackingID: req.TrackingID,
		Output:              "",
		ExitCode:            DefaultExitCodeNotExecuted,
		ExecutionStarted:    false,
		ExecutionFinished:   false,
		ActionTitle:         "notfound",
		ActionIcon:          "&#x1f4a9;",
		Username:            req.AuthenticatedUser.Username,
	}
}

func bindingIDForTrace(req *ExecutionRequest) string {
	if req.Binding == nil {
		return ""
	}

	return req.Binding.ID
}

func (e *Executor) execChain(req *ExecutionRequest, wg *sync.WaitGroup) bool {
	if !req.skipRequestRegistration {
		finished, queued := e.registerOrQueueRequest(req, wg)
		if finished || queued {
			return queued
		}
	}

	e.runExecutionSteps(req)
	e.finishExecChain(req)

	return false
}

func (e *Executor) registerOrQueueRequest(req *ExecutionRequest, wg *sync.WaitGroup) (finished bool, queued bool) {
	if !stepRequestAction(req) {
		e.finishExecChain(req)
		return true, false
	}

	if !actionNeedsGroupLimit(req) || e.groupsHaveCapacityForActive(req) {
		return false, false
	}

	return e.queueRequestAfterACL(req, wg)
}

func (e *Executor) queueRequestAfterACL(req *ExecutionRequest, wg *sync.WaitGroup) (finished bool, queued bool) {
	if !stepACLCheck(req) {
		e.finishExecChain(req)
		return true, false
	}

	e.queueRequest(req, wg)
	notifyListenersStarted(req)

	return false, true
}

func (e *Executor) runExecutionSteps(req *ExecutionRequest) {
	for _, step := range e.chainOfCommand[1:] {
		if !step(req) {
			break
		}
	}
}

func (e *Executor) finishExecChain(req *ExecutionRequest) {
	req.mutateLogEntry(func(entry *InternalLogEntry) {
		if entry.DatetimeFinished.IsZero() {
			entry.DatetimeFinished = time.Now()
		}

		entry.ExecutionFinished = true
	})

	notifyListenersFinished(req)
	e.drainGroupQueue()
}

func getConcurrentCount(req *ExecutionRequest) int {
	concurrentCount := 0

	req.executor.logmutex.RLock()
	logs := req.executor.LogsByBindingId[req.Binding.ID]

	for _, logEntry := range logs {
		if !logEntry.ExecutionFinished && !logEntry.Queued {
			concurrentCount += 1
		}
	}

	req.executor.logmutex.RUnlock()

	return concurrentCount
}

func stepConcurrencyCheck(req *ExecutionRequest) bool {
	concurrentCount := getConcurrentCount(req)

	// Note that the current execution is counted int the logs, so when checking we +1
	if concurrentCount >= (req.Binding.Action.MaxConcurrent + 1) {
		log.WithFields(log.Fields{
			"actionTitle":     req.logEntry.ActionTitle,
			"concurrentCount": concurrentCount,
			"maxConcurrent":   req.Binding.Action.MaxConcurrent,
		}).Warnf("Blocked from executing due to concurrency limit")

		req.mutateLogEntry(func(entry *InternalLogEntry) {
			entry.Output = "Blocked from executing due to concurrency limit"
			entry.Blocked = true
		})
		return false
	}

	return true
}

func parseDuration(rate config.RateSpec) time.Duration {
	duration, err := time.ParseDuration(rate.Duration)

	if err != nil {
		log.Warnf("Could not parse duration: %v", rate.Duration)

		return -1 * time.Minute
	}

	return duration
}

func entityPrefixForRequest(req *ExecutionRequest) string {
	if req.Binding != nil && req.Binding.Entity != nil {
		return req.Binding.Entity.UniqueKey
	}

	return ""
}

func rateExecutionMatchesScope(logEntry *InternalLogEntry, req *ExecutionRequest, entityPrefix string) bool {
	if logEntry.EntityPrefix != entityPrefix {
		return false
	}

	return !logEntry.Queued && logEntry.ExecutionTrackingID != req.TrackingID
}

func logEntryStartedInWindow(logEntry *InternalLogEntry, windowStart time.Time) bool {
	return logEntry.DatetimeStarted.After(windowStart) && !logEntry.Blocked
}

func rateExecutionCountsForRate(logEntry *InternalLogEntry, req *ExecutionRequest, entityPrefix string, windowStart time.Time) bool {
	return rateExecutionMatchesScope(logEntry, req, entityPrefix) && logEntryStartedInWindow(logEntry, windowStart)
}

func countRateExecutions(logs []*InternalLogEntry, req *ExecutionRequest, entityPrefix string, windowStart time.Time) int {
	executions := 0

	for _, logEntry := range logs {
		if rateExecutionCountsForRate(logEntry, req, entityPrefix, windowStart) {
			executions += 1
		}
	}

	return executions
}

func getExecutionsCount(rate config.RateSpec, req *ExecutionRequest) int {
	duration := parseDuration(rate)
	then := time.Now().Add(-duration)

	req.executor.logmutex.RLock()
	logs := req.executor.LogsByBindingId[req.Binding.ID]
	executions := countRateExecutions(logs, req, entityPrefixForRequest(req), then)
	req.executor.logmutex.RUnlock()

	return executions
}

func stepRateCheck(req *ExecutionRequest) bool {
	for _, rate := range req.Binding.Action.MaxRate {
		executions := getExecutionsCount(rate, req)

		if executions >= rate.Limit {
			log.WithFields(log.Fields{
				"actionTitle": req.logEntry.ActionTitle,
				"executions":  executions,
				"limit":       rate.Limit,
				"duration":    rate.Duration,
			}).Infof("Blocked from executing due to rate limit")

			req.mutateLogEntry(func(entry *InternalLogEntry) {
				entry.Output = "Blocked from executing due to rate limit"
				entry.Blocked = true
			})
			return false
		}
	}

	return true
}

func stepACLCheck(req *ExecutionRequest) bool {
	canExec := acl.IsAllowedExec(req.Cfg, req.AuthenticatedUser, req.Binding.Action)

	if !canExec {
		req.mutateLogEntry(func(entry *InternalLogEntry) {
			entry.Output = "ACL check failed. Blocked from executing."
			entry.Blocked = true
		})

		log.WithFields(log.Fields{
			"actionTitle": req.logEntry.ActionTitle,
		}).Warnf("ACL check failed. Blocked from executing.")
	}

	return canExec
}

func stepParseArgs(req *ExecutionRequest) bool {
	ensureArgumentMap(req)

	if !hasBindingAndAction(req) {
		return fail(req, fmt.Errorf("cannot parse arguments: Binding or Action is nil"))
	}

	filterToDefinedArgumentsOnly(req)
	if err := injectSystemArgs(req); err != nil {
		return fail(req, err)
	}
	mangleInvalidArgumentValues(req)

	if hasExec(req) {
		return handleExecBranch(req)
	} else {
		return handleShellBranch(req)
	}
}

func handleExecBranch(req *ExecutionRequest) bool {
	args, err := parseActionExec(req.Arguments, req.Binding.Action, req.Binding.Entity)

	if err != nil {
		return fail(req, err)
	}

	req.useDirectExec = true
	req.execArgs = args
	return true
}

func handleShellBranch(req *ExecutionRequest) bool {
	if hasWebhookTag(req) {
		return fail(req, fmt.Errorf("webhooks cannot use Shell execution; use exec instead. See https://docs.olivetin.app/action_execution/shellvsexec.html"))
	}
	if err := checkShellArgumentSafety(req.Binding.Action); err != nil {
		return fail(req, err)
	}

	cmd, err := parseActionArguments(req)

	if err != nil {
		return fail(req, err)
	}

	req.useDirectExec = false
	req.finalParsedCommand = cmd
	return true
}

func ensureArgumentMap(req *ExecutionRequest) {
	if req.Arguments == nil {
		req.Arguments = make(map[string]string)
	}
}

func filterToDefinedArgumentsOnly(req *ExecutionRequest) {
	definedNames := make(map[string]struct{})
	for _, arg := range req.Binding.Action.Arguments {
		definedNames[arg.Name] = struct{}{}
	}
	filtered := make(map[string]string)
	for k, v := range req.Arguments {
		if keepArgument(k, definedNames) {
			filtered[k] = v
		}
	}
	req.Arguments = filtered
}

func keepArgument(name string, definedNames map[string]struct{}) bool {
	_, ok := definedNames[name]
	return ok
}

func hasWebhookTag(req *ExecutionRequest) bool {
	for _, tag := range req.Tags {
		if tag == "webhook" {
			return true
		}
	}
	return false
}

var systemArgumentDefinitions = []config.ActionArgument{
	{Name: "ot_executionTrackingId", Type: "ascii_identifier", RejectNull: true},
	{Name: "ot_username", Type: "shell_safe_identifier", RejectNull: true},
}

func injectSystemArgs(req *ExecutionRequest) error {
	args, err := validatedSystemArgs(req)
	if err != nil {
		return err
	}

	for name, value := range args {
		req.Arguments[name] = value
	}

	return nil
}

func validatedSystemArgs(req *ExecutionRequest) (map[string]string, error) {
	values := map[string]string{
		"ot_executionTrackingId": req.TrackingID,
		"ot_username":            req.AuthenticatedUser.Username,
	}

	for i := range systemArgumentDefinitions {
		arg := &systemArgumentDefinitions[i]
		if err := ValidateArgument(arg, values[arg.Name], req.Binding.Action); err != nil {
			return nil, fmt.Errorf("system argument %q failed validation: %w", arg.Name, err)
		}
	}

	return values, nil
}

func hasBindingAndAction(req *ExecutionRequest) bool {
	return !(req.Binding == nil || req.Binding.Action == nil)
}

func hasExec(req *ExecutionRequest) bool {
	return len(req.Binding.Action.Exec) > 0
}

func fail(req *ExecutionRequest, err error) bool {
	req.mutateLogEntry(func(entry *InternalLogEntry) {
		entry.Output = err.Error()
	})
	log.Warn(err.Error())
	return false
}

func stepRequestAction(req *ExecutionRequest) bool {
	metricActionsRequested.Inc()

	if !stepRequestActionHasBinding(req) {
		return false
	}

	stepRequestActionPopulateLogEntry(req)
	stepRequestActionRegisterLog(req)

	log.WithFields(log.Fields{
		"actionTitle": req.logEntry.ActionTitle,
		"tags":        req.Tags,
	}).Infof("Action requested")

	notifyListenersStarted(req)

	return true
}

func stepRequestActionHasBinding(req *ExecutionRequest) bool {
	if req.Binding == nil || req.Binding.Action == nil {
		log.Warnf("Action request has no binding/action; skipping execution")
		return false
	}
	return true
}

func stepRequestActionPopulateLogEntry(req *ExecutionRequest) {
	req.mutateLogEntry(func(entry *InternalLogEntry) {
		entry.Binding = req.Binding
		entry.ActionConfigTitle = req.Binding.Action.Title
		entry.ActionTitle = tpl.ParseTemplateOfActionBeforeExec(req.Binding.Action.Title, req.Binding.Entity)
		entry.ActionIcon = req.Binding.Action.Icon
		entry.Tags = req.Tags
		if req.Binding.Entity != nil {
			entry.EntityPrefix = req.Binding.Entity.UniqueKey
		}
	})
}

func stepRequestActionRegisterLog(req *ExecutionRequest) {
	req.executor.logmutex.Lock()
	defer req.executor.logmutex.Unlock()

	if _, containsKey := req.executor.LogsByBindingId[req.Binding.ID]; !containsKey {
		req.executor.LogsByBindingId[req.Binding.ID] = make([]*InternalLogEntry, 0)
	}
	req.executor.LogsByBindingId[req.Binding.ID] = append(req.executor.LogsByBindingId[req.Binding.ID], req.logEntry)
}

func stepLogStart(req *ExecutionRequest) bool {
	log.WithFields(log.Fields{
		"actionTitle": req.logEntry.ActionTitle,
		"timeout":     req.Binding.Action.Timeout,
	}).Infof("Action started")

	return true
}

func stepLogFinish(req *ExecutionRequest) bool {
	req.mutateLogEntry(func(entry *InternalLogEntry) {
		entry.ExecutionFinished = true
	})

	log.WithFields(log.Fields{
		"actionTitle":  req.logEntry.ActionTitle,
		"outputLength": len(req.logEntry.Output),
		"timedOut":     req.logEntry.TimedOut,
		"exit":         req.logEntry.ExitCode,
	}).Infof("Action finished")

	return true
}

func notifyListenersFinished(req *ExecutionRequest) {
	for _, listener := range req.executor.listeners {
		listener.OnExecutionFinished(req.logEntry)
	}
}

func notifyListenersStarted(req *ExecutionRequest) {
	for _, listener := range req.executor.listeners {
		listener.OnExecutionStarted(req.logEntry)
	}
}

func appendErrorToStderr(req *ExecutionRequest, err error) {
	if err == nil {
		return
	}

	req.mutateLogEntry(func(entry *InternalLogEntry) {
		entry.Output = err.Error() + "\n\n" + entry.Output
	})
}

type OutputStreamer struct {
	Req    *ExecutionRequest
	output bytes.Buffer
}

func (ost *OutputStreamer) Write(o []byte) (n int, err error) {
	for _, listener := range ost.Req.executor.listeners {
		listener.OnOutputChunk(o, ost.Req.TrackingID)
	}

	return ost.output.Write(o)
}

func (ost *OutputStreamer) String() string {
	return ost.output.String()
}

func buildEnv(args map[string]string) []string {
	ret := append(os.Environ(), "OLIVETIN=1")

	for k, v := range args {
		varName := fmt.Sprintf("%v", strings.TrimSpace(strings.ToUpper(k)))

		// Skip arguments that might not have a name (eg, confirmation), as this causes weird bugs on Windows.
		if varName == "" {
			continue
		}

		ret = append(ret, fmt.Sprintf("%v=%v", varName, v))
	}

	return ret
}

func stepExec(req *ExecutionRequest) bool {
	ctx, cancel := newTimeoutContext(context.Background(), time.Duration(req.Binding.Action.Timeout)*time.Second, req.executor)
	defer cancel()
	streamer := &OutputStreamer{Req: req}
	cmd := buildCommand(ctx, req)
	if cmd == nil {
		req.mutateLogEntry(func(entry *InternalLogEntry) {
			entry.Output = "Cannot execute: no command arguments provided"
		})
		log.Warn("Cannot execute: no command arguments provided")
		return false
	}
	prepareCommand(cmd, streamer, req)
	runerr := cmd.Start()
	req.mutateLogEntry(func(entry *InternalLogEntry) {
		entry.Process = cmd.Process
	})
	ctx.setProcess(cmd.Process)
	waiterr := cmd.Wait()
	req.mutateLogEntry(func(entry *InternalLogEntry) {
		entry.ExitCode = int32(cmd.ProcessState.ExitCode())
		entry.Output = streamer.String()
	})

	appendErrorToStderr(req, runerr)
	appendErrorToStderr(req, waiterr)

	if ctx.Err() == context.DeadlineExceeded {
		log.WithFields(log.Fields{
			"actionTitle": req.logEntry.ActionTitle,
		}).Warnf("Action timed out")

		req.mutateLogEntry(func(entry *InternalLogEntry) {
			entry.TimedOut = true
			entry.Output += "OliveTin::timeout - this action timed out after " + fmt.Sprintf("%v", req.Binding.Action.Timeout) + " seconds. If you need more time for this action, set a longer timeout. See https://docs.olivetin.app/action_customization/timeouts.html for more help."
		})
	}

	req.mutateLogEntry(func(entry *InternalLogEntry) {
		entry.DatetimeFinished = time.Now()
	})

	return true
}

func buildCommand(ctx context.Context, req *ExecutionRequest) *exec.Cmd {
	if req.useDirectExec {
		return wrapCommandDirect(ctx, req.execArgs)
	}
	return wrapCommandInShell(ctx, req.finalParsedCommand)
}

func prepareCommand(cmd *exec.Cmd, streamer *OutputStreamer, req *ExecutionRequest) {
	cmd.Stdout = streamer
	cmd.Stderr = streamer
	cmd.Env = buildEnv(req.Arguments)
	req.mutateLogEntry(func(entry *InternalLogEntry) {
		entry.ExecutionStarted = true
	})
}

func stepExecAfter(req *ExecutionRequest) bool {
	ctx, cancel := newTimeoutContext(context.Background(), time.Duration(req.Binding.Action.Timeout)*time.Second, req.executor)
	defer cancel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd, args, err := buildShellAfterCommand(ctx, req, &stdout, &stderr)
	if err != nil {
		return fail(req, err)
	}
	if cmd == nil {
		return true
	}

	cmd.Env = buildEnv(args)

	runerr := cmd.Start()
	ctx.setProcess(cmd.Process)

	waiterr := cmd.Wait()

	req.mutateLogEntry(func(entry *InternalLogEntry) {
		entry.Output += "\n"
		entry.Output += "OliveTin::shellAfterCompleted stdout\n"
		entry.Output += stdout.String()
		entry.Output += "OliveTin::shellAfterCompleted stderr\n"
		entry.Output += stderr.String()
		entry.Output += "OliveTin::shellAfterCompleted errors and summary\n"
	})

	appendErrorToStderr(req, runerr)
	appendErrorToStderr(req, waiterr)

	if ctx.Err() == context.DeadlineExceeded {
		req.mutateLogEntry(func(entry *InternalLogEntry) {
			entry.Output += "Your shellAfterCompleted command timed out."
		})
	}

	req.mutateLogEntry(func(entry *InternalLogEntry) {
		entry.Output += fmt.Sprintf("Your shellAfterCompleted exited with code %v\n", cmd.ProcessState.ExitCode())
		entry.Output += "OliveTin::shellAfterCompleted output complete\n"
	})

	return true
}

func buildShellAfterCommand(ctx context.Context, req *ExecutionRequest, stdout, stderr *bytes.Buffer) (*exec.Cmd, map[string]string, error) {
	if req.Binding.Action.ShellAfterCompleted == "" {
		return nil, nil, nil
	}

	args, err := buildShellAfterArgs(req)
	if err != nil {
		return nil, nil, err
	}

	finalParsedCommand, err := tpl.ParseTemplateWithActionContext(req.Binding.Action.ShellAfterCompleted, req.Binding.Entity, args)
	if err != nil {
		msg := "Could not prepare shellAfterCompleted command: " + err.Error() + "\n"
		req.mutateLogEntry(func(entry *InternalLogEntry) {
			entry.Output += msg
		})
		log.Warn(msg)
		return nil, nil, nil
	}

	cmd := wrapCommandInShell(ctx, finalParsedCommand)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	return cmd, args, nil
}

func buildShellAfterArgs(req *ExecutionRequest) (map[string]string, error) {
	args, err := validatedSystemArgs(req)
	if err != nil {
		return nil, err
	}

	args["output"] = req.logEntry.Output
	args["exitCode"] = fmt.Sprintf("%v", req.logEntry.ExitCode)

	return args, nil
}

//gocyclo:ignore
func stepTrigger(req *ExecutionRequest) bool {
	if req.Binding.Action.Triggers == nil {
		return true
	}

	if req.TriggerDepth >= MaxTriggerDepth {
		log.WithFields(log.Fields{
			"actionTitle": req.logEntry.ActionTitle,
			"depth":       req.TriggerDepth,
		}).Warnf("Trigger action reached maximum depth of %v. Not triggering further actions.", MaxTriggerDepth)
		req.mutateLogEntry(func(entry *InternalLogEntry) {
			entry.Output += fmt.Sprintf("OliveTin::trigger - this action reached maximum trigger depth of %v. Not triggering further actions.", MaxTriggerDepth)
		})
		return true
	}

	if len(req.Tags) > 0 && req.Tags[0] == "trigger" {
		log.Warnf("Trigger action is triggering another trigger action. This is allowed, but be careful not to create trigger loops.")
	}

	triggerLoop(req)

	return true
}

func triggerLoop(req *ExecutionRequest) {
	for _, triggerTitle := range req.Binding.Action.Triggers {
		binding := req.executor.findBindingByActionTitle(triggerTitle, "")
		if binding == nil {
			log.WithFields(log.Fields{
				"triggerTitle": triggerTitle,
				"fromAction":   req.logEntry.ActionTitle,
			}).Warnf("Trigger references unknown action title; skipping")
			continue
		}
		trigger := &ExecutionRequest{
			Binding:           binding,
			TrackingID:        uuid.NewString(),
			Tags:              []string{"trigger"},
			AuthenticatedUser: req.AuthenticatedUser,
			Arguments:         req.Arguments,
			Cfg:               req.Cfg,
			TriggerDepth:      req.TriggerDepth + 1,
		}

		req.executor.ExecRequest(trigger)
	}
}

func stepSaveLog(req *ExecutionRequest) bool {
	filename := fmt.Sprintf("%v.%v.%v", req.logEntry.ActionTitle, req.logEntry.DatetimeStarted.Unix(), req.logEntry.ExecutionTrackingID)

	saveLogResults(req, filename)
	saveLogOutput(req, filename)

	return true
}

func firstNonEmpty(one, two string) string {
	if one != "" {
		return one
	}

	return two
}

func saveLogResults(req *ExecutionRequest, filename string) {
	dir := firstNonEmpty(req.Binding.Action.SaveLogs.ResultsDirectory, req.Cfg.SaveLogs.ResultsDirectory)

	if dir != "" {
		data, err := yaml.Marshal(req.logEntry)

		if err != nil {
			log.Warnf("%v", err)
		}

		filepath := path.Join(dir, filename+".yaml")
		err = os.WriteFile(filepath, data, 0600)

		if err != nil {
			log.Warnf("%v", err)
		}
	}
}

func saveLogOutput(req *ExecutionRequest, filename string) {
	dir := firstNonEmpty(req.Binding.Action.SaveLogs.OutputDirectory, req.Cfg.SaveLogs.OutputDirectory)

	if dir != "" {
		data := req.logEntry.Output
		filepath := path.Join(dir, filename+".log")
		err := os.WriteFile(filepath, []byte(data), 0600)

		if err != nil {
			log.Warnf("%v", err)
		}
	}
}
