// Package usecase contains application use-cases.
// Each use-case communicates progress to the TUI layer by sending typed events
// over a channel that the caller provides.  This keeps use-cases completely
// decoupled from the presentation layer.
package usecase

// Progress event types ────────────────────────────────────────────────────────

// StepStarted signals that a named step has begun.
type StepStarted struct{ Label string }

// StepDone signals that a named step completed successfully.
type StepDone struct{ Label string }

// StepFailed signals that a named step failed.
type StepFailed struct {
	Label string
	Err   error
}

// LogLine is a single line of structured log output from a subprocess.
type LogLine struct {
	Level string // "info" | "warn" | "error"
	Text  string
}

// GeneratedFile records a single file that was created or updated.
type GeneratedFile struct {
	Path    string
	Updated bool // true = existing file was modified, false = newly created
}

// UseCaseDone signals that the use-case finished without error.
type UseCaseDone struct {
	Files []GeneratedFile
}

// Progress is the channel type passed into every use-case.
// Callers own the channel and close it after the use-case goroutine exits.
type Progress chan<- any

// send is a helper so use-cases can push events without type assertions.
func send(ch Progress, event any) {
	if ch != nil {
		ch <- event
	}
}

// SendLog is an exported helper so cmd/ can push log lines without importing internals.
func SendLog(ch Progress, level, text string) {
	send(ch, LogLine{Level: level, Text: text})
}

// SendDone is an exported helper so cmd/ can signal completion.
func SendDone(ch Progress) {
	send(ch, UseCaseDone{})
}
