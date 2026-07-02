package cli

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ajaykumarsingh/flow/internal/app"
	"github.com/ajaykumarsingh/flow/internal/task"
	"github.com/spf13/cobra"
)

func newFocusCmd(a *app.App) *cobra.Command {
	var minutes int
	cmd := &cobra.Command{
		Use:     "focus [task-id]",
		Short:   "Start a timed focus session",
		Example: "  flow focus\n  flow focus 01ARZ3N --minutes 45",
		RunE: func(cmd *cobra.Command, args []string) error {
			var taskID *string
			var taskTitle string

			if len(args) > 0 {
				t, err := a.Tasks.Get(args[0])
				if err != nil {
					t, err = findByPrefix(a.Tasks, args[0])
				}
				if err == nil {
					taskID = &t.ID
					taskTitle = t.Title
					_ = a.Tasks.SetDoing(t.ID)
				}
			} else {
				checkin, _ := a.Moods.Latest(4 * time.Hour)
				energy := 3
				if checkin != nil {
					energy = checkin.Energy
				}
				all, _ := a.Tasks.List(false)
				if s := task.Suggest(all, energy); s != nil {
					taskID = &s.ID
					taskTitle = s.Title
					_ = a.Tasks.SetDoing(s.ID)
				}
			}

			sess, err := a.Focus.Start(taskID)
			if err != nil {
				return err
			}

			deadline := time.Now().Add(time.Duration(minutes) * time.Minute)
			fmt.Println()
			if taskTitle != "" {
				fmt.Printf("  %s %s\n", styleAccent.Render("focusing:"), styleTask.Bold(true).Render(taskTitle))
			}
			fmt.Printf("  %s %d min — press Ctrl+C to stop early\n\n", styleDim.Render("duration:"), minutes)

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
					fmt.Printf("\r  %s  %02d:%02d  ", styleAccent2.Render("⏱"), int(remaining.Minutes()), int(remaining.Seconds())%60)
				}
			}

			_ = a.Focus.End(sess.ID, interrupted)
			fmt.Printf("\r  %s  session ended — %s focused\n\n",
				styleGreen.Render("✓"),
				styleAccent.Render(time.Since(sess.StartedAt).Round(time.Second).String()),
			)
			return nil
		},
	}
	cmd.Flags().IntVar(&minutes, "minutes", 25, "Focus duration in minutes")
	return cmd
}
