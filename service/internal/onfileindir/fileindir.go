package onfileindir

import (
	"fmt"
	"github.com/OliveTin/OliveTin/internal/acl"
	"github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/executor"
	"github.com/OliveTin/OliveTin/internal/filehelper"
	"os"
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
		ActionTitle:       action.Title,
		Cfg:               cfg,
		Tags:              []string{},
		Arguments:         args,
		AuthenticatedUser: acl.UserFromSystem(cfg, "fileindir"),
	}

	ex.ExecRequest(req)
}
