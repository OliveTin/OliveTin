//go:build windows
// +build windows

package executor

import (
	"context"
	"os"
	"os/exec"
)

func (e *Executor) Kill(execReq *InternalLogEntry) error {
	return execReq.Process.Kill()
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
