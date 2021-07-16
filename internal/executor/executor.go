package executor

import (
	pb "github.com/jamesread/OliveTin/gen/grpc"
	config "github.com/jamesread/OliveTin/internal/config"
	log "github.com/sirupsen/logrus"

	"context"
	"errors"
	"os/exec"
	"time"
)

// ExecAction executes an action.
func ExecAction(cfg *config.Config, action string) *pb.StartActionResponse {
	log.WithFields(log.Fields{
		"actionName": action,
	}).Infof("StartAction")

	actualAction, err := findAction(cfg, action)

	if err != nil {
		log.Errorf("Error finding action %s, %s", err, action)

		return &pb.StartActionResponse{
			TimedOut: false,
		}
	}

	return execAction(cfg, actualAction)
}

func execAction(cfg *config.Config, actualAction *config.ActionButton) *pb.StartActionResponse {
	res := &pb.StartActionResponse{}
	res.TimedOut = false

	log.WithFields(log.Fields{
		"title":   actualAction.Title,
		"timeout": actualAction.Timeout,
	}).Infof("Found action")

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(actualAction.Timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", actualAction.Shell)
	stdout, stderr := cmd.Output()

	res.ExitCode = int32(cmd.ProcessState.ExitCode())
	res.Stdout = string(stdout)

	if stderr == nil {
		res.Stderr = ""
	} else {
		res.Stderr = stderr.Error()
	}

	if ctx.Err() == context.DeadlineExceeded {
		res.TimedOut = true
	}

	log.WithFields(log.Fields{
		"stdout":   res.Stdout,
		"stderr":   res.Stderr,
		"timedOut": res.TimedOut,
		"exit":     res.ExitCode,
	}).Infof("Finished command.")

	return res
}

func sanitizeAction(action *config.ActionButton) {
	if action.Timeout < 3 {
		action.Timeout = 3
	}
}

func findAction(cfg *config.Config, actionTitle string) (*config.ActionButton, error) {
	for _, action := range cfg.ActionButtons {
		if action.Title == actionTitle {
			sanitizeAction(&action)

			return &action, nil
		}
	}

	return nil, errors.New("Action not found")
}
