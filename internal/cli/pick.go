package cli

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ajaykumarsingh/flow/internal/app"
	"github.com/ajaykumarsingh/flow/internal/task"
	"github.com/spf13/cobra"
)

// ── picker TUI ────────────────────────────────────────────────────────────────

type pickModel struct {
	tasks    []*task.Task
	cursor   int
	selected *task.Task
	quit     bool
}

func (m pickModel) Init() tea.Cmd { return nil }

func (m pickModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.tasks)-1 {
				m.cursor++
			}
		case "enter", " ":
			m.selected = m.tasks[m.cursor]
			return m, tea.Quit
		case "q", "ctrl+c", "esc":
			m.quit = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m pickModel) View() string {
	if len(m.tasks) == 0 {
		return "\n  " + styleDim.Render("No tasks in your queue.") + "\n"
	}

	s := "\n  " + styleAccent.Render("Pick a task") + styleDim.Render("  ↑↓ navigate  Enter to start  q to cancel") + "\n\n"
	for i, t := range m.tasks {
		cursor := "  "
		title := styleDim.Render(t.Title)
		if i == m.cursor {
			cursor = styleAccent.Render("▶ ")
			title = styleTask.Bold(true).Render(t.Title)
		}

		due := ""
		if d := task.FormatDue(t); d != "" {
			if t.IsOverdue() {
				due = "  " + styleDanger.Render("⚠ "+d)
			} else {
				due = "  " + styleDim.Render(d)
			}
		}

		s += fmt.Sprintf("  %s%s%s\n", cursor, title, due)
	}
	s += "\n"
	return s
}

// ── command ───────────────────────────────────────────────────────────────────

func newPickCmd(a *app.App) *cobra.Command {
	var minutes int

	cmd := &cobra.Command{
		Use:   "pick",
		Short: "Choose a task, focus on it, then mark it done",
		Long:  `Shows your task queue ranked by energy and urgency. Pick one, a focus timer starts, and you're prompted to mark it done when time's up.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// get ranked tasks
			checkin, _ := a.Moods.Latest(4 * time.Hour)
			energy := 3
			if checkin != nil {
				energy = checkin.Energy
			}
			all, err := a.Tasks.List(false)
			if err != nil {
				return err
			}
			ranked := task.Ranked(all, energy)
			if len(ranked) == 0 {
				fmt.Println(styleDim.Render("\n  Nothing in your queue. Add something: flow add \"...\"\n"))
				return nil
			}

			// show picker
			m := pickModel{tasks: ranked}
			p := tea.NewProgram(m)
			result, err := p.Run()
			if err != nil {
				return err
			}
			final := result.(pickModel)
			if final.quit || final.selected == nil {
				fmt.Println(styleDim.Render("\n  cancelled\n"))
				return nil
			}

			chosen := final.selected
			_ = a.Tasks.SetDoing(chosen.ID)

			// start focus timer
			sess, err := a.Focus.Start(&chosen.ID)
			if err != nil {
				return err
			}

			deadline := time.Now().Add(time.Duration(minutes) * time.Minute)
			fmt.Printf("\n  %s %s\n", styleAccent.Render("focusing:"), styleTask.Bold(true).Render(chosen.Title))
			fmt.Printf("  %s %d min — Ctrl+C to stop early\n\n", styleDim.Render("duration:"), minutes)

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
					fmt.Printf("\r  %s  %02d:%02d  ",
						styleAccent2.Render("⏱"),
						int(remaining.Minutes()),
						int(remaining.Seconds())%60,
					)
				}
			}

			_ = a.Focus.End(sess.ID, interrupted)
			elapsed := time.Since(sess.StartedAt).Round(time.Second)
			fmt.Printf("\r  %s  %s focused\n\n", styleGreen.Render("✓ time's up"), styleAccent.Render(elapsed.String()))

			// prompt to mark done
			fmt.Printf("  Mark \"%s\" as done? [y/n] ", chosen.Title)
			var answer string
			fmt.Scanln(&answer)
			if answer == "y" || answer == "Y" || answer == "yes" {
				if err := a.Tasks.Complete(chosen.ID); err != nil {
					return err
				}
				fmt.Printf("\n  %s  %s\n\n", styleGreen.Render("✓ done"), styleTask.Render(chosen.Title))
			} else {
				fmt.Println(styleDim.Render("\n  kept in queue — pick it up again anytime\n"))
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&minutes, "minutes", 25, "Focus duration in minutes")
	return cmd
}
