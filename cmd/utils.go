// utils.go re-exports the tui logger helpers so legacy call sites in this
// package can keep using Info/Success/Warn/ErrorLog without any change.
// All styling is now handled by internal/tui — this file contains no ANSI codes.
package cmd

import "github.com/mmycin/GoForge/internal/tui"

// Info prints a styled informational line to stdout.
func Info(format string, a ...any) { tui.Info(format, a...) }

// Success prints a styled success line to stdout.
func Success(format string, a ...any) { tui.Success(format, a...) }

// Warning prints a styled warning line to stderr.
func Warning(format string, a ...any) { tui.Warn(format, a...) }

// ErrorLog prints a styled error line to stderr.
func ErrorLog(format string, a ...any) { tui.Error(format, a...) }
