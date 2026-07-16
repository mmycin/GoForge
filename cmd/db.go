package cmd

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mmycin/GoForge/internal/tui"
	"github.com/mmycin/GoForge/internal/tui/confirm"
	"github.com/mmycin/GoForge/internal/tui/progress"
	"github.com/mmycin/GoForge/internal/tui/viewport"
	"github.com/mmycin/GoForge/internal/usecase"
	"github.com/spf13/cobra"
)

// ── migrate ───────────────────────────────────────────────────────────────────

func newMigrateCmd(d *Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Apply pending database migrations",
		Long:  `Execute pending migrations to synchronise the database schema with your models.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMigrate(d)
		},
	}
}

func runMigrate(d *Deps) error {
	info := d.Migrate.Detect()

	body := fmt.Sprintf(
		"Migrator    %s\nConnection  %s\nDatabase    %s",
		info.Migrator, info.Connection,
		formatDBLabel(info.DBName, info.DBHost, info.DBPort),
	)

	cm := confirm.New("Run Database Migrations", body, false)
	cp := tea.NewProgram(cm, tea.WithAltScreen())
	finalModel, err := cp.Run()
	if err != nil {
		return err
	}
	fm, ok := finalModel.(confirm.Model)
	if !ok || !fm.Confirmed() {
		tui.Info("Cancelled.")
		return nil
	}

	return runWithViewport(d, "Migration Output", func(ch chan<- any) error {
		return d.Migrate.Run(ch)
	})
}

// ── gen:migration ─────────────────────────────────────────────────────────────

func newGenMigrationCmd(d *Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "gen:migration [name]",
		Short: "Create a new migration file",
		Long:  `Generate a new Atlas migration file for the current model state.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := ""
			if len(args) == 1 {
				name = args[0]
			}
			return runGenMigration(d, name)
		},
	}
}

func runGenMigration(d *Deps, name string) error {
	if name == "" {
		var err error
		name, err = promptInput(
			"Migration Name",
			"Describe what this migration does, in snake_case.",
			"add_users_table",
		)
		if err != nil {
			return err
		}
		if name == "" {
			tui.Info("Cancelled.")
			return nil
		}
	}

	return runWithViewport(d, fmt.Sprintf("Creating migration: %s", name), func(ch chan<- any) error {
		return d.GenMigration.Run(name, ch)
	})
}

// ── rem:migration ─────────────────────────────────────────────────────────────

func newRemMigrationCmd(d *Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "rem:migration",
		Short: "Remove the latest migration file",
		Long:  `Delete the most recent migration file and rehash the Atlas migration directory.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRemMigration(d)
		},
	}
}

func runRemMigration(d *Deps) error {
	body := "This will delete the latest .sql file in internal/database/migrations/ and update the atlas hash."
	cm := confirm.New("Remove Latest Migration", body, true)
	cp := tea.NewProgram(cm, tea.WithAltScreen())
	finalModel, err := cp.Run()
	if err != nil {
		return err
	}
	fm, ok := finalModel.(confirm.Model)
	if !ok || !fm.Confirmed() {
		tui.Info("Cancelled.")
		return nil
	}

	progressCh := make(chan any, 32)
	steps := []string{"Deleting latest migration", "Updating atlas hash"}
	pm := progress.New("Removing latest migration", steps)

	var runErr error
	go func() {
		runErr = d.RemMigration.Run(progressCh)
		close(progressCh)
	}()

	prog := tea.NewProgram(newProgressRunner(pm, progressCh), tea.WithAltScreen())
	if _, err := prog.Run(); err != nil {
		return err
	}
	if runErr != nil {
		tui.Error("rem:migration failed: %v", runErr)
	}
	return runErr
}

// ── gen:sqlc ──────────────────────────────────────────────────────────────────

func newGenSqlcCmd(d *Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "gen:sqlc",
		Short: "Generate type-safe SQL query code",
		Long:  `Run sqlc generate and inject the Sqlc field into database.go.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenSqlc(d)
		},
	}
}

