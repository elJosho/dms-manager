package cmd

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joshw/dms-manager/internal/tui"
	"github.com/joshw/dms-manager/pkg/dms"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive TUI",
	Long:  `Launch an interactive terminal user interface for managing DMS tasks.`,
	Run:   runTUI,
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}

func runTUI(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	client, err := dms.NewClient(ctx, GetProfile(), GetRegion())
	if err != nil {
		exitWithError(fmt.Errorf("failed to create DMS client: %w", err))
	}

	p := tea.NewProgram(tui.NewModel(client), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		exitWithError(fmt.Errorf("TUI error: %w", err))
	}
}
