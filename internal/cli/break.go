package cli

import (
	"fmt"

	"github.com/ajaykumarsingh/flow/internal/app"
	"github.com/ajaykumarsingh/flow/internal/task"
	"github.com/spf13/cobra"
)

func newBreakCmd(a *app.App) *cobra.Command {
	return &cobra.Command{
		Use:     "break <task-id> <step1> <step2> ...",
		Short:   "Break a task into micro-steps",
		Example: `  flow break 01ARZ3N "open the doc" "write intro" "fill section 1"`,
		Args:    cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			parentID := args[0]
			parent, err := a.Tasks.Get(parentID)
			if err != nil {
				parent, err = findByPrefix(a.Tasks, parentID)
				if err != nil {
					return fmt.Errorf("task not found: %s", parentID)
				}
			}
			fmt.Printf("\n  Breaking down: %s\n\n", styleTask.Bold(true).Render(parent.Title))
			for i, step := range args[1:] {
				sub, err := a.Tasks.Add(step, task.SizeXS, parent.Energy, nil, &parent.ID)
				if err != nil {
					return fmt.Errorf("create step %d: %w", i+1, err)
				}
				fmt.Printf("  %s  %s\n", styleAccent2.Render(fmt.Sprintf("%d.", i+1)), styleTask.Render(sub.Title))
			}
			fmt.Println()
			return nil
		},
	}
}
