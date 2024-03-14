package main

import (
	"context"
	"os/exec"
	"time"
)

// RunCommand runs a command, not a shell, governed by the
// passed in context.
//
// Commands are given 1s to complete.
func RunCommand(ctx context.Context, name string, args ...string) (string, error) {
	cctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	cmd := exec.CommandContext(cctx, name, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}
