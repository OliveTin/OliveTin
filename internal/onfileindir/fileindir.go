package onfileindir

import (
	"github.com/OliveTin/OliveTin/internal/acl"
	"github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/executor"
	"github.com/OliveTin/OliveTin/internal/filehelper"
)

func WatchFilesInDirectory(cfg *config.Config, ex *executor.Executor) {
	for _, action := range cfg.Actions {
		for _, dirname := range action.ExecOnFileChangedInDir {
			filehelper.WatchDirectoryWrite(dirname, func(filename string) {
				scheduleExec(action, cfg, ex, filename)
			})
		}

		for _, dirname := range action.ExecOnFileCreatedInDir {
			filehelper.WatchDirectoryCreate(dirname, func(filename string) {
				scheduleExec(action, cfg, ex, filename)
			})
		}
	}
}

func scheduleExec(action *config.Action, cfg *config.Config, ex *executor.Executor, filename string) {
	req := &executor.ExecutionRequest{
		ActionTitle: action.Title,
		Cfg:         cfg,
		Tags:        []string{"fileindir"},
		Arguments: map[string]string{
			"filename": filename,
		},
		AuthenticatedUser: &acl.AuthenticatedUser{
			Username: "fileindir",
		},
	}

	ex.ExecRequest(req)
}
