package executor

import (
	acl "github.com/OliveTin/OliveTin/internal/acl"
	config "github.com/OliveTin/OliveTin/internal/config"
	sv "github.com/OliveTin/OliveTin/internal/stringvariables"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"sync"
	"time"
)

var (
	metricActionsRequested = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "olivetin_actions_requested_count",
		Help: "The actions requested count",
	})
)

// Executor represents a helper class for executing commands. It's main method
// is ExecRequest
type Executor struct {
	Logs map[string]*InternalLogEntry

	listeners []listener

	chainOfCommand []executorStepFunc
}

// ExecutionRequest is a request to execute an action. It's passed to an
// Executor. They're created from the grpcapi.
type ExecutionRequest struct {
	ActionTitle       string
	Action            *config.Action
	Arguments         map[string]string
	TrackingID        string
	Tags              []string
	Cfg               *config.Config
	AuthenticatedUser *acl.AuthenticatedUser
	EntityPrefix      string

	logEntry           *InternalLogEntry
	finalParsedCommand string
	executor           *Executor
}

// InternalLogEntry objects are created by an Executor, and represent the final
// state of execution (even if the command is not executed). It's designed to be
// easily serializable.
type InternalLogEntry struct {
	DatetimeStarted     string
	DatetimeFinished    string
	Stdout              string
	Stderr              string
	StdoutBuffer        io.ReadCloser
	StderrBuffer        io.ReadCloser
	TimedOut            bool
	Blocked             bool
	ExitCode            int32
	Tags                []string
	ExecutionStarted    bool
	ExecutionFinished   bool
	ExecutionTrackingID string

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
func DefaultExecutor() *Executor {
	e := Executor{}
	e.Logs = make(map[string]*InternalLogEntry)

	e.chainOfCommand = []executorStepFunc{
		stepRequestAction,
		stepConcurrencyCheck,
		stepACLCheck,
		stepParseArgs,
		stepLogStart,
		stepExec,
		stepExecAfter,
		stepLogFinish,
		stepTrigger,
	}

	return &e
}

type listener interface {
	OnExecutionStarted(actionTitle string)
	OnExecutionFinished(logEntry *InternalLogEntry)
}

func (e *Executor) AddListener(m listener) {
	e.listeners = append(e.listeners, m)
}

// ExecRequest processes an ExecutionRequest
func (e *Executor) ExecRequest(req *ExecutionRequest) (*sync.WaitGroup, string) {
	req.executor = e

	// req.UUID is now set by the client, so that they can track the request
	// from start to finish. This means that a malicious client could send
	// duplicate UUIDs (or just random strings), but this is the only way.

	req.logEntry = &InternalLogEntry{
		DatetimeStarted:     time.Now().Format("2006-01-02 15:04:05"),
		ExecutionTrackingID: req.TrackingID,
		Stdout:              "",
		Stderr:              "",
		ExitCode:            -1337, // If an Action is not actually executed, this is the default exit code.
		ExecutionStarted:    false,
		ExecutionFinished:   false,
		ActionId:            "",
		ActionTitle:         "notfound",
		ActionIcon:          "&#x1f4a9;",
	}

	_, foundLog := e.Logs[req.TrackingID]

	if foundLog || req.TrackingID == "" {
		req.TrackingID = uuid.NewString()
	}

	e.Logs[req.TrackingID] = req.logEntry

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
	notifyListeners(req)
}

func getConcurrentCount(req *ExecutionRequest) int {
	concurrentCount := 0

	for _, log := range req.executor.Logs {
		if log.ActionId == req.Action.ID && !log.ExecutionFinished {
			concurrentCount += 1
		}
	}

	return concurrentCount
}

func stepConcurrencyCheck(req *ExecutionRequest) bool {
	concurrentCount := getConcurrentCount(req)

	// Note that the current execution is counted int the logs, so when checking we +1
	if concurrentCount >= (req.Action.MaxConcurrent + 1) {
		msg := fmt.Sprintf("Blocked from executing. This would mean this action is running %d times concurrently, but this action has maxExecutions set to %d.", concurrentCount, req.Action.MaxConcurrent)

		log.WithFields(log.Fields{
			"actionTitle": req.logEntry.ActionTitle,
		}).Warnf(msg)

		req.logEntry.Stdout = msg
		req.logEntry.Blocked = true
		return false
	}

	return true
}

func stepACLCheck(req *ExecutionRequest) bool {
	return acl.IsAllowedExec(req.Cfg, req.AuthenticatedUser, req.Action)
}

func stepParseArgs(req *ExecutionRequest) bool {
	var err error

	req.finalParsedCommand, err = parseActionArguments(req.Action.Shell, req.Arguments, req.Action, req.logEntry.ActionTitle, req.EntityPrefix)

	if err != nil {
		req.logEntry.Stdout = err.Error()

		log.Warnf(err.Error())

		return false
	}

	return true
}

func stepRequestAction(req *ExecutionRequest) bool {
	// The grpc API always tries to find the action by ID, but it may
	if req.Action == nil {
		log.WithFields(log.Fields{
			"actionTitle": req.ActionTitle,
		}).Infof("Action finding by title")

		req.Action = req.Cfg.FindAction(req.ActionTitle)

		if req.Action == nil {
			log.WithFields(log.Fields{
				"actionTitle": req.ActionTitle,
			}).Warnf("Action requested, but not found")

			req.logEntry.Stderr = "Action not found: " + req.ActionTitle

			return false
		}
	}

	metricActionsRequested.Inc()

	req.logEntry.ActionTitle = sv.ReplaceEntityVars(req.EntityPrefix, req.Action.Title)
	req.logEntry.ActionIcon = req.Action.Icon
	req.logEntry.ActionId = req.Action.ID

	log.WithFields(log.Fields{
		"actionTitle": req.logEntry.ActionTitle,
	}).Infof("Action requested")

	return true
}

func stepLogStart(req *ExecutionRequest) bool {
	log.WithFields(log.Fields{
		"actionTitle": req.logEntry.ActionTitle,
		"timeout":     req.Action.Timeout,
	}).Infof("Action starting")

	return true
}

func stepLogFinish(req *ExecutionRequest) bool {
	log.WithFields(log.Fields{
		"actionTitle": req.logEntry.ActionTitle,
		"stdout":      req.logEntry.Stdout,
		"stderr":      req.logEntry.Stderr,
		"timedOut":    req.logEntry.TimedOut,
		"exit":        req.logEntry.ExitCode,
	}).Infof("Action finished")

	return true
}

func notifyListeners(req *ExecutionRequest) {
	for _, listener := range req.executor.listeners {
		listener.OnExecutionFinished(req.logEntry)
	}
}

func wrapCommandInShell(ctx context.Context, finalParsedCommand string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.CommandContext(ctx, "cmd", "/C", finalParsedCommand)
	}

	return exec.CommandContext(ctx, "sh", "-c", finalParsedCommand)
}

