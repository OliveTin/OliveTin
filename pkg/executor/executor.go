package executor 

import (
	config "github.com/jamesread/OliveTin/pkg/config"
	log "github.com/sirupsen/logrus"
	pb "github.com/jamesread/OliveTin/gen/grpc"

	"errors"
	"os/exec"
	"context"
	"time"
)

var (
	Cfg *config.Config;
)

func ExecAction(action string) (*pb.StartActionResponse) {
	res := &pb.StartActionResponse{}
	res.TimedOut = false;

	log.WithFields(log.Fields{
		"actionName": action,
	}).Infof("StartAction")

	actualAction, err := findAction(action)

	if err != nil {
		log.Errorf("Error finding action %s, %s", err, action)
		return res
	}

	log.WithFields(log.Fields {
		"title": actualAction.Title,
		"timeout": actualAction.Timeout,
	}).Infof("Found action")

	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", actualAction.Shell)
	stdout, stderr := cmd.Output()

	res.ExitCode = int64(cmd.ProcessState.ExitCode())
	res.Stdout = string(stdout)

	if stderr == nil {
		res.Stderr = ""
	} else {
		res.Stderr = stderr.Error()
	}

	if ctx.Err() == context.DeadlineExceeded {
		res.TimedOut = true
	}

	log.WithFields(log.Fields {
		"stdout": res.Stdout,
		"stderr": res.Stderr,
		"timedOut": res.TimedOut,
		"exit": res.ExitCode,
	}).Infof("Finished command.")

	return res
}

func sanitizeAction(action *config.ActionButton) {
	if action.Timeout < 3 {
		action.Timeout = 3
	}
}

func findAction(actionTitle string) (*config.ActionButton, error) {
	for _, action := range Cfg.ActionButtons {
		if action.Title == actionTitle {
			sanitizeAction(&action)

			return &action, nil
		}
	}

	return nil, errors.New("Action not found")
}


