package executor

import (
	"syscall"
	"context"
	"os/exec"
)

func (e *Executor) Kill(execReq *InternalLogEntry) error {
	// A negative PID means to kill the whole process group. This is *nix specific behavior.
	return syscall.Kill(-execReq.Process.Pid, syscall.SIGKILL)
}

func wrapCommandInShell(ctx context.Context, finalParsedCommand string) *exec.Cmd {
	return exec.CommandContext(ctx, "cmd", "/C", finalParsedCommand)
}
