package cli

import "github.com/charmbracelet/lipgloss"

var (
	styleAccent  = lipgloss.NewStyle().Foreground(lipgloss.Color("#F5A327")).Bold(true)
	styleAccent2 = lipgloss.NewStyle().Foreground(lipgloss.Color("#6C63FF"))
	styleGreen   = lipgloss.NewStyle().Foreground(lipgloss.Color("#3ECFA4"))
	styleDim     = lipgloss.NewStyle().Foreground(lipgloss.Color("#7A8499"))
	styleBold    = lipgloss.NewStyle().Bold(true)
	styleTask    = lipgloss.NewStyle().Foreground(lipgloss.Color("#E4DFD0"))

	styleBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#1E2A38")).
			Padding(0, 1)

	stylePrompt = lipgloss.NewStyle().Foreground(lipgloss.Color("#F5A327")).Bold(true)
)

func energyColor(e string) lipgloss.Style {
	switch e {
	case "low":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#3ECFA4"))
	case "high":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#E05252"))
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#F5A327"))
	}
}

func sizeLabel(s string) string {
	switch s {
	case "xs":
		return "·"
	case "s":
		return "··"
	case "m":
		return "···"
	case "l":
		return "····"
	case "xl":
		return "·····"
	}
	return "···"
}
