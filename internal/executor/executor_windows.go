//go:build windows
// +build windows

package executor

import (
	"context"
	"os/exec"
)

func (e *Executor) Kill(execReq *InternalLogEntry) error {
	return execReq.Process.Kill()
}

func wrapCommandInShell(ctx context.Context, finalParsedCommand string) *exec.Cmd {
	return exec.CommandContext(ctx, "cmd", "/C", finalParsedCommand)
}
