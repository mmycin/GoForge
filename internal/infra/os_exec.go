package infra

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
)

// OsExecutor is the production Executor implementation backed by os/exec.
type OsExecutor struct{}

// NewOsExecutor returns an Executor that runs real subprocesses.
func NewOsExecutor() Executor {
	return &OsExecutor{}
}

// Run executes name with args, inheriting the current process's env.
func (e *OsExecutor) Run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Run()
}

// RunInDir executes name with args inside the given directory.
func (e *OsExecutor) RunInDir(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RunWithEnv executes name with args using the provided environment slice.
// If env is nil, the current process environment is inherited.
func (e *OsExecutor) RunWithEnv(env []string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	if env != nil {
		cmd.Env = env
	}
	return cmd.Run()
}

// RunWithOutput executes name with args and returns combined stdout+stderr.
func (e *OsExecutor) RunWithOutput(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	return buf.Bytes(), err
}

// RunStreamed executes name with args and streams combined output into out.
// The context is respected — cancellation will kill the subprocess.
func (e *OsExecutor) RunStreamed(ctx context.Context, out io.Writer, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = out
	cmd.Stderr = out
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// RunStreamedWithEnv is like RunStreamed but uses the provided environment.
func (e *OsExecutor) RunStreamedWithEnv(ctx context.Context, env []string, out io.Writer, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = env
	cmd.Stdout = out
	cmd.Stderr = out
	return cmd.Run()
}
