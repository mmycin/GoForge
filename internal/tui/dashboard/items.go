// Package dashboard implements the root interactive command browser.
package dashboard

// CommandItem represents a single runnable command in the dashboard.
type CommandItem struct {
	Name      string // cobra command name, e.g. "gen:service"
	Desc      string // short description shown on the right
	HasWizard bool   // true = requires additional input before running
}
