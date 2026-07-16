// Package infra defines the narrow interfaces that abstract all external I/O operations.
// Real OS-backed implementations live alongside this file.
// Test doubles can be injected anywhere these interfaces are accepted.
package infra

import (
	"context"
	"io"
	"io/fs"
	"os"
)

// FileSystem abstracts all file-system operations so that code generation
// logic is never coupled to the real OS file system.
type FileSystem interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
	Remove(path string) error
	RemoveAll(path string) error
	Stat(path string) (os.FileInfo, error)
	ReadDir(path string) ([]os.DirEntry, error)
	WalkDir(root string, fn fs.WalkDirFunc) error
}

// Executor abstracts os/exec invocations.
// RunStreamed pipes the subprocess's combined stdout+stderr into out while the
// process runs — suitable for feeding a live-output TUI viewport.
type Executor interface {
	Run(name string, args ...string) error
	RunInDir(dir, name string, args ...string) error
	RunWithEnv(env []string, name string, args ...string) error
	RunWithOutput(name string, args ...string) ([]byte, error)
	RunStreamed(ctx context.Context, out io.Writer, name string, args ...string) error
	RunStreamedWithEnv(ctx context.Context, env []string, out io.Writer, name string, args ...string) error
}

// Git abstracts git operations used during project creation.
type Git interface {
	IsAvailable() bool
	Clone(url, dest, branch string) error
	Init(dir string) error
	AddAll(dir string) error
	Commit(dir, message string) error
}
