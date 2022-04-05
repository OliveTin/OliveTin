package executor

import (
	pb "github.com/jamesread/OliveTin/gen/grpc"
	acl "github.com/jamesread/OliveTin/internal/acl"
	config "github.com/jamesread/OliveTin/internal/config"
	log "github.com/sirupsen/logrus"

	"bytes"
	"context"
	"errors"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

var (
	typecheckRegex = map[string]string{
		"very_dangerous_raw_string": "",
		"int":                       "^[\\d]+$",
		"ascii":                     "^[a-zA-Z0-9]+$",
		"ascii_identifier":          "^[a-zA-Z0-9\\-\\.\\_]+$",
		"ascii_sentence":            "^[a-zA-Z0-9 \\,\\.]+$",
	}
)

// InternalLogEntry objects are created by an Executor, and represent the final
// state of execution (even if the command is not executed). It's designed to be
// easily serializable.
type InternalLogEntry struct {
	Datetime string
	Stdout   string
	Stderr   string
	TimedOut bool
	ExitCode int32

	/*
		The following two properties are obviously on Action normally, but it's useful
		that logs are lightweight (so we don't need to have an action associated to
		logs, etc. Therefore, we duplicate those values here.
	*/
	ActionTitle string
	ActionIcon  string
}

// ExecutionRequest is a request to execute an action. It's passed to an
// Executor. They're created from the grpcapi.
type ExecutionRequest struct {
	ActionName         string
	Arguments          map[string]string
	action             *config.Action
	Cfg                *config.Config
	User               *acl.User
	logEntry           *InternalLogEntry
	finalParsedCommand string
}

type executorStep interface {
	Exec(*ExecutionRequest) bool
}

// Executor represents a helper class for executing commands. It's main method
// is ExecRequest
type Executor struct {
	Logs []InternalLogEntry

	chainOfCommand []executorStep
}

// DefaultExecutor returns an Executor, with a sensible "chain of command" for
// executing actions.
func DefaultExecutor() *Executor {
	e := Executor{}
	e.chainOfCommand = []executorStep{
		stepFindAction{},
		stepACLCheck{},
		stepParseArgs{},
		stepLogStart{},
		stepExec{},
		stepLogFinish{},
	}

	return &e
}

type stepFindAction struct{}

func (s stepFindAction) Exec(req *ExecutionRequest) bool {
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

type stepACLCheck struct{}

func (s stepACLCheck) Exec(req *ExecutionRequest) bool {
	return acl.IsAllowedExec(req.Cfg, req.User, req.action)
}

// ExecRequest processes an ExecutionRequest
func (e *Executor) ExecRequest(req *ExecutionRequest) *pb.StartActionResponse {
	req.logEntry = &InternalLogEntry{
		Datetime:    time.Now().Format("2006-01-02 15:04:05"),
		ActionTitle: req.ActionName,
		Stdout:      "",
		Stderr:      "",
		ExitCode:    -1337, // If an Action is not actually executed, this is the default exit code.
	}

	for _, step := range e.chainOfCommand {
		if !step.Exec(req) {
			break
		}
	}

	e.Logs = append(e.Logs, *req.logEntry)

	return &pb.StartActionResponse{
		LogEntry: &pb.LogEntry{
			ActionTitle: req.logEntry.ActionTitle,
			ActionIcon:  req.logEntry.ActionIcon,
			Datetime:    req.logEntry.Datetime,
			Stderr:      req.logEntry.Stderr,
			Stdout:      req.logEntry.Stdout,
			TimedOut:    req.logEntry.TimedOut,
			ExitCode:    req.logEntry.ExitCode,
		},
	}
}

type stepLogStart struct{}

func (e stepLogStart) Exec(req *ExecutionRequest) bool {
	log.WithFields(log.Fields{
		"title":   req.action.Title,
		"timeout": req.action.Timeout,
	}).Infof("Action starting")

	return true
}

type stepLogFinish struct{}

func (e stepLogFinish) Exec(req *ExecutionRequest) bool {
	log.WithFields(log.Fields{
		"title":    req.action.Title,
		"stdout":   req.logEntry.Stdout,
		"stderr":   req.logEntry.Stderr,
		"timedOut": req.logEntry.TimedOut,
		"exit":     req.logEntry.ExitCode,
	}).Infof("Action finished")

	return true
}

type stepParseArgs struct{}

func (e stepParseArgs) Exec(req *ExecutionRequest) bool {
	var err error

	req.finalParsedCommand, err = parseActionArguments(req.action.Shell, req.Arguments, req.action)

	if err != nil {
		req.logEntry.Stdout = err.Error()

		log.Warnf(err.Error())

		return false
	}

	return true
}

type stepExec struct{}

func (e stepExec) Exec(req *ExecutionRequest) bool {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(req.action.Timeout)*time.Second)
	defer cancel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.CommandContext(ctx, "sh", "-c", req.finalParsedCommand)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runerr := cmd.Run()

	req.logEntry.ExitCode = int32(cmd.ProcessState.ExitCode())
	req.logEntry.Stdout = stdout.String()
	req.logEntry.Stderr = stderr.String()

	if runerr != nil {
		req.logEntry.Stderr = runerr.Error() + "\n\n" + req.logEntry.Stderr
	}

	if ctx.Err() == context.DeadlineExceeded {
		req.logEntry.TimedOut = true
	}

	return true
}

func parseActionArguments(rawShellCommand string, values map[string]string, action *config.Action) (string, error) {
	log.WithFields(log.Fields{
		"cmd": rawShellCommand,
	}).Infof("Before Parse Args")

	r := regexp.MustCompile("{{ *?([a-z]+?) *?}}")
	matches := r.FindAllStringSubmatch(rawShellCommand, -1)

	for _, match := range matches {
		argValue, argProvided := values[match[1]]

		if !argProvided {
			log.Infof("%v", values)
			return "", errors.New("Required arg not provided: " + match[1])
		}

		err := typecheckActionArgument(match[1], argValue, action)

		if err != nil {
			return "", err
		}

		log.WithFields(log.Fields{
			"name":  match[1],
			"value": argValue,
		}).Debugf("Arg assigned")

		rawShellCommand = strings.ReplaceAll(rawShellCommand, match[0], argValue)
	}

	log.WithFields(log.Fields{
		"cmd": rawShellCommand,
	}).Infof("After Parse Args")

	return rawShellCommand, nil
}

func typecheckActionArgument(name string, value string, action *config.Action) error {
	arg := action.FindArg(name)

	if arg == nil {
		return errors.New("Action arg not defined: " + name)
	}

	if len(arg.Choices) > 0 {
		return typecheckChoice(value, arg)
	}

	return TypeSafetyCheck(name, value, arg.Type)
}

func typecheckChoice(value string, arg *config.ActionArgument) error {
	for _, choice := range arg.Choices {
		if value == choice.Value {
			return nil
		}
	}

	return errors.New("argument value is not one of the predefined choices")
}

// TypeSafetyCheck checks argument values match a specific type. The types are
// defined in typecheckRegex, and, you guessed it, uses regex to check for allowed
// characters.
func TypeSafetyCheck(name string, value string, typ string) error {
	pattern, found := typecheckRegex[typ]

	if !found {
		return errors.New("argument type not implemented " + typ)
	}

	matches, _ := regexp.MatchString(pattern, value)

	if !matches {
		log.WithFields(log.Fields{
			"name":  name,
			"type":  typ,
			"value": value,
		}).Warn("Arg type check safety failure")

		return errors.New("invalid argument, doesn't match " + typ)
	}

	return nil
}