func runGenSqlc(d *Deps) error {
	info := d.GenSqlc.Detect()
	body := fmt.Sprintf(
		"Engine    %s  (detected from .env)\nConfig    %s\nOutput    %s\n\n"+
			"This will run sqlc generate and inject the Sqlc field into database.go.",
		info.Engine, info.Config, info.Output,
	)

	cm := confirm.New("Generate SQLC Bindings", body, false)
	cp := tea.NewProgram(cm, tea.WithAltScreen())
	finalModel, err := cp.Run()
	if err != nil {
		return err
	}
	fm, ok := finalModel.(confirm.Model)
	if !ok || !fm.Confirmed() {
		tui.Info("Cancelled.")
		return nil
	}

	return runWithViewport(d, "SQLC Code Generation", func(ch chan<- any) error {
		return d.GenSqlc.Run(ch)
	})
}

// ── rem:sqlc ──────────────────────────────────────────────────────────────────

func newRemSqlcCmd(d *Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "rem:sqlc",
		Short: "Remove generated SQLC integration",
		Long:  `Remove the Sqlc field from database.go and delete core/database/gen/.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRemSqlc(d)
		},
	}
}

func runRemSqlc(d *Deps) error {
	body := "This will:\n  • Remove the Sqlc field from core/database/database.go\n  • Delete core/database/gen/\n\nThis cannot be undone."
	cm := confirm.New("Remove SQLC Integration", body, true)
	cp := tea.NewProgram(cm, tea.WithAltScreen())
	finalModel, err := cp.Run()
	if err != nil {
		return err
	}
	fm, ok := finalModel.(confirm.Model)
	if !ok || !fm.Confirmed() {
		tui.Info("Cancelled.")
		return nil
	}

	steps := []string{"Reverting database.go", "Deleting core/database/gen/"}
	pm := progress.New("Removing SQLC Integration", steps)
	progressCh := make(chan any, 16)
	var runErr error

	go func() {
		runErr = d.RemSqlc.Run(progressCh)
		close(progressCh)
	}()

	prog := tea.NewProgram(newProgressRunner(pm, progressCh), tea.WithAltScreen())
	if _, err := prog.Run(); err != nil {
		return err
	}
	if runErr != nil {
		tui.Error("rem:sqlc failed: %v", runErr)
	}
	return runErr
}

// ── loader ────────────────────────────────────────────────────────────────────

func newLoaderCmd(d *Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "loader",
		Short: "Inspect GORM schema output",
		Long:  `Run a temporary Go program that loads GORM models and prints the SQL schema.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLoader(d)
		},
	}
}

func runLoader(d *Deps) error {
	return runWithViewport(d, "GORM Schema Loader", func(ch chan<- any) error {
		return runLoaderUseCase(d, ch)
	})
}

// runLoaderUseCase executes the temporary loader and streams its output.
func runLoaderUseCase(d *Deps, progress usecase.Progress) error {
	from, _ := d.FS.ReadFile("go.mod")
	modulePath := "github.com/mmycin/goforge"
	for _, line := range strings.Split(string(from), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			modulePath = strings.TrimPrefix(line, "module ")
			break
		}
	}

	dbConn := "sqlite"
	if envData, _ := d.FS.ReadFile(".env"); envData != nil {
		for _, l := range strings.Split(string(envData), "\n") {
			if strings.HasPrefix(l, "DB_CONNECTION=") {
				dbConn = strings.TrimPrefix(l, "DB_CONNECTION=")
			}
		}
	}

	src := fmt.Sprintf(`package main
import (
	"fmt"
	"os"
	"ariga.io/atlas-provider-gorm/gormschema"
	"%s/internal/services"
)
func main() {
	loader := gormschema.New("%s")
	stmts, err := loader.Load(services.Model()...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(stmts)
}`, modulePath, dbConn)

	tmp := "goforge_tmp_loader.go"
	if err := d.FS.WriteFile(tmp, []byte(src), 0644); err != nil {
		return err
	}
	defer d.FS.Remove(tmp)

	usecase.SendLog(progress, "info", "Running GORM schema loader…")

	out, err := d.Exec.RunWithOutput("go", "run", tmp)
	for _, line := range strings.Split(string(out), "\n") {
		if strings.TrimSpace(line) != "" {
			usecase.SendLog(progress, "info", line)
		}
	}
	usecase.SendDone(progress)
	return err
}

