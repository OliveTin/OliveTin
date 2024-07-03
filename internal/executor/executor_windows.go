package executor

import (
	"context"
	"os/exec"
)

func (e *Executor) Kill(execReq *InternalLogEntry) error {
	return execReq.Process.Kill()
}

func wrapCommandInShell(ctx context.Context, finalParsedCommand string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, "sh", "-c", finalParsedCommand)

	// This is to ensure that the process group is killed when the parent process is killed.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	return cmd

}
