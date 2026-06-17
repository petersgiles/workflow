package script

import (
	"context"
	"os/exec"
	"time"
)

type OSRunner struct{}

func NewOSRunner() *OSRunner { return &OSRunner{} }

func (r *OSRunner) Exec(ctx context.Context, cmd string, args []string, env map[string]string, timeout time.Duration) (string, string, int, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	c := exec.CommandContext(ctx, cmd, args...)
	out, err := c.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return string(out), string(exitErr.Stderr), exitErr.ExitCode(), nil
		}
		return string(out), "", -1, err
	}
	return string(out), "", 0, nil
}
