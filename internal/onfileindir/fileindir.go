package onfileindir

import (
	"github.com/OliveTin/OliveTin/internal/acl"
	"github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/executor"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
)

func WatchFilesInDirectory(cfg *config.Config, ex *executor.Executor) {
	for _, action := range cfg.Actions {
		for _, dirname := range action.ExecOnFileChangedInDir {
			watch(dirname, action, cfg, ex, fsnotify.Write)
		}

		for _, dirname := range action.ExecOnFileCreatedInDir {
			watch(dirname, action, cfg, ex, fsnotify.Create)
		}
	}
}

func watch(directory string, action *config.Action, cfg *config.Config, ex *executor.Executor, eventType fsnotify.Op) {
	log.WithFields(log.Fields{
		"dir":       directory,
		"eventType": eventType,
	}).Infof("Watching dir")

	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		log.Errorf("Could not watch for files being created: %v", err)
		return
	}

	defer watcher.Close()

	done := make(chan bool)

	go func() {
		for {
			processEvent(watcher, action, cfg, ex, eventType)
		}
	}()

	err = watcher.Add("/tmp")

	if err != nil {
		log.Errorf("Could not create watcher: %v", err)
	}

	<-done
}

func processEvent(watcher *fsnotify.Watcher, action *config.Action, cfg *config.Config, ex *executor.Executor, eventType fsnotify.Op) {
	select {
	case event, ok := <-watcher.Events:
		if !ok {
			return
		}

		checkEvent(&event, action, cfg, ex, eventType)
		break
	case err := <-watcher.Errors:
		log.Errorf("Error in fsnotify: %v", err)
		return
	}
}

func checkEvent(event *fsnotify.Event, action *config.Action, cfg *config.Config, ex *executor.Executor, eventType fsnotify.Op) {
	if event.Has(eventType) {
		req := &executor.ExecutionRequest{
			ActionTitle: action.Title,
			Cfg:         cfg,
			Tags:        []string{"fileindir"},
			Arguments: map[string]string{
				"filename": event.Name,
			},
			AuthenticatedUser: &acl.AuthenticatedUser{
				Username: "fileindir",
			},
		}

		ex.ExecRequest(req)
	}
}
