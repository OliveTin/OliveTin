//go:build !windows
// +build !windows

package executor

import (
	"context"
	"os/exec"
	"syscall"
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
		return syscall.Kill(-execReq.Process.Pid, syscall.SIGKILL)
	}
	return nil
}

func wrapCommandInShell(ctx context.Context, finalParsedCommand string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, "sh", "-c", finalParsedCommand)

	// This is to ensure that the process group is killed when the parent process is killed.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	return cmd
}

func wrapCommandDirect(ctx context.Context, execArgs []string) *exec.Cmd {
	if len(execArgs) == 0 {
		return nil
	}

	cmd := exec.CommandContext(ctx, execArgs[0], execArgs[1:]...)

	// This is to ensure that the process group is killed when the parent process is killed.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	return cmd
}

func wrapCommandExecTool(ctx context.Context, name string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, "olivetin-"+name, "exec")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return cmd
}
