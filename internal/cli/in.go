package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ajaykumarsingh/flow/internal/app"
	"github.com/ajaykumarsingh/flow/internal/mood"
	"github.com/spf13/cobra"
)

type checkinModel struct {
	step   int
	mood   int
	energy int
	note   []rune
	svc    *mood.Service
	saved  bool
	err    error
}

func (m checkinModel) Init() tea.Cmd { return nil }

func (m checkinModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.step {
		case 0:
			switch msg.String() {
			case "1", "2", "3", "4", "5":
				m.mood = int(msg.String()[0] - '0')
				m.step = 1
			case "ctrl+c", "q":
				return m, tea.Quit
			}
		case 1:
			switch msg.String() {
			case "1", "2", "3", "4", "5":
				m.energy = int(msg.String()[0] - '0')
				m.step = 2
			case "ctrl+c", "q":
				return m, tea.Quit
			}
		case 2:
			switch msg.String() {
			case "enter":
				_, m.err = m.svc.Save(m.mood, m.energy, string(m.note))
				m.saved = true
				m.step = 3
				return m, tea.Quit
			case "ctrl+c":
				return m, tea.Quit
			case "backspace":
				if len(m.note) > 0 {
					m.note = m.note[:len(m.note)-1]
				}
			default:
				if len(msg.String()) == 1 {
					m.note = append(m.note, rune(msg.String()[0]))
				}
			}
		}
	}
	return m, nil
}

var moodEmojis = []string{"😩", "😔", "😐", "🙂", "😄"}
var moodLabels = []string{"rough", "low", "okay", "good", "great"}
var energyLabels = []string{"drained", "low", "medium", "good", "charged"}

func (m checkinModel) View() string {
	switch m.step {
	case 0:
		s := "\n  " + styleAccent.Render("How are you feeling right now?") + "\n\n"
		for i, e := range moodEmojis {
			n := i + 1
			style := styleDim
			if m.mood == n {
				style = styleAccent
			}
			s += fmt.Sprintf("  %s  %s %s\n", style.Render(fmt.Sprintf("%d", n)), e, style.Render(moodLabels[i]))
		}
		return s + "\n  " + styleDim.Render("press 1–5") + "\n"
	case 1:
		s := "\n  " + styleAccent.Render("Energy level?") + "\n\n"
		bars := []string{"▪", "▪▪", "▪▪▪", "▪▪▪▪", "▪▪▪▪▪"}
		barStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#3ECFA4"))
		for i, bar := range bars {
			n := i + 1
			style := styleDim
			if m.energy == n {
				style = styleAccent
			}
			s += fmt.Sprintf("  %s  %s  %s\n", style.Render(fmt.Sprintf("%d", n)), barStyle.Render(bar), style.Render(energyLabels[i]))
		}
		return s + "\n  " + styleDim.Render("press 1–5") + "\n"
	case 2:
		s := "\n  " + styleAccent.Render("Anything on your mind? (optional, press Enter to save)") + "\n\n"
		return s + "  " + styleTask.Render(string(m.note)) + styleDim.Render("_") + "\n"
	}
	return ""
}

func newInCmd(a *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "in",
		Short: "Log your current mood and energy (30 seconds)",
		RunE: func(cmd *cobra.Command, args []string) error {
			m := checkinModel{svc: a.Moods}
			p := tea.NewProgram(m)
			result, err := p.Run()
			if err != nil {
				return err
			}
			final := result.(checkinModel)
			if !final.saved {
				fmt.Println(styleDim.Render("\n  check-in cancelled"))
				return nil
			}
			if final.err != nil {
				return final.err
			}
			fmt.Printf("\n  %s  mood %s  energy %s\n\n",
				styleGreen.Render("✓ logged"),
				styleAccent.Render(fmt.Sprintf("%d/5", final.mood)),
				styleAccent2.Render(fmt.Sprintf("%d/5", final.energy)),
			)
			return nil
		},
	}
}
