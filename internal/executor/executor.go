package executor

import (
	acl "github.com/OliveTin/OliveTin/internal/acl"
	config "github.com/OliveTin/OliveTin/internal/config"
	log "github.com/sirupsen/logrus"

	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"sync"
	"time"
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
	ActionName         string
	Arguments          map[string]string
	UUID               string
	Tags               []string
	action             *config.Action
	Cfg                *config.Config
	AuthenticatedUser  *acl.AuthenticatedUser
	logEntry           *InternalLogEntry
	finalParsedCommand string
	executor           *Executor
}

// InternalLogEntry objects are created by an Executor, and represent the final
// state of execution (even if the command is not executed). It's designed to be
// easily serializable.
type InternalLogEntry struct {
	DatetimeStarted   string
	DatetimeFinished  string
	Stdout            string
	Stderr            string
	StdoutBuffer      io.ReadCloser
	StderrBuffer      io.ReadCloser
	TimedOut          bool
	Blocked           bool
	ExitCode          int32
	Tags              []string
	ExecutionStarted  bool
	ExecutionFinished bool

	/*
		The following 3 properties are obviously on Action normally, but it's useful
		that logs are lightweight (so we don't need to have an action associated to
		logs, etc. Therefore, we duplicate those values here.
	*/
	ActionTitle string
	ActionIcon  string
	UUID        string
}

type executorStepFunc func(*ExecutionRequest) bool

// DefaultExecutor returns an Executor, with a sensible "chain of command" for
// executing actions.
func DefaultExecutor() *Executor {
	e := Executor{}
	e.Logs = make(map[string]*InternalLogEntry)

	e.chainOfCommand = []executorStepFunc{
		stepLogRequested,
		stepFindAction,
		stepConcurrencyCheck,
		stepACLCheck,
		stepParseArgs,
		stepLogStart,
		stepExec,
		stepExecAfter,
		stepLogFinish,
	}

	return &e
}

type listener interface {
	OnExecutionStarted(actionName string)
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
		DatetimeStarted:   time.Now().Format("2006-01-02 15:04:05"),
		ActionTitle:       req.ActionName,
		UUID:              req.UUID,
		Stdout:            "",
		Stderr:            "",
		ExitCode:          -1337, // If an Action is not actually executed, this is the default exit code.
		ExecutionStarted:  false,
		ExecutionFinished: false,
	}

	e.Logs[req.UUID] = req.logEntry

	for _, listener := range e.listeners {
		listener.OnExecutionStarted(req.ActionName)
	}

	wg := new(sync.WaitGroup)
	wg.Add(1)

	go func() {
		e.execChain(req)
		defer wg.Done()
	}()

	return wg, req.UUID
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
		if log.ActionTitle == req.ActionName && !log.ExecutionFinished {
			concurrentCount += 1
		}
	}

	return concurrentCount
}

func stepConcurrencyCheck(req *ExecutionRequest) bool {
	concurrentCount := getConcurrentCount(req)

	// Note that the current execution is counted int the logs, so when checking we +1
	if concurrentCount >= (req.action.MaxConcurrent + 1) {
		msg := fmt.Sprintf("Blocked from executing. This would mean this action is running %d times concurrently, but this action has maxExecutions set to %d.", concurrentCount, req.action.MaxConcurrent)

		log.Warnf(msg)

		req.logEntry.Stdout = msg
		req.logEntry.Blocked = true
		return false
	}

	return true
}

func stepFindAction(req *ExecutionRequest) bool {
	actualAction := req.Cfg.FindAction(req.ActionName)

	if actualAction == nil {
		log.WithFields(log.Fields{
			"actionName": req.ActionName,
		}).Warnf("Action not found")

		req.logEntry.Stderr = "Action not found"

		return false
	}

	req.action = actualAction
	req.logEntry.ActionIcon = actualAction.Icon

	return true
}

func stepACLCheck(req *ExecutionRequest) bool {
	return acl.IsAllowedExec(req.Cfg, req.AuthenticatedUser, req.action)
}

func stepParseArgs(req *ExecutionRequest) bool {
	var err error

	req.finalParsedCommand, err = parseActionArguments(req.action.Shell, req.Arguments, req.action)

	if err != nil {
		req.logEntry.Stdout = err.Error()

		log.Warnf(err.Error())

		return false
	}

	return true
}

func stepLogRequested(req *ExecutionRequest) bool {
	log.WithFields(log.Fields{
		"actionTitle": req.ActionName,
	}).Infof("Action requested")

	return true
}

func stepLogStart(req *ExecutionRequest) bool {
	log.WithFields(log.Fields{
		"actionTitle": req.action.Title,
		"timeout":     req.action.Timeout,
	}).Infof("Action starting")

	return true
}

func stepLogFinish(req *ExecutionRequest) bool {
	log.WithFields(log.Fields{
		"actionTitle": req.action.Title,
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(req.action.Timeout)*time.Second)
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
	if req.action.ShellAfterCompleted == "" {
		return true
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(req.action.Timeout)*time.Second)
	defer cancel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	args := map[string]string{
		"stdout":   req.logEntry.Stdout,
		"exitCode": fmt.Sprintf("%v", req.logEntry.ExitCode),
	}

	finalParsedCommand, _ := parseActionArguments(req.action.ShellAfterCompleted, args, req.action)

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
