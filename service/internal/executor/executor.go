package executor

import (
	acl "github.com/OliveTin/OliveTin/internal/acl"
	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/entities"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gopkg.in/yaml.v3"

	"bytes"
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

const (
	DefaultExitCodeNotExecuted = -1337
	MaxTriggerDepth            = 10
)

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
	LogsByActionId        map[string][]*InternalLogEntry

	logmutex sync.RWMutex

	MapActionIdToBinding     map[string]*ActionBinding
	MapActionIdToBindingLock sync.RWMutex

	Cfg *config.Config

	listeners []listener

	chainOfCommand []executorStepFunc
}

// ExecutionRequest is a request to execute an action. It's passed to an
// Executor. They're created from the api.
type ExecutionRequest struct {
	Binding           *ActionBinding
	Arguments         map[string]string
	TrackingID        string
	Tags              []string
	Cfg               *config.Config
	AuthenticatedUser *acl.AuthenticatedUser
	TriggerDepth      int

	logEntry           *InternalLogEntry
	finalParsedCommand string
	executor           *Executor
}

// InternalLogEntry objects are created by an Executor, and represent the final
// state of execution (even if the command is not executed). It's designed to be
// easily serializable.
type InternalLogEntry struct {
	Binding             *ActionBinding
	BindingID           string
	DatetimeStarted     time.Time
	DatetimeFinished    time.Time
	Output              string
	TimedOut            bool
	Blocked             bool
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
	ActionId    string
}

type executorStepFunc func(*ExecutionRequest) bool

