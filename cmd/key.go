package cmd

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mmycin/GoForge/internal/tui"
	"github.com/mmycin/GoForge/internal/tui/confirm"
	"github.com/spf13/cobra"
)

func newGenKeyCmd(d *Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "gen:key",
		Short: "Generate & save a secure APP_KEY",
		Long:  `Generate a cryptographically secure 32-byte key and write it to .env as APP_KEY.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenKey(d)
		},
	}
}

func runGenKey(_ *Deps) error {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return fmt.Errorf("could not generate random key: %w", err)
	}
	encoded := base64.StdEncoding.EncodeToString(key)

	if err := updateEnvKey("APP_KEY", encoded); err != nil {
		tui.Error("Failed to write .env: %v", err)
		return err
	}

	tui.Success("APP_KEY generated and saved to .env")
	tui.Info("Key: %s", tui.CodeStyle.Render(encoded))
	return nil
}

func newRemKeyCmd(d *Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "rem:key",
		Short: "Clear the APP_KEY from .env",
		Long:  `Remove the APP_KEY value from your project's .env file.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRemKey(d)
		},
	}
}

func runRemKey(_ *Deps) error {
	body := "This will clear APP_KEY in your .env file.\n\nYou can regenerate it at any time with goforge gen:key."
	cm := confirm.New("Clear APP_KEY", body, true)
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

	if err := updateEnvKey("APP_KEY", ""); err != nil {
		tui.Error("Failed to update .env: %v", err)
		return err
	}
	tui.Success("APP_KEY cleared from .env.")
	return nil
}

// updateEnvKey sets key=value in .env, creating the file if it doesn't exist.
func updateEnvKey(key, value string) error {
	content, err := os.ReadFile(".env")
	if err != nil {
		if os.IsNotExist(err) {
			return os.WriteFile(".env", []byte(fmt.Sprintf("%s=%s\n", key, value)), 0644)
		}
		return err
	}

	lines := strings.Split(string(content), "\n")
	found := false
	for i, line := range lines {
		if strings.HasPrefix(line, key+"=") {
			lines[i] = key + "=" + value
			found = true
			break
		}
	}
	if !found {
		lines = append(lines, key+"="+value)
	}
	return os.WriteFile(".env", []byte(strings.Join(lines, "\n")), 0644)
}
