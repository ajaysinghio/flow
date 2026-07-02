package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/ajaykumarsingh/flow/internal/app"
	"github.com/ajaykumarsingh/flow/internal/store"
	"github.com/ajaykumarsingh/flow/internal/task"
	"github.com/spf13/cobra"
)

func NewRoot(a *app.App) *cobra.Command {
	root := &cobra.Command{
		Use:   "flow",
		Short: "One task at a time. For brains that need it.",
		Long:  `flow — a neurodivergent-aware task + mood CLI. Run 'flow' to get your next task.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNow(a)
		},
	}

	root.AddCommand(
		newAddCmd(a),
		newInCmd(a),
		newDoneCmd(a),
		newLsCmd(a),
		newBreakCmd(a),
		newNoteCmd(a),
		newFocusCmd(a),
		newInsightsCmd(a),
		newMCPCmd(a),
		newServeCmd(a),
		newTrayCmd(a),
	)
	return root
}

func runNow(a *app.App) error {
	tasks := a.Tasks
	moods := a.Moods
	checkin, _ := moods.Latest(4 * time.Hour)

	energyLevel := 3 // default medium
	if checkin != nil {
		energyLevel = checkin.Energy
	}

	all, err := tasks.List(false)
	if err != nil {
		return err
	}

	suggested := task.Suggest(all, energyLevel)

	if checkin == nil {
		fmt.Println(styleDim.Render("  No recent check-in found. Assuming energy 3/5."))
		fmt.Println(styleDim.Render("  Run 'flow in' to log your current state.\n"))
	}

	if suggested == nil {
		fmt.Println(styleGreen.Render("  ✓ Nothing left in your queue."))
		fmt.Println(styleDim.Render("  Add something with: flow add \"task title\""))
		return nil
	}

	fmt.Println()
	fmt.Println(stylePrompt.Render("  → right now:"))
	fmt.Println()
	fmt.Printf("  %s %s\n",
		styleTask.Bold(true).Render(suggested.Title),
		styleDim.Render("["+string(suggested.Size)+"]"),
	)
	fmt.Printf("  %s  %s\n",
		styleDim.Render("id:"),
		styleDim.Render(suggested.ID[:8]+"…"),
	)
	fmt.Printf("  %s  %s\n\n",
		styleDim.Render("energy:"),
		energyColor(string(suggested.Energy)).Render(string(suggested.Energy)),
	)
	fmt.Println(styleDim.Render("  flow done        — mark it complete"))
	fmt.Println(styleDim.Render("  flow focus        — start a focus timer"))
	fmt.Println(styleDim.Render("  flow break <id>   — break it into steps"))
	fmt.Println()
	return nil
}

func Execute() {
	db, err := store.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	root := NewRoot(app.New(db))
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
