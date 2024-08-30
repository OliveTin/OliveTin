package oncalendarfile

import (
	"context"
	"github.com/OliveTin/OliveTin/internal/acl"
	"github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/executor"
	"github.com/OliveTin/OliveTin/internal/filehelper"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

var calendar map[*config.Action][]*time.Timer

func Schedule(cfg *config.Config, ex *executor.Executor) {
	calendar = make(map[*config.Action][]*time.Timer)

	for _, action := range cfg.Actions {
		if action.ExecOnCalendarFile != "" {
			x := func(filename string) {
				parseCalendarFile(action, cfg, ex, filename)
			}

			go filehelper.WatchFileWrite(action.ExecOnCalendarFile, x)

			x(action.ExecOnCalendarFile)
		}
	}
}

func parseCalendarFile(action *config.Action, cfg *config.Config, ex *executor.Executor, filename string) {
	filehelper.Touch(action.ExecOnCalendarFile, "calendar file")

	log.WithFields(log.Fields{
		"actionTitle": action.Title,
		"filename":    filename,
	}).Infof("Parsing calendar file")

	yfile, err := os.ReadFile(action.ExecOnCalendarFile)

	if err != nil {
		log.Errorf("ReadIn: %v", err)
		return
	}

	data := make([]string, 1)

	err = yaml.Unmarshal(yfile, &data)

	if err != nil {
		log.Errorf("Unmarshal: %v", err)
	}

	scheduleCalendarActions(data, action, cfg, ex)
}

func scheduleCalendarActions(entries []string, action *config.Action, cfg *config.Config, ex *executor.Executor) {
	ctx := context.Background()

	for _, instant := range entries {
		if instant == "" {
			continue
		}

		until, _ := time.Parse(time.RFC3339, instant)

		go sleepUntil(ctx, until, action, cfg, ex)
	}
}

func sleepUntil(ctx context.Context, instant time.Time, action *config.Action, cfg *config.Config, ex *executor.Executor) {
	if time.Now().After(instant) {
		log.WithFields(log.Fields{
			"instant":     instant,
			"actionTitle": action.Title,
		}).Warnf("Not scheduling stale calendar action")

		return
	}

	log.WithFields(log.Fields{
		"instant":     instant,
		"actionTitle": action.Title,
	}).Infof("Scheduling action on calendar")

	timer := time.NewTimer(time.Until(instant))

	defer timer.Stop()

	select {
	case <-timer.C:
		exec(instant, action, cfg, ex)
		return
	case <-ctx.Done():
		log.Infof("Cancelled scheduled action")
		return
	}
}

func exec(instant time.Time, action *config.Action, cfg *config.Config, ex *executor.Executor) {
	// calendar[action] = append(calendar[action], timer)
	log.WithFields(log.Fields{
		"instant":     instant,
		"actionTitle": action.Title,
	}).Infof("Executing action from calendar")

	req := &executor.ExecutionRequest{
		Action:            action,
		Cfg:               cfg,
		Tags:              []string{"calendar"},
		AuthenticatedUser: acl.UserFromSystem(cfg, "calendar"),
	}

	ex.ExecRequest(req)
}
