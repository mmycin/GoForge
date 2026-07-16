package tui

import (
	"fmt"
	"os"
)

// ── Styled log helpers ───────────────────────────────────────────────────────
// These replace the old cmd/utils.go ANSI helpers.
// In non-TTY / piped contexts they emit plain text; in a terminal they use
// Lipgloss styles defined in theme.go.

// Info prints a cyan informational line: "→  <msg>"
func Info(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	fmt.Fprintln(os.Stdout, InfoStyle.Render(IconInfo+"  ")+msg)
}

// Success prints a green success line: "✓  <msg>"
func Success(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	fmt.Fprintln(os.Stdout, SuccessStyle.Render(IconSuccess+"  ")+msg)
}

// Warn prints an amber warning line: "!  <msg>"
func Warn(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	fmt.Fprintln(os.Stderr, WarnStyle.Render(IconWarn+"  ")+msg)
}

// Error prints a red error line: "✗  <msg>"
func Error(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	fmt.Fprintln(os.Stderr, ErrorStyle.Render(IconError+"  ")+msg)
}
