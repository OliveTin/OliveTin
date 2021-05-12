package executor 

import (
	config "github.com/jamesread/OliveTin/pkg/config"
	log "github.com/sirupsen/logrus"
	pb "github.com/jamesread/OliveTin/gen/grpc"

	"errors"
	"os/exec"
	"context"
	"time"
	"fmt"
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

	log.Infof("Found action %s", actualAction.Title)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", actualAction.Shell)
	stdout, stderr := cmd.Output()

	res.Stdout = string(stdout)
	res.Stderr = fmt.Sprintf("%s", stderr)

	if ctx.Err() == context.DeadlineExceeded {
		res.TimedOut = true
	}

	log.Infof("Command %v stdout %v", actualAction.Title, res.Stdout)
	log.Infof("Command %v stderr %v", actualAction.Title, res.Stderr)

	return res
}

func findAction(actionTitle string) (*config.ActionButton, error) {
	for _, action := range Cfg.ActionButtons {
		if action.Title == actionTitle {
			return &action, nil
		}
	}

	return nil, errors.New("Action not found")
}


