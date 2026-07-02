package cli

import (
	"fmt"

	"github.com/ajaykumarsingh/flow/internal/store"
	"github.com/ajaykumarsingh/flow/internal/task"
	"github.com/spf13/cobra"
)

func newBreakCmd(db *store.DB) *cobra.Command {
	return &cobra.Command{
		Use:     "break <task-id> <step1> <step2> ...",
		Short:   "Break a task into micro-steps",
		Example: `  flow break 01ARZ3N "open the doc" "write intro" "fill section 1"`,
		Args:    cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc := task.NewService(db)
			parentID := args[0]

			parent, err := svc.Get(parentID)
			if err != nil {
				parent, err = findByPrefix(svc, parentID)
				if err != nil {
					return fmt.Errorf("task not found: %s", parentID)
				}
			}

			steps := args[1:]
			fmt.Printf("\n  Breaking down: %s\n\n", styleTask.Bold(true).Render(parent.Title))

			for i, step := range steps {
				subtask, err := svc.Add(step, task.SizeXS, parent.Energy, nil, &parent.ID)
				if err != nil {
					return fmt.Errorf("create step %d: %w", i+1, err)
				}
				fmt.Printf("  %s  %s\n",
					styleAccent2.Render(fmt.Sprintf("%d.", i+1)),
					styleTask.Render(subtask.Title),
				)
			}
			fmt.Println()
			return nil
		},
	}
}