func stepExec(req *ExecutionRequest) bool {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(req.Action.Timeout)*time.Second)
	defer cancel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := wrapCommandInShell(ctx, req.finalParsedCommand)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	req.logEntry.StdoutBuffer, _ = cmd.StdoutPipe()
	req.logEntry.StderrBuffer, _ = cmd.StderrPipe()

	req.logEntry.ExecutionStarted = true

	runerr := cmd.Start()

	cmd.Wait()

	// req.logEntry.Stdout = req.logEntry.StdoutBuffer.String()
	// req.logEntry.Stderr = req.logEntry.StderrBuffer.String()

	req.logEntry.ExitCode = int32(cmd.ProcessState.ExitCode())
	req.logEntry.Stdout = stdout.String()
	req.logEntry.Stderr = stderr.String()

	if runerr != nil {
		req.logEntry.Stderr = runerr.Error() + "\n\n" + req.logEntry.Stderr
	}

	if ctx.Err() == context.DeadlineExceeded {
		req.logEntry.TimedOut = true
	}

	req.logEntry.Tags = req.Tags
	req.logEntry.DatetimeFinished = time.Now().Format("2006-01-02 15:04:05")

	return true
}

func stepExecAfter(req *ExecutionRequest) bool {
	if req.Action.ShellAfterCompleted == "" {
		return true
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(req.Action.Timeout)*time.Second)
	defer cancel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	args := map[string]string{
		"stdout":   req.logEntry.Stdout,
		"exitCode": fmt.Sprintf("%v", req.logEntry.ExitCode),
	}

	finalParsedCommand, _ := parseActionArguments(req.Action.ShellAfterCompleted, args, req.Action, req.logEntry.ActionTitle, req.EntityPrefix)

	cmd := wrapCommandInShell(ctx, finalParsedCommand)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	req.logEntry.StdoutBuffer, _ = cmd.StdoutPipe()
	req.logEntry.StderrBuffer, _ = cmd.StderrPipe()

	runerr := cmd.Start()

	cmd.Wait()

	req.logEntry.Stdout += "---\n" + stdout.String()
	req.logEntry.Stderr += "---\n" + stderr.String()

	if runerr != nil {
		req.logEntry.Stderr = runerr.Error() + "\n\n" + req.logEntry.Stderr
	}

	if ctx.Err() == context.DeadlineExceeded {
		req.logEntry.Stderr += "Your shellAfterCommand command timed out."
	}

	req.logEntry.Stdout += fmt.Sprintf("Your shellAfterCommand exited with code %v", cmd.ProcessState.ExitCode())

	return true
}

func stepTrigger(req *ExecutionRequest) bool {
	if req.Action.Trigger != "" {
		trigger := &ExecutionRequest{
			ActionTitle:       req.Action.Trigger,
			TrackingID:        uuid.NewString(),
			Tags:              []string{"trigger"},
			AuthenticatedUser: req.AuthenticatedUser,
			Cfg:               req.Cfg,
		}

		req.executor.ExecRequest(trigger)
	}

	return true
}
