// Package tui provides shared visual tokens, chrome components, and logging
// utilities for the GoForge TUI.  Nothing in this package calls os/exec or
// touches the file system — it is pure presentation.
package tui

import "github.com/charmbracelet/lipgloss"

// ── Color tokens ─────────────────────────────────────────────────────────────

var (
	PrimaryColor   = lipgloss.Color("#7C3AED")
	SecondaryColor = lipgloss.Color("#38BDF8")
	SuccessColor   = lipgloss.Color("#34D399")
	WarningColor   = lipgloss.Color("#FBBF24")
	ErrorColor     = lipgloss.Color("#F87171")
	MutedColor     = lipgloss.Color("#6B7280")
	BorderColor    = lipgloss.Color("#374151")
	HighlightBg    = lipgloss.Color("#1D1D2E")
	SurfaceBg      = lipgloss.Color("#111827")
)

// ── Base styles ──────────────────────────────────────────────────────────────

var (
	// Bold purple heading — used for section titles inside content areas.
	HeadingStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(PrimaryColor)

	// Secondary blue — command names, file paths, informational text.
	InfoStyle = lipgloss.NewStyle().
			Foreground(SecondaryColor)

	// Emerald green — success states, ✓ icons.
	SuccessStyle = lipgloss.NewStyle().
			Foreground(SuccessColor)

	// Amber — warnings, ! icons.
	WarnStyle = lipgloss.NewStyle().
			Foreground(WarningColor)

	// Red — errors, ✗ icons, destructive labels.
	ErrorStyle = lipgloss.NewStyle().
			Foreground(ErrorColor)

	// Gray — hints, descriptions, step counters, muted annotations.
	MutedStyle = lipgloss.NewStyle().
			Foreground(MutedColor)

	// Inline code / file paths — italic + secondary.
	CodeStyle = lipgloss.NewStyle().
			Italic(true).
			Foreground(SecondaryColor)

	// Highlighted list item background.
	HighlightStyle = lipgloss.NewStyle().
			Background(HighlightBg).
			Foreground(lipgloss.Color("#FFFFFF"))

	// Box — rounded border, used for cards and panels.
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(BorderColor).
			Padding(0, 1)

	// Content area inside a box — left-padded.
	ContentStyle = lipgloss.NewStyle().
			Padding(0, 2)

	// Divider line — full-width horizontal rule.
	DividerStyle = lipgloss.NewStyle().
			Foreground(BorderColor)

	// Danger — red-bordered box for destructive confirmation dialogs.
	DangerBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ErrorColor).
			Padding(0, 1)
)

// ── Icon constants ────────────────────────────────────────────────────────────

const (
	IconBrand   = "⚡"
	IconSuccess = "✓"
	IconInfo    = "→"
	IconPending = "·"
	IconError   = "✗"
	IconWarn    = "!"
	IconDanger  = "⚠"
	IconSelect  = "▶"
)

// ── Step-state helpers ───────────────────────────────────────────────────────

// StepIcon returns the styled icon for a step given its state.
// state: "done" | "running" | "failed" | "pending"
func StepIcon(state string) string {
	switch state {
	case "done":
		return SuccessStyle.Render(IconSuccess)
	case "running":
		return InfoStyle.Render(IconInfo)
	case "failed":
		return ErrorStyle.Render(IconError)
	default:
		return MutedStyle.Render(IconPending)
	}
}
