package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

const appVersion = "v0.1.0"

// termWidth returns the current terminal width, clamped to [60, 120].
func termWidth() int {
	w, _, err := term.GetSize(0)
	if err != nil || w < 60 {
		return 64
	}
	if w > 120 {
		return 120
	}
	return w
}

// Header renders the top chrome bar.
//
//	╭──────────────────────────────────────────────────────────╮
//	│  ⚡ GoForge  ·  New Project                 Step 1 of 4  │
//	╰──────────────────────────────────────────────────────────╯
func Header(title string, extras ...string) string {
	width := termWidth()

	left := lipgloss.NewStyle().
		Bold(true).
		Foreground(PrimaryColor).
		Render(IconBrand + " GoForge")

	separator := MutedStyle.Render("  ·  ")
	center := InfoStyle.Render(title)

	right := MutedStyle.Render(appVersion)
	if len(extras) > 0 && extras[0] != "" {
		right = MutedStyle.Render(extras[0])
	}

	inner := left + separator + center
	innerW := lipgloss.Width(inner)
	rightW := lipgloss.Width(right)

	// -4 for border (2) + padding (2)
	available := width - 4
	gap := available - innerW - rightW
	if gap < 1 {
		gap = 1
	}
	content := inner + strings.Repeat(" ", gap) + right

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(BorderColor).
		Padding(0, 1).
		Width(width - 2). // -2 for the border chars themselves
		Render(content)
}

// Footer renders the bottom keybind bar.
// bindings is a flat list of alternating key/action pairs:
//
//	Footer("↑↓", "navigate", "enter", "confirm", "esc", "back")
func Footer(bindings ...string) string {
	width := termWidth()

	var parts []string
	for i := 0; i+1 < len(bindings); i += 2 {
		key := lipgloss.NewStyle().
			Bold(true).
			Foreground(SecondaryColor).
			Render(bindings[i])
		action := MutedStyle.Render(bindings[i+1])
		parts = append(parts, key+"  "+action)
	}
	inner := strings.Join(parts, "   ")

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(BorderColor).
		Padding(0, 1).
		Width(width - 2).
		Render(inner)
}

// Divider renders a full-width horizontal rule line.
func Divider(width int) string {
	return DividerStyle.Render(strings.Repeat("─", width))
}

// StatusLine renders a single log line with the appropriate icon and color.
func StatusLine(level, text string) string {
	switch level {
	case "success":
		return SuccessStyle.Render(IconSuccess+"  ") + text
	case "warn":
		return WarnStyle.Render(IconWarn+"  ") + text
	case "error":
		return ErrorStyle.Render(IconError+"  ") + text
	default:
		return InfoStyle.Render(IconInfo+"  ") + text
	}
}

// Label renders a key-value row for info cards (e.g. pre-flight details).
func Label(key, value string) string {
	k := lipgloss.NewStyle().Width(12).Foreground(MutedColor).Render(key)
	v := InfoStyle.Render(value)
	return fmt.Sprintf("  %s  %s", k, v)
}
