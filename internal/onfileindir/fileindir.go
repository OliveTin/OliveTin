package onfileindir

import (
	"github.com/OliveTin/OliveTin/internal/acl"
	"github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/executor"
	"github.com/OliveTin/OliveTin/internal/filehelper"
	"path/filepath"
)

func WatchFilesInDirectory(cfg *config.Config, ex *executor.Executor) {
	for _, action := range cfg.Actions {
		for _, dirname := range action.ExecOnFileChangedInDir {
			go filehelper.WatchDirectoryWrite(dirname, func(filename string) {
				scheduleExec(action, cfg, ex, filename)
			})
		}

		for _, dirname := range action.ExecOnFileCreatedInDir {
			go filehelper.WatchDirectoryCreate(dirname, func(filename string) {
				scheduleExec(action, cfg, ex, filename)
			})
		}
	}
}

func scheduleExec(action *config.Action, cfg *config.Config, ex *executor.Executor, path string) {
	req := &executor.ExecutionRequest{
		ActionTitle: action.Title,
		Cfg:         cfg,
		Tags:        []string{"fileindir"},
		Arguments: map[string]string{
			"filepath": filepath.Base(path),
			"filename": filepath.Base(path),
			"filedir": filepath.Dir(path),
			"fileext": filepath.Ext(path),

		},
		AuthenticatedUser: &acl.AuthenticatedUser{
			Username: "fileindir",
		},
	}

	ex.ExecRequest(req)
}
