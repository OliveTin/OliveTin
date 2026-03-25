//go:build windows
// +build windows

package executor

import (
	"context"
	"os"
	"os/exec"
)

func (e *Executor) Kill(execReq *InternalLogEntry) error {
	if execReq == nil {
		return nil
	}
	helper := ""
	killID := ""
	if execReq.Attributes != nil {
		helper = execReq.Attributes["helper"]
		killID = execReq.Attributes["kill_id"]
	}
	if helper != "" && killID != "" {
		killCmd := exec.CommandContext(context.Background(), "olivetin-"+helper, "kill", killID)
		_ = killCmd.Run()
	}
	if execReq.Process != nil {
		return execReq.Process.Kill()
	}
	return nil
}

func wrapCommandInShell(ctx context.Context, finalParsedCommand string) *exec.Cmd {
	winCodepage := os.Getenv("OT_WIN_FLAG_U")

	if winCodepage == "0" {
		return exec.CommandContext(ctx, "cmd", "/C", finalParsedCommand)
	} else {
		return exec.CommandContext(ctx, "cmd", "/u", "/C", finalParsedCommand)
	}
}

func wrapCommandDirect(ctx context.Context, execArgs []string) *exec.Cmd {
	if len(execArgs) == 0 {
		return nil
	}

	return exec.CommandContext(ctx, execArgs[0], execArgs[1:]...)
}

func wrapCommandExecTool(ctx context.Context, name string) *exec.Cmd {
	return exec.CommandContext(ctx, "olivetin-"+name, "exec")
}