// ── shared viewport runner ────────────────────────────────────────────────────

// runWithViewport runs a use-case that streams events and displays them in a
// live-output viewport. The viewport exits cleanly on completion or error.
func runWithViewport(_ *Deps, title string, run func(chan<- any) error) error {
	progressCh := make(chan any, 128)
	var runErr error

	go func() {
		runErr = run(progressCh)
		close(progressCh)
	}()

	vp := viewport.New(title, 80, 30)
	prog := tea.NewProgram(newViewportRunner(vp, progressCh), tea.WithAltScreen())
	if _, err := prog.Run(); err != nil {
		return err
	}
	return runErr
}

func formatDBLabel(dbName, host, port string) string {
	if dbName == "" {
		return "(sqlite)"
	}
	label := dbName
	if host != "" {
		label += fmt.Sprintf("  (host: %s", host)
		if port != "" {
			label += ":" + port
		}
		label += ")"
	}
	return label
}

// ── viewport runner model ─────────────────────────────────────────────────────

type viewportRunner struct {
	vp         viewport.Model
	progressCh <-chan any
	done       bool
}

func newViewportRunner(vp viewport.Model, ch <-chan any) viewportRunner {
	return viewportRunner{vp: vp, progressCh: ch}
}

type vpTickMsg struct{ event any }
type vpDoneMsg struct{ err error }

func (r viewportRunner) Init() tea.Cmd {
	return tea.Batch(r.vp.Init(), r.waitNext())
}

func (r viewportRunner) waitNext() tea.Cmd {
	return func() tea.Msg {
		e, ok := <-r.progressCh
		if !ok {
			// Channel closed — subprocess finished (successfully or with error).
			return vpDoneMsg{}
		}
		return vpTickMsg{event: e}
	}
}

func (r viewportRunner) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case vpTickMsg:
		var cmd tea.Cmd
		switch e := msg.event.(type) {
		case usecase.LogLine:
			r.vp, cmd = viewport.UpdateModel(r.vp, viewport.AppendLineMsg{Line: e.Text})
		case usecase.StepStarted:
			r.vp, cmd = viewport.UpdateModel(r.vp, viewport.AppendLineMsg{Line: tui.MutedStyle.Render("→  " + e.Label + "…")})
		case usecase.StepDone:
			r.vp, cmd = viewport.UpdateModel(r.vp, viewport.AppendLineMsg{Line: tui.SuccessStyle.Render("✓  "+e.Label)})
		case usecase.StepFailed:
			errMsg := e.Label
			if e.Err != nil {
				errMsg += ": " + e.Err.Error()
			}
			r.vp, cmd = viewport.UpdateModel(r.vp, viewport.AppendLineMsg{Line: tui.ErrorStyle.Render("✗  " + errMsg)})
			r.vp, _ = viewport.UpdateModel(r.vp, viewport.DoneMsg{Err: e.Err})
			r.done = true
			return r, tea.Batch(cmd, r.waitNext())
		case usecase.UseCaseDone:
			r.vp, cmd = viewport.UpdateModel(r.vp, viewport.DoneMsg{})
			r.done = true
			return r, cmd
		}
		return r, tea.Batch(cmd, r.waitNext())

	case vpDoneMsg:
		r.vp, _ = viewport.UpdateModel(r.vp, viewport.DoneMsg{Err: msg.err})
		r.done = true
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
		var cmd tea.Cmd
		r.vp, cmd = viewport.UpdateModel(r.vp, msg)
		return r, cmd

	case tea.WindowSizeMsg:
		var cmd tea.Cmd
		r.vp, cmd = viewport.UpdateModel(r.vp, msg)
		return r, cmd
	}
	return r, nil
}

func (r viewportRunner) View() string { return r.vp.View() }
