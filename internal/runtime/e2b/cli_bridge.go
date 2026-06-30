package e2b

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"dunk/internal/runtime"
)

type cliBridge struct{}

func (cliBridge) run(ctx context.Context, ws *runtime.Workspace, cmd runtime.CommandSpec) (*runtime.CommandResult, error) {
	if _, err := exec.LookPath("e2b"); err != nil {
		return nil, missingCLIError()
	}
	args := []string{"sandbox", "exec"}
	if cmd.Workdir != "" {
		args = append(args, "--cwd", cmd.Workdir)
	}
	for k, v := range cmd.Env {
		args = append(args, "--env", k+"="+v)
	}
	args = append(args, ws.ID, cmd.Command)

	ec := exec.CommandContext(ctx, "e2b", args...)
	var stdout, stderr bytes.Buffer
	ec.Stdout = &stdout
	ec.Stderr = &stderr
	err := ec.Run()
	res := &runtime.CommandResult{Stdout: stdout.String(), Stderr: stderr.String()}
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			res.ExitCode = ee.ExitCode()
		}
		return res, fmt.Errorf("e2b exec failed: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return res, nil
}

func (cliBridge) attach(ctx context.Context, ws *runtime.Workspace, s runtime.SessionSpec, opts runtime.AttachOptions) error {
	if _, err := exec.LookPath("e2b"); err != nil {
		return missingCLIError()
	}
	fmt.Fprintf(opts.Stdout, "\nDunk synced the workspace to %s.\n", s.Workdir)
	fmt.Fprintf(opts.Stdout, "Inside the sandbox shell, run:\n  cd %s && %s\n\n", s.Workdir, s.Command)
	cmd := exec.CommandContext(ctx, "e2b", "sandbox", "connect", ws.ID)
	cmd.Stdin = opts.Stdin
	cmd.Stdout = opts.Stdout
	cmd.Stderr = opts.Stderr
	return cmd.Run()
}

func missingCLIError() error {
	return errors.New("E2B CLI not found; install with `brew install e2b` or `npm i -g @e2b/cli`")
}
