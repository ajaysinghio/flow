package cli

import (
	"github.com/ajaykumarsingh/flow/internal/app"
	"github.com/ajaykumarsingh/flow/internal/tray"
	"github.com/spf13/cobra"
)

func newTrayCmd(a *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "tray",
		Short: "Run flow as a macOS menu bar / Windows system tray app",
		Long: `Adds flow to the macOS menu bar (or Windows system tray).

Shows your current suggested task and lets you mark it done, add tasks,
check in your energy level, and refresh — without opening a terminal.

Runs until you choose Quit from the menu.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			tray.Run(a)
			return nil
		},
	}
}
