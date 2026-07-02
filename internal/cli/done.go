package cli

import (
	"fmt"
	"time"

	"github.com/ajaykumarsingh/flow/internal/mood"
	"github.com/ajaykumarsingh/flow/internal/store"
	"github.com/ajaykumarsingh/flow/internal/task"
	"github.com/spf13/cobra"
)

func newDoneCmd(db *store.DB) *cobra.Command {
	return &cobra.Command{
		Use:     "done [task-id]",
		Short:   "Mark the suggested task (or a specific task) as done",
		Example: "  flow done\n  flow done 01ARZ3N",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc := task.NewService(db)
			moods := mood.NewService(db)

			var id string
			if len(args) > 0 {
				id = args[0]
			} else {
				// resolve the currently suggested task
				checkin, _ := moods.Latest(4 * time.Hour)
				energy := 3
				if checkin != nil {
					energy = checkin.Energy
				}
				all, err := svc.List(false)
				if err != nil {
					return err
				}
				suggested := task.Suggest(all, energy)
				if suggested == nil {
					fmt.Println(styleDim.Render("  Nothing in queue to complete."))
					return nil
				}
				id = suggested.ID
			}

			t, err := svc.Get(id)
			if err != nil {
				// try prefix match
				t, err = findByPrefix(svc, id)
				if err != nil {
					return fmt.Errorf("task not found: %s", id)
				}
			}

			if err := svc.Complete(t.ID); err != nil {
				return err
			}

			fmt.Printf("\n  %s  %s\n\n",
				styleGreen.Render("✓ done"),
				styleTask.Render(t.Title),
			)
			return nil
		},
	}
}

func findByPrefix(svc *task.Service, prefix string) (*task.Task, error) {
	all, err := svc.List(true)
	if err != nil {
		return nil, err
	}
	for _, t := range all {
		if len(t.ID) >= len(prefix) && t.ID[:len(prefix)] == prefix {
			return t, nil
		}
	}
	return nil, fmt.Errorf("not found")
}
