package infra

import (
	"os"
	"os/exec"
)

// OsGit is the production Git implementation that shells out to the system git binary.
type OsGit struct {
	exec Executor
}

// NewOsGit returns a Git implementation that uses the provided Executor.
func NewOsGit(exec Executor) Git {
	return &OsGit{exec: exec}
}

// IsAvailable reports whether git is present in PATH.
func (g *OsGit) IsAvailable() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

// Clone performs a shallow clone of url into dest, checking out branch.
func (g *OsGit) Clone(url, dest, branch string) error {
	return g.exec.RunWithEnv(os.Environ(), "git", "clone", "--branch", branch, "--quiet", url, dest)
}

// Init initialises a new git repository in dir.
func (g *OsGit) Init(dir string) error {
	return g.exec.RunWithEnv(os.Environ(), "git", "-C", dir, "init", "--quiet")
}

// AddAll stages all files in dir.
func (g *OsGit) AddAll(dir string) error {
	return g.exec.RunWithEnv(os.Environ(), "git", "-C", dir, "add", ".")
}

// Commit creates a commit with message in dir.
func (g *OsGit) Commit(dir, message string) error {
	return g.exec.RunWithEnv(os.Environ(), "git", "-C", dir, "commit", "-m", message, "--quiet")
}
