package cmd

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/huh"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mmycin/GoForge/internal/tui"
	"github.com/mmycin/GoForge/internal/tui/progress"
	"github.com/mmycin/GoForge/internal/usecase"
	"github.com/spf13/cobra"
)

var modulePathRe = regexp.MustCompile(`^[a-zA-Z0-9._\-]+(/[a-zA-Z0-9._\-]+)+$`)

func newNewCmd(d *Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "new [project-name] [module-path]",
		Short: "Create a new GoForge project",
		Long:  `Launch a 4-step wizard to scaffold a new GoForge project from the official template, or pass project name and module path directly.`,
		Args:  cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNewWizard(d, args)
		},
	}
}

func runNewWizard(d *Deps, args []string) error {
	var (
		projectName string
		modulePath  string
		databases   []string
		initGit     = true
	)

	gitAvailable := d.Git.IsAvailable()

	if len(args) == 1 {
		return fmt.Errorf("please provide both project name and module path")
	}

	if len(args) == 2 {
		projectName = strings.TrimSpace(args[0])
		modulePath = strings.TrimSpace(args[1])
		databases = []string{"sqlite"}

		if projectName == "" {
			return fmt.Errorf("project name is required")
		}
		if modulePath == "" {
			return fmt.Errorf("module path is required")
		}
		if modulePath == "github.com/" || strings.HasSuffix(modulePath, "/") {
			return fmt.Errorf("please provide a full path, e.g. github.com/you/my-app")
		}
		if !modulePathRe.MatchString(modulePath) {
			return fmt.Errorf("must be a valid Go module path")
		}
	} else {

		// ── Step 1-4: huh form ────────────────────────────────────────────────
		form := huh.NewForm(
			// Step 1 — Project name
			huh.NewGroup(
				huh.NewInput().
					Title("Project Name").
					Description("The folder name for your project. Use kebab-case.").
					Placeholder("my-app").
					Value(&projectName).
					Validate(func(s string) error {
						if strings.TrimSpace(s) == "" {
							return fmt.Errorf("project name is required")
						}
						return nil
					}),
			),
			// Step 2 — Module path
			huh.NewGroup(
				huh.NewInput().
					Title("Module Path").
					Description("e.g. github.com/yourname/my-app\nUsed as the root Go import path throughout the project.").
					Placeholder("github.com/yourname/my-app").
					Value(&modulePath).
					Validate(func(s string) error {
						s = strings.TrimSpace(s)
						if s == "" {
							return fmt.Errorf("module path is required")
						}
						if s == "github.com/" || strings.HasSuffix(s, "/") {
							return fmt.Errorf("please provide a full path, e.g. github.com/you/my-app")
						}
						if !modulePathRe.MatchString(s) {
							return fmt.Errorf("must be a valid Go module path")
						}
						return nil
					}),
			),
			// Step 3 — Databases
			huh.NewGroup(
				huh.NewMultiSelect[string]().
					Title("Select Databases").
					Description("At least one is required. SQLite is always safe for development.").
					Options(
						huh.NewOption("SQLite  — zero-config, great for development", "sqlite"),
						huh.NewOption("MySQL   — popular relational database", "mysql"),
						huh.NewOption("PostgreSQL — advanced open-source RDBMS", "postgresql"),
						huh.NewOption("SQL Server — Microsoft enterprise database", "sqlserver"),
					).
					Value(&databases).
					Validate(func(s []string) error {
						if len(s) == 0 {
							return fmt.Errorf("select at least one database")
						}
						return nil
					}),
			),
			// Step 4 — Git init (only if git is available)
			huh.NewGroup(
				huh.NewConfirm().
					Title("Initialize a Git repository?").
					Description(gitDescription(gitAvailable)).
					Value(&initGit),
			),
		).WithTheme(huh.ThemeCatppuccin())

		p := tea.NewProgram(newWizardRunner(form, "New Project", 4), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			return err
		}
		if form.State == huh.StateAborted {
			tui.Info("Cancelled — no project was created.")
			return nil
		}
	}

	// Sanitise
	projectName = strings.TrimSpace(projectName)
	modulePath = strings.TrimSpace(modulePath)
	if len(databases) == 0 {
		databases = []string{"sqlite"}
	}

	// ── Progress screen ───────────────────────────────────────────────────────
	steps := []string{
		"Creating project structure",
		"Initialising module path",
		"Initializing database drivers",
		"Installing dependencies",
	}
	if initGit && gitAvailable {
		steps = append(steps, "Initialising git repository", "Creating initial commit")
	}

	pm := progress.New(fmt.Sprintf("Setting up %s", projectName), steps)
	progressCh := make(chan any, 64)

	input := usecase.NewProjectInput{
		ProjectName: projectName,
		ModulePath:  modulePath,
		Databases:   databases,
		InitGit:     initGit && gitAvailable,
	}

	var runErr error
	go func() {
		runErr = d.NewProject.Run(input, progressCh)
		close(progressCh)
	}()

	prog := tea.NewProgram(newProgressRunner(pm, progressCh), tea.WithAltScreen())
	if _, err := prog.Run(); err != nil {
		return err
	}

	if runErr != nil {
		tui.Error("Project creation failed: %v", runErr)
		return runErr
	}

	// ── Success screen ────────────────────────────────────────────────────────
	fmt.Println(renderSuccessCard(projectName))
	return nil
}

