package cli

import (
	"fmt"

	"github.com/ajaykumarsingh/flow/internal/app"
	"github.com/ajaykumarsingh/flow/internal/task"
	"github.com/spf13/cobra"
)

func newLsCmd(a *app.App) *cobra.Command {
	var all bool
	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			tasks, err := a.Tasks.List(all)
			if err != nil {
				return err
			}
			if len(tasks) == 0 {
				fmt.Println(styleDim.Render("\n  Queue is empty. Add something: flow add \"...\"\n"))
				return nil
			}
			fmt.Println()
			for _, t := range tasks {
				icon, st := "○", styleDim
				if t.Status == task.StatusDoing {
					icon, st = "●", styleAccent
				} else if t.Status == task.StatusDone {
					icon, st = "✓", styleGreen
				}
				fmt.Printf("  %s  %s  %s  %s\n",
					st.Render(icon),
					styleTask.Render(t.Title),
					styleDim.Render(t.ID[:8]+"…"),
					energyColor(string(t.Energy)).Render(sizeLabel(string(t.Size))),
				)
			}
			fmt.Println()
			return nil
		},
	}
	cmd.Flags().BoolVar(&all, "all", false, "Include completed tasks")
	return cmd
}
