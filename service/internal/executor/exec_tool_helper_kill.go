package executor

import (
	"context"
	"os/exec"
)

func runExecToolHelperKillCommand(attrs map[string]string) {
	helper, killID := "", ""
	if attrs != nil {
		helper = attrs["helper"]
		killID = attrs["kill_id"]
	}
	if helper == "" || killID == "" {
		return
	}
	killCmd := exec.CommandContext(context.Background(), "olivetin-"+helper, "kill", killID)
	_ = killCmd.Run()
}
