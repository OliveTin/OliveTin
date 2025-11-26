package onfileindir

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/OliveTin/OliveTin/internal/auth"
	"github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/executor"
	"github.com/OliveTin/OliveTin/internal/filehelper"
)

func WatchFilesInDirectory(cfg *config.Config, ex *executor.Executor) {
	for _, action := range cfg.Actions {
		for _, dirname := range action.ExecOnFileChangedInDir {
			// Pass values into anonymous function because of this issue
			// https://github.com/OliveTin/OliveTin/issues/503

			go func(act *config.Action, dir string) {
				filehelper.WatchDirectoryWrite(dir, func(filename string) {
					scheduleExec(act, cfg, ex, filename)
				})
			}(action, dirname)

			go func(act *config.Action, dir string) {
				filehelper.WatchDirectoryCreate(dir, func(filename string) {
					scheduleExec(act, cfg, ex, filename)
				})
			}(action, dirname)
		}
	}
}

func scheduleExec(action *config.Action, cfg *config.Config, ex *executor.Executor, path string) {
	args := map[string]string{
		"filepath": path,
		"filename": filepath.Base(path),
		"filedir":  filepath.Dir(path),
		"fileext":  filepath.Ext(path),
	}

	if stat, err := os.Stat(path); err == nil {
		args["filesizebytes"] = fmt.Sprintf("%v", stat.Size())
		args["filemode"] = fmt.Sprintf("%#o", stat.Mode())
		args["filemtime"] = fmt.Sprintf("%v", stat.ModTime())
		args["fileisdir"] = fmt.Sprintf("%v", stat.IsDir())
	}

	fmt.Printf("%+v", args)

	req := &executor.ExecutionRequest{
		Binding:           ex.FindBindingWithNoEntity(action),
		Cfg:               cfg,
		Tags:              []string{},
		Arguments:         args,
		AuthenticatedUser: auth.UserFromSystem(cfg, "fileindir"),
	}

	ex.ExecRequest(req)
}
