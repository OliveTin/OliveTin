package oncalendarfile

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/OliveTin/OliveTin/internal/auth"
	"github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/executor"
	"github.com/OliveTin/OliveTin/internal/filehelper"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type timerEntry struct {
	timer  *time.Timer
	cancel context.CancelFunc
}

type existingTimers struct {
	timers map[time.Time]timerEntry
}

var (
	scheduleMap      = make(map[string]existingTimers)
	scheduleMapMutex sync.RWMutex
)

func Schedule(cfg *config.Config, ex *executor.Executor) {
	for _, action := range cfg.Actions {
		captured := action

		if action.ExecOnCalendarFile != "" {
			x := func(filename string) {
				parseCalendarFile(captured, cfg, ex, filename)
			}

			go filehelper.WatchFileWrite(action.ExecOnCalendarFile, x)

			x(action.ExecOnCalendarFile)
		}
	}
}

func clearExistingTimers(action *config.Action) {
	scheduleMapMutex.Lock()
	defer scheduleMapMutex.Unlock()

	if _, exists := scheduleMap[action.ID]; exists {
		for instant, entry := range scheduleMap[action.ID].timers {
			log.WithFields(log.Fields{
				"instant":     instant,
				"actionTitle": action.Title,
			}).Infof("Clearing existing scheduled action from calendar")

			entry.cancel()
			entry.timer.Stop()
		}
	}

	scheduleMap[action.ID] = existingTimers{
		timers: make(map[time.Time]timerEntry),
	}
}

func parseCalendarFile(action *config.Action, cfg *config.Config, ex *executor.Executor, filename string) {
	clearExistingTimers(action)

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

		until, err := time.Parse(time.RFC3339, instant)

		if err != nil {
			log.WithFields(log.Fields{
				"instant":     instant,
				"actionTitle": action.Title,
			}).Warnf("Invalid calendar entry, skipping: %v", err)
			continue
		}

		go sleepUntil(ctx, until, action, cfg, ex)
	}
}

func registerTimer(action *config.Action, instant time.Time, timer *time.Timer, cancel context.CancelFunc) {
	scheduleMapMutex.Lock()
	defer scheduleMapMutex.Unlock()

	if _, exists := scheduleMap[action.ID]; !exists {
		scheduleMap[action.ID] = existingTimers{
			timers: make(map[time.Time]timerEntry),
		}
	}
	scheduleMap[action.ID].timers[instant] = timerEntry{
		timer:  timer,
		cancel: cancel,
	}
}

func unregisterTimer(action *config.Action, instant time.Time) {
	scheduleMapMutex.Lock()
	v := scheduleMap[action.ID]
	if v.timers != nil {
		delete(v.timers, instant)
	}
	scheduleMapMutex.Unlock()
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

	childCtx, cancel := context.WithCancel(ctx)
	timer := time.NewTimer(time.Until(instant))

	registerTimer(action, instant, timer, cancel)

	defer timer.Stop()
	defer cancel()

	select {
	case <-timer.C:
		unregisterTimer(action, instant)
		exec(instant, action, cfg, ex)
		return
	case <-childCtx.Done():
		unregisterTimer(action, instant)
		log.Infof("Cancelled scheduled action")
		return
	}
}

func exec(instant time.Time, action *config.Action, cfg *config.Config, ex *executor.Executor) {
	log.WithFields(log.Fields{
		"instant":     instant,
		"actionTitle": action.Title,
	}).Infof("Executing action from calendar")

	req := &executor.ExecutionRequest{
		Binding:           ex.FindBindingWithNoEntity(action),
		Cfg:               cfg,
		Tags:              []string{},
		AuthenticatedUser: auth.UserFromSystem(cfg, "calendar"),
	}

	ex.ExecRequest(req)
}
