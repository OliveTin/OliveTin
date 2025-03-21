//go:build !windows
// +build !windows

package executor

import (
	"context"
	"os/exec"
	"syscall"
)

func (e *Executor) Kill(execReq *InternalLogEntry) error {
	// A negative PID means to kill the whole process group. This is *nix specific behavior.
	return syscall.Kill(-execReq.Process.Pid, syscall.SIGKILL)
}

func wrapCommandInShell(ctx context.Context, finalParsedCommand string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, "sh", "-c", finalParsedCommand)

	// This is to ensure that the process group is killed when the parent process is killed.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	return cmd

}
