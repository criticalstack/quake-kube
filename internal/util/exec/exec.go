package exec

import (
	"context"
	"os/exec"
)

type Cmd struct {
	*exec.Cmd
}

func (cmd *Cmd) Restart(ctx context.Context) error {
	if cmd.Process != nil {
		if err := cmd.Process.Kill(); err != nil {
			return err
		}
	}
	newCmd := exec.CommandContext(ctx, cmd.Args[0], cmd.Args[1:]...)
	newCmd.Dir = cmd.Dir
	newCmd.Env = cmd.Env
	newCmd.Stdin = cmd.Stdin
	newCmd.Stdout = cmd.Stdout
	newCmd.Stderr = cmd.Stderr
	cmd.Cmd = newCmd
	return cmd.Start()
}

func CommandContext(ctx context.Context, name string, args ...string) *Cmd {
	return &Cmd{Cmd: exec.CommandContext(ctx, name, args...)}
}
