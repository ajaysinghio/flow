package cli

import (
	"fmt"
	"time"

	"github.com/ajaykumarsingh/flow/internal/mood"
	"github.com/ajaykumarsingh/flow/internal/store"
	"github.com/ajaykumarsingh/flow/internal/task"
	"github.com/spf13/cobra"
)

func newInsightsCmd(db *store.DB) *cobra.Command {
	var days int

	cmd := &cobra.Command{
		Use:   "insights",
		Short: "Mood trends and task patterns",
		RunE: func(cmd *cobra.Command, args []string) error {
			moodSvc := mood.NewService(db)
			taskSvc := task.NewService(db)

			checkins, err := moodSvc.Recent(days * 3)
			if err != nil {
				return err
			}
			tasks, err := taskSvc.List(true)
			if err != nil {
				return err
			}

			fmt.Println()
			fmt.Println(styleAccent.Render("  insights") + styleDim.Render(fmt.Sprintf("  (last %d days)", days)))
			fmt.Println()

			// Mood average
			if len(checkins) > 0 {
				var moodSum, energySum int
				for _, c := range checkins {
					moodSum += c.Mood
					energySum += c.Energy
				}
				avgMood := float64(moodSum) / float64(len(checkins))
				avgEnergy := float64(energySum) / float64(len(checkins))
				fmt.Printf("  %s  mood %.1f/5   energy %.1f/5   (%d check-ins)\n",
					styleDim.Render("avg"),
					avgMood, avgEnergy, len(checkins),
				)
			} else {
				fmt.Println(styleDim.Render("  No check-ins yet. Run 'flow in' to start."))
			}

			// Tasks
			cutoff := time.Now().AddDate(0, 0, -days)
			var done, total int
			for _, t := range tasks {
				if t.CreatedAt.After(cutoff) {
					total++
					if t.Status == task.StatusDone {
						done++
					}
				}
			}
			if total > 0 {
				pct := float64(done) / float64(total) * 100
				fmt.Printf("  %s  %d/%d tasks completed (%.0f%%)\n",
					styleDim.Render("tasks"),
					done, total, pct,
				)
			}

			// Mini mood chart (last 7 days)
			fmt.Println()
			fmt.Println(styleDim.Render("  mood  last 7 days"))
			fmt.Println()
			for i := 6; i >= 0; i-- {
				d := time.Now().AddDate(0, 0, -i)
				dayLabel := d.Format("Mon")
				var dayMoods []int
				for _, c := range checkins {
					if c.Timestamp.Format("2006-01-02") == d.Format("2006-01-02") {
						dayMoods = append(dayMoods, c.Mood)
					}
				}
				bar := "  " + styleDim.Render(dayLabel) + "  "
				if len(dayMoods) == 0 {
					bar += styleDim.Render("·")
				} else {
					avg := 0
					for _, m := range dayMoods {
						avg += m
					}
					avg /= len(dayMoods)
					blocks := []string{"▁", "▂", "▃", "▄", "█"}
					block := blocks[avg-1]
					bar += styleAccent.Render(block) + styleDim.Render(fmt.Sprintf(" %d", avg))
				}
				fmt.Println(bar)
			}
			fmt.Println()
			return nil
		},
	}
	cmd.Flags().IntVar(&days, "days", 7, "Number of days to include")
	return cmd
}
