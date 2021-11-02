package executor

import (
	pb "github.com/jamesread/OliveTin/gen/grpc"
	acl "github.com/jamesread/OliveTin/internal/acl"
	config "github.com/jamesread/OliveTin/internal/config"
	log "github.com/sirupsen/logrus"

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

type ExecutionRequest struct {
	ActionName         string
	Arguments          map[string]string
	action             *config.Action
	Cfg                *config.Config
	User               *acl.User
	logEntry           *InternalLogEntry
	finalParsedCommand string
}

type ExecutorStep interface {
	Exec(*ExecutionRequest) bool
}

type Executor struct {
	Logs []InternalLogEntry

	chainOfCommand []ExecutorStep
}

func DefaultExecutor() *Executor {
	e := Executor{}
	e.chainOfCommand = []ExecutorStep{
		StepFindAction{},
		StepAclCheck{},
		StepParseArgs{},
		StepLogStart{},
		StepExec{},
		StepLogFinish{},
	}

	return &e
}

type StepFindAction struct{}

func (s StepFindAction) Exec(req *ExecutionRequest) bool {
	actualAction := req.Cfg.FindAction(req.ActionName)

	if actualAction == nil {
		log.WithFields(log.Fields{
			"actionName": req.ActionName,
		}).Warnf("Action not found")

		req.logEntry.Stderr = "Action not found"
		req.logEntry.ExitCode = -1337

		return false
	}

	req.action = actualAction
	req.logEntry.ActionIcon = actualAction.Icon

	return true
}

type StepAclCheck struct{}

func (s StepAclCheck) Exec(req *ExecutionRequest) bool {
	return acl.IsAllowedExec(req.Cfg, req.User, req.action)
}

// ExecRequest processes an ExecutionRequest
func (e *Executor) ExecRequest(req *ExecutionRequest) *pb.StartActionResponse {
	req.logEntry = &InternalLogEntry{
		Datetime:    time.Now().Format("2006-01-02 15:04:05"),
		ActionTitle: req.ActionName,
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

type StepLogStart struct{}

func (e StepLogStart) Exec(req *ExecutionRequest) bool {
	log.WithFields(log.Fields{
		"title":   req.action.Title,
		"timeout": req.action.Timeout,
	}).Infof("Action starting")

	return true
}

type StepLogFinish struct{}

func (e StepLogFinish) Exec(req *ExecutionRequest) bool {
	log.WithFields(log.Fields{
		"title":    req.action.Title,
		"stdout":   req.logEntry.Stdout,
		"stderr":   req.logEntry.Stderr,
		"timedOut": req.logEntry.TimedOut,
		"exit":     req.logEntry.ExitCode,
	}).Infof("Action finished")

	return true
}

type StepParseArgs struct{}

func (e StepParseArgs) Exec(req *ExecutionRequest) bool {
	var err error

	req.finalParsedCommand, err = parseActionArguments(req.action.Shell, req.Arguments, req.action)

	if err != nil {
		req.logEntry.ExitCode = -1337
		req.logEntry.Stderr = ""
		req.logEntry.Stdout = err.Error()

		log.Warnf(err.Error())

		return false
	}

	return true
}

type StepExec struct{}

func (e StepExec) Exec(req *ExecutionRequest) bool {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(req.action.Timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", req.finalParsedCommand)
	stdout, stderr := cmd.Output()

	if stderr != nil {
		req.logEntry.Stderr = stderr.Error()
	}

	if ctx.Err() == context.DeadlineExceeded {
		req.logEntry.TimedOut = true
	}

	req.logEntry.ExitCode = int32(cmd.ProcessState.ExitCode())
	req.logEntry.Stdout = string(stdout)

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

		rawShellCommand = strings.Replace(rawShellCommand, match[0], argValue, -1)
	}

	log.WithFields(log.Fields{
		"cmd": rawShellCommand,
	}).Infof("After Parse Args")

	return rawShellCommand, nil
}

func typecheckActionArgument(name string, value string, action *config.Action) error {
	arg := findArg(name, action)

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

	return errors.New("Arg value is not one of the predefined choices")
}

func TypeSafetyCheck(name string, value string, typ string) error {
	pattern, found := typecheckRegex[typ]

	log.Infof("%v %v", pattern, typ)

	if !found {
		return errors.New("Arg type not implemented " + typ)
	}

	matches, _ := regexp.MatchString(pattern, value)

	if !matches {
		log.WithFields(log.Fields{
			"name":  name,
			"type":  typ,
			"value": value,
		}).Warn("Arg type check safety failure")

		return errors.New("Invalid argument, doesn't match " + typ)
	}

	return nil
}

func findArg(name string, action *config.Action) *config.ActionArgument {
	for _, arg := range action.Arguments {
		if arg.Name == name {
			return &arg
		}
	}

	return nil
}
