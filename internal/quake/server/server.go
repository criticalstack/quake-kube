package server

import (
	"context"
	"os"
	"os/exec"
)

func Start(ctx context.Context, dir string) error {
	cmd := exec.CommandContext(ctx, "ioq3ded", "+set", "dedicated", "1", "+exec", "server.cfg", "+exec", "maps.cfg")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Wait()
}
