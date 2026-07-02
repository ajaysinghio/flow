package cli

import (
	"fmt"
	"strings"

	"github.com/ajaykumarsingh/flow/internal/app"
	"github.com/ajaykumarsingh/flow/internal/task"
	"github.com/spf13/cobra"
)

func newAddCmd(a *app.App) *cobra.Command {
	var due string

	cmd := &cobra.Command{
		Use:   "add <title>",
		Short: "Capture a task",
		Example: `  flow add "call the dentist"
  flow add "finish report" --due friday
  flow add "renew subscription" --due 2026-07-10`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			title := strings.Join(args, " ")
			dueDate := task.ParseDue(due)

			t, err := a.Tasks.AddWithDue(title, task.SizeM, task.EnergyMed, nil, nil, dueDate)
			if err != nil {
				return err
			}

			fmt.Printf("\n  %s %s", styleGreen.Render("✓ added"), styleTask.Render(t.Title))
			if dueDate != nil {
				fmt.Printf("  %s", styleAccent.Render("due "+task.FormatDue(t)))
			}
			fmt.Println()
			fmt.Printf("  %s\n\n", styleDim.Render(t.ID[:8]+"…"))
			return nil
		},
	}

	cmd.Flags().StringVar(&due, "due", "", "Due date: today, tomorrow, friday, next week, 2026-07-10")
	return cmd
}
