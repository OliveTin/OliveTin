package onstartup

import (
	"github.com/OliveTin/OliveTin/internal/acl"
	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/executor"
	log "github.com/sirupsen/logrus"
)

func Execute(cfg *config.Config, ex *executor.Executor) {
	user := acl.UserFromSystem(cfg, "startup-user")

	for _, action := range cfg.Actions {
		if action.ExecOnStartup {
			log.WithFields(log.Fields{
				"action": action.Title,
			}).Infof("Startup action")

			req := &executor.ExecutionRequest{
				ActionTitle:       action.Title,
				Arguments:         nil,
				Cfg:               cfg,
				Tags:              []string{"startup"},
				AuthenticatedUser: user,
			}

			ex.ExecRequest(req)
		}
	}
}
