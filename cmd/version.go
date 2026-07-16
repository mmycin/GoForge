package cmd

import (
	"fmt"
	"runtime"

	"github.com/charmbracelet/lipgloss"
	"github.com/mmycin/GoForge/internal/tui"
	"github.com/spf13/cobra"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version and build info",
		Long:  `Display a styled card with GoForge CLI version and build metadata.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(renderVersionCard())
		},
	}
}

func renderVersionCard() string {
	goVersion := runtime.Version()

	row := func(key, val string) string {
		k := lipgloss.NewStyle().Width(10).Foreground(tui.MutedColor).Render(key)
		v := tui.InfoStyle.Render(val)
		return fmt.Sprintf("   %s  %s", k, v)
	}

	inner := fmt.Sprintf("\n   %s  %s\n\n%s\n%s\n%s\n%s\n%s\n",
		lipgloss.NewStyle().Foreground(tui.SecondaryColor).Bold(true).Render(tui.IconBrand+"  GoForge CLI"),
		"",
		row("Version", "v0.1.0"),
		row("Go", goVersion),
		row("Author", "Tahcin Ul Karim Mycin"),
		row("License", "MIT"),
		row("Docs", "https://goforge.dev"),
	)

	card := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(tui.BorderColor).
		Padding(0, 1).
		Width(44).
		Render(inner)

	return "\n" + card + "\n"
}