func gitDescription(available bool) string {
	if available {
		return "Runs git init and creates an initial commit."
	}
	return "git not found in PATH — version control setup will be skipped."
}

func renderSuccessCard(name string) string {
	tui.Success("  %s is ready\n", name)
	tui.Info("  Get started:\n")
	for _, line := range []string{
		"cd " + name,
		"cp .env.example .env",
		"goforge gen:key",
		"goforge gen:migration init",
		"goforge migrate",
		"goforge app serve",
	} {
		fmt.Printf("    %s\n", tui.CodeStyle.Render(line))
	}
	fmt.Println()
	fmt.Println("  Happy building. 🚀")
	return ""
}

// ── Internal adapter models ───────────────────────────────────────────────────

// wizardRunner is a thin tea.Model that wraps a huh.Form with GoForge chrome.
type wizardRunner struct {
	form       *huh.Form
	title      string
	totalSteps int
}

func newWizardRunner(form *huh.Form, title string, steps int) wizardRunner {
	return wizardRunner{form: form, title: title, totalSteps: steps}
}

func (w wizardRunner) Init() tea.Cmd { return w.form.Init() }

func (w wizardRunner) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	form, cmd := w.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		w.form = f
	}
	if w.form.State == huh.StateCompleted || w.form.State == huh.StateAborted {
		return w, tea.Quit
	}
	return w, cmd
}

func (w wizardRunner) View() string {
	header := tui.Header(w.title)
	footer := tui.Footer("enter", "continue", "esc", "back")
	return header + "\n\n" + w.form.View() + "\n" + footer
}

// progressRunner is a tea.Model that drives a progress.Model from a channel.
type progressRunner struct {
	pm         progress.Model
	progressCh <-chan any
	done       bool
}

func newProgressRunner(pm progress.Model, ch <-chan any) progressRunner {
	return progressRunner{pm: pm, progressCh: ch}
}

type progressTickMsg struct{ event any }
type progressDoneMsg struct{}

func (r progressRunner) Init() tea.Cmd {
	return tea.Batch(r.pm.Init(), r.waitForEvent())
}

func (r progressRunner) waitForEvent() tea.Cmd {
	return func() tea.Msg {
		event, ok := <-r.progressCh
		if !ok {
			return progressDoneMsg{}
		}
		return progressTickMsg{event: event}
	}
}

func (r progressRunner) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case progressTickMsg:
		switch e := msg.event.(type) {
		case usecase.StepStarted:
			idx := r.pm.IndexOf(e.Label)
			r.pm.Advance(idx)
		case usecase.StepDone:
			idx := r.pm.IndexOf(e.Label)
			r.pm.Complete(idx)
		case usecase.StepFailed:
			idx := r.pm.IndexOf(e.Label)
			r.pm.Fail(idx, e.Err)
		case usecase.UseCaseDone:
			r.pm.SetDone()
			r.done = true
			// Don't queue waitForEvent — we're done.
			updated, cmd := r.pm.Update(msg)
			if m, ok := updated.(progress.Model); ok {
				r.pm = m
			}
			return r, cmd
		}
		updated, cmd := r.pm.Update(msg)
		if m, ok := updated.(progress.Model); ok {
			r.pm = m
		}
		return r, tea.Batch(cmd, r.waitForEvent())

	case progressDoneMsg:
		// Channel closed without an explicit UseCaseDone — mark done.
		if !r.done {
			r.pm.SetDone()
			r.done = true
		}
		return r, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return r, tea.Quit
		case "enter", "q", "esc":
			if r.done {
				return r, tea.Quit
			}
		}

	default:
		updated, cmd := r.pm.Update(msg)
		if m, ok := updated.(progress.Model); ok {
			r.pm = m
		}
		return r, cmd
	}
	return r, nil
}

func (r progressRunner) View() string { return r.pm.View() }
