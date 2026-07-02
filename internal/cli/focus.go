package cli

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ajaykumarsingh/flow/internal/focus"
	"github.com/ajaykumarsingh/flow/internal/mood"
	"github.com/ajaykumarsingh/flow/internal/store"
	"github.com/ajaykumarsingh/flow/internal/task"
	"github.com/spf13/cobra"
)

func newFocusCmd(db *store.DB) *cobra.Command {
	var minutes int

	cmd := &cobra.Command{
		Use:     "focus [task-id]",
		Short:   "Start a timed focus session",
		Example: "  flow focus\n  flow focus 01ARZ3N --minutes 45",
		RunE: func(cmd *cobra.Command, args []string) error {
			focusSvc := focus.NewService(db)
			taskSvc := task.NewService(db)
			moodSvc := mood.NewService(db)

			var taskID *string
			var taskTitle string

			if len(args) > 0 {
				t, err := taskSvc.Get(args[0])
				if err != nil {
					t, err = findByPrefix(taskSvc, args[0])
				}
				if err == nil {
					taskID = &t.ID
					taskTitle = t.Title
					_ = taskSvc.SetDoing(t.ID)
				}
			} else {
				checkin, _ := moodSvc.Latest(4 * time.Hour)
				energy := 3
				if checkin != nil {
					energy = checkin.Energy
				}
				all, _ := taskSvc.List(false)
				if s := task.Suggest(all, energy); s != nil {
					taskID = &s.ID
					taskTitle = s.Title
					_ = taskSvc.SetDoing(s.ID)
				}
			}

			sess, err := focusSvc.Start(taskID)
			if err != nil {
				return err
			}

			duration := time.Duration(minutes) * time.Minute
			deadline := time.Now().Add(duration)

			fmt.Println()
			if taskTitle != "" {
				fmt.Printf("  %s %s\n", styleAccent.Render("focusing:"), styleTask.Bold(true).Render(taskTitle))
			}
			fmt.Printf("  %s %d min — press Ctrl+C to stop early\n\n", styleDim.Render("duration:"), minutes)

			// intercept Ctrl+C for clean exit
			sig := make(chan os.Signal, 1)
			signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()
			interrupted := false

		loop:
			for {
				select {
				case <-sig:
					interrupted = true
					break loop
				case now := <-ticker.C:
					remaining := deadline.Sub(now)
					if remaining <= 0 {
						break loop
					}
					mins := int(remaining.Minutes())
					secs := int(remaining.Seconds()) % 60
					fmt.Printf("\r  %s  %02d:%02d  ", styleAccent2.Render("⏱"), mins, secs)
				}
			}

			_ = focusSvc.End(sess.ID, interrupted)
			elapsed := time.Since(sess.StartedAt).Round(time.Second)

			fmt.Printf("\r  %s  session ended — %s focused\n\n",
				styleGreen.Render("✓"),
				styleAccent.Render(elapsed.String()),
			)
			return nil
		},
	}

	cmd.Flags().IntVar(&minutes, "minutes", 25, "Focus duration in minutes")
	return cmd
}
