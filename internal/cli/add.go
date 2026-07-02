package cli

import (
	"fmt"
	"strings"

	"github.com/ajaykumarsingh/flow/internal/store"
	"github.com/ajaykumarsingh/flow/internal/task"
	"github.com/spf13/cobra"
)

func newAddCmd(db *store.DB) *cobra.Command {
	var size, energy, parentID string
	var tags []string

	cmd := &cobra.Command{
		Use:     "add <title>",
		Short:   "Capture a task",
		Example: `  flow add "write the report" --size l --energy high`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			title := strings.Join(args, " ")
			svc := task.NewService(db)

			var pid *string
			if parentID != "" {
				pid = &parentID
			}

			t, err := svc.Add(title, task.Size(size), task.Energy(energy), tags, pid)
			if err != nil {
				return err
			}

			fmt.Printf("\n  %s %s\n",
				styleGreen.Render("✓ added"),
				styleTask.Render(t.Title),
			)
			fmt.Printf("  %s  %s  %s  %s\n\n",
				styleDim.Render(t.ID[:8]+"…"),
				styleDim.Render("size:"+string(t.Size)),
				energyColor(string(t.Energy)).Render("energy:"+string(t.Energy)),
				styleDim.Render(sizeLabel(string(t.Size))),
			)
			return nil
		},
	}

	cmd.Flags().StringVar(&size, "size", "m", "Task size: xs s m l xl")
	cmd.Flags().StringVar(&energy, "energy", "med", "Energy needed: low med high")
	cmd.Flags().StringVar(&parentID, "parent", "", "Parent task ID (creates a subtask)")
	cmd.Flags().StringSliceVar(&tags, "tag", nil, "Tags (repeatable)")
	return cmd
}