// DefaultExecutor returns an Executor, with a sensible "chain of command" for
// executing actions.
func DefaultExecutor(cfg *config.Config) *Executor {
	e := Executor{}
	e.Cfg = cfg
	e.logs = make(map[string]*InternalLogEntry)
	e.logsTrackingIdsByDate = make([]string, 0)
	e.LogsByActionId = make(map[string][]*InternalLogEntry)
	e.MapActionIdToBinding = make(map[string]*ActionBinding)

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

func (e *Executor) GetLog(trackingID string) (*InternalLogEntry, bool) {
	e.logmutex.RLock()

	entry, found := e.logs[trackingID]

	e.logmutex.RUnlock()

	return entry, found
}

func (e *Executor) GetLogsByActionId(actionId string) []*InternalLogEntry {
	e.logmutex.RLock()

	logs, found := e.LogsByActionId[actionId]

	e.logmutex.RUnlock()

	if !found {
		return make([]*InternalLogEntry, 0)
	}

	return logs
}

func (e *Executor) SetLog(trackingID string, entry *InternalLogEntry) {
	e.logmutex.Lock()

	entry.Index = int64(len(e.logsTrackingIdsByDate))

	e.logs[trackingID] = entry
	e.logsTrackingIdsByDate = append(e.logsTrackingIdsByDate, trackingID)

	e.logmutex.Unlock()
}

// ExecRequest processes an ExecutionRequest
func (e *Executor) ExecRequest(req *ExecutionRequest) (*sync.WaitGroup, string) {
	if req.AuthenticatedUser == nil {
		req.AuthenticatedUser = acl.UserGuest(req.Cfg)
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
		ActionId:            "",
		ActionTitle:         "notfound",
		ActionIcon:          "&#x1f4a9;",
		Username:            req.AuthenticatedUser.Username,
	}

	_, isDuplicate := e.GetLog(req.TrackingID)

	if isDuplicate || req.TrackingID == "" {
		req.TrackingID = uuid.NewString()
	}

	log.Tracef("executor.ExecRequest(): %v", req)

	e.SetLog(req.TrackingID, req.logEntry)

	wg := new(sync.WaitGroup)
	wg.Add(1)

	go func() {
		e.execChain(req)
		defer wg.Done()
	}()

	return wg, req.TrackingID
}

func (e *Executor) execChain(req *ExecutionRequest) {
	for _, step := range e.chainOfCommand {
		if !step(req) {
			break
		}
	}

	req.logEntry.ExecutionFinished = true

	// This isn't a step, because we want to notify all listeners, irrespective
	// of how many steps were actually executed.
	notifyListenersFinished(req)
}

func getConcurrentCount(req *ExecutionRequest) int {
	concurrentCount := 0

	req.executor.logmutex.RLock()

	for _, log := range req.executor.GetLogsByActionId(req.Binding.Action.ID) {
		if !log.ExecutionFinished {
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

		req.logEntry.Output = "Blocked from executing due to concurrency limit"
		req.logEntry.Blocked = true
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

//gocyclo:ignore
func getExecutionsCount(rate config.RateSpec, req *ExecutionRequest) int {
	executions := -1 // Because we will find ourself when checking execution logs

	duration := parseDuration(rate)

	then := time.Now().Add(-duration)

	for _, logEntry := range req.executor.GetLogsByActionId(req.Binding.Action.ID) {
		// FIXME
		/*
			if logEntry.EntityPrefix != req.EntityPrefix {
				continue
			}
		*/

		if logEntry.DatetimeStarted.After(then) && !logEntry.Blocked {

			executions += 1
		}
	}

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

			req.logEntry.Output = "Blocked from executing due to rate limit"
			req.logEntry.Blocked = true
			return false
		}
	}

	return true
}

func stepACLCheck(req *ExecutionRequest) bool {
	canExec := acl.IsAllowedExec(req.Cfg, req.AuthenticatedUser, req.Binding.Action)

	if !canExec {
		req.logEntry.Output = "ACL check failed. Blocked from executing."
		req.logEntry.Blocked = true

		log.WithFields(log.Fields{
			"actionTitle": req.logEntry.ActionTitle,
		}).Warnf("ACL check failed. Blocked from executing.")
	}

	return canExec
}

func stepParseArgs(req *ExecutionRequest) bool {
	var err error

	if req.Arguments == nil {
		req.Arguments = make(map[string]string)
	}

	req.Arguments["ot_executionTrackingId"] = req.TrackingID
	req.Arguments["ot_username"] = req.AuthenticatedUser.Username

	mangleInvalidArgumentValues(req)

	req.finalParsedCommand, err = parseActionArguments(req.Arguments, req.Binding.Action, req.Binding.Entity)

	if err != nil {
		req.logEntry.Output = err.Error()

		log.Warn(err.Error())

		return false
	}

	return true
}

func stepRequestAction(req *ExecutionRequest) bool {
	metricActionsRequested.Inc()

	// If there is no binding or action, do not proceed. Leave default
	// log entry values (icon/title/id) and stop execution gracefully.
	if req.Binding == nil || req.Binding.Action == nil {
		log.Warnf("Action request has no binding/action; skipping execution")
		return false
	}

	req.logEntry.ActionConfigTitle = req.Binding.Action.Title
	req.logEntry.ActionTitle = entities.ParseTemplateWith(req.Binding.Action.Title, req.Binding.Entity)
	req.logEntry.ActionIcon = req.Binding.Action.Icon
	req.logEntry.ActionId = req.Binding.Action.ID
	req.logEntry.Tags = req.Tags

	req.executor.logmutex.Lock()

	if _, containsKey := req.executor.LogsByActionId[req.Binding.Action.ID]; !containsKey {
		req.executor.LogsByActionId[req.Binding.Action.ID] = make([]*InternalLogEntry, 0)
	}

	req.executor.LogsByActionId[req.Binding.Action.ID] = append(req.executor.LogsByActionId[req.Binding.Action.ID], req.logEntry)

	req.executor.logmutex.Unlock()

	log.WithFields(log.Fields{
		"actionTitle": req.logEntry.ActionTitle,
		"tags":        req.Tags,
	}).Infof("Action requested")

	notifyListenersStarted(req)

	return true
}

func stepLogStart(req *ExecutionRequest) bool {
	log.WithFields(log.Fields{
		"actionTitle": req.logEntry.ActionTitle,
		"timeout":     req.Binding.Action.Timeout,
	}).Infof("Action started")

	return true
}

func stepLogFinish(req *ExecutionRequest) bool {
	req.logEntry.ExecutionFinished = true

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

func appendErrorToStderr(err error, logEntry *InternalLogEntry) {
	if err != nil {
		logEntry.Output = err.Error() + "\n\n" + logEntry.Output
	}
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(req.Binding.Action.Timeout)*time.Second)
	defer cancel()

	streamer := &OutputStreamer{Req: req}

	cmd := wrapCommandInShell(ctx, req.finalParsedCommand)
	cmd.Stdout = streamer
	cmd.Stderr = streamer
	cmd.Env = buildEnv(req.Arguments)

	req.logEntry.ExecutionStarted = true

	runerr := cmd.Start()

	req.logEntry.Process = cmd.Process

	waiterr := cmd.Wait()

	req.logEntry.ExitCode = int32(cmd.ProcessState.ExitCode())
	req.logEntry.Output = streamer.String()

	appendErrorToStderr(runerr, req.logEntry)
	appendErrorToStderr(waiterr, req.logEntry)

	if ctx.Err() == context.DeadlineExceeded {
		log.WithFields(log.Fields{
			"actionTitle": req.logEntry.ActionTitle,
		}).Warnf("Action timed out")

		// The context timeout should kill the process, but let's make sure.
		err := req.executor.Kill(req.logEntry)

		if err != nil {
			log.WithFields(log.Fields{
				"actionTitle": req.logEntry.ActionTitle,
			}).Warnf("could not kill process: %v", err)
		}

		req.logEntry.TimedOut = true
		req.logEntry.Output += "OliveTin::timeout - this action timed out after " + fmt.Sprintf("%v", req.Binding.Action.Timeout) + " seconds. If you need more time for this action, set a longer timeout. See https://docs.olivetin.app/action_customization/timeouts.html for more help."
	}

	req.logEntry.DatetimeFinished = time.Now()

	return true
}

func stepExecAfter(req *ExecutionRequest) bool {
	if req.Binding.Action.ShellAfterCompleted == "" {
		return true
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(req.Binding.Action.Timeout)*time.Second)
	defer cancel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	args := map[string]string{
		"output":                 req.logEntry.Output,
		"exitCode":               fmt.Sprintf("%v", req.logEntry.ExitCode),
		"ot_executionTrackingId": req.TrackingID,
		"ot_username":            req.AuthenticatedUser.Username,
	}

	finalParsedCommand, err := parseCommandForReplacements(req.Binding.Action.ShellAfterCompleted, args, req.Binding.Entity)

	if err != nil {
		msg := "Could not prepare shellAfterCompleted command: " + err.Error() + "\n"
		req.logEntry.Output += msg
		log.Warn(msg)
		return true
	}

	cmd := wrapCommandInShell(ctx, finalParsedCommand)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	cmd.Env = buildEnv(args)

	runerr := cmd.Start()

	waiterr := cmd.Wait()

	req.logEntry.Output += "\n"
	req.logEntry.Output += "OliveTin::shellAfterCompleted stdout\n"
	req.logEntry.Output += stdout.String()

	req.logEntry.Output += "OliveTin::shellAfterCompleted stderr\n"
	req.logEntry.Output += stderr.String()

	req.logEntry.Output += "OliveTin::shellAfterCompleted errors and summary\n"
	appendErrorToStderr(runerr, req.logEntry)
	appendErrorToStderr(waiterr, req.logEntry)

	if ctx.Err() == context.DeadlineExceeded {
		req.logEntry.Output += "Your shellAfterCompleted command timed out."
	}

	req.logEntry.Output += fmt.Sprintf("Your shellAfterCompleted exited with code %v\n", cmd.ProcessState.ExitCode())

	req.logEntry.Output += "OliveTin::shellAfterCompleted output complete\n"

	return true
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
		req.logEntry.Output += fmt.Sprintf("OliveTin::trigger - this action reached maximum trigger depth of %v. Not triggering further actions.", MaxTriggerDepth)
		return true
	}

	if len(req.Tags) > 0 && req.Tags[0] == "trigger" {
		log.Warnf("Trigger action is triggering another trigger action. This is allowed, but be careful not to create trigger loops.")
	}

	triggerLoop(req)

	return true
}

func triggerLoop(req *ExecutionRequest) {
	for _, triggerReq := range req.Binding.Action.Triggers {
		binding := req.executor.FindBindingByID(triggerReq)
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
		err = os.WriteFile(filepath, data, 0644)

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
		err := os.WriteFile(filepath, []byte(data), 0644)

		if err != nil {
			log.Warnf("%v", err)
		}
	}
}
