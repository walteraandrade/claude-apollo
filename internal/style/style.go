package style

import "github.com/charmbracelet/lipgloss"

var (
	Unreviewed = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	Reviewed   = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	Ignored    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	HeaderBar = lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("220")).
			Bold(true).
			Padding(0, 2)

	StatusBar = lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("252")).
			Padding(0, 2)

	Selected = lipgloss.NewStyle().
			Background(lipgloss.Color("237")).
			Foreground(lipgloss.Color("255")).
			Bold(true)

	Normal = lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	Muted = lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	DetailLabel = lipgloss.NewStyle().
			Foreground(lipgloss.Color("220")).
			Bold(true)

	DetailValue = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	Border = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240"))

	Error = lipgloss.NewStyle().
		Foreground(lipgloss.Color("196"))

	HelpKey = lipgloss.NewStyle().
		Foreground(lipgloss.Color("220")).
		Bold(true)

	HelpDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))
)

func StatusIcon(status string) string {
	switch status {
	case "unreviewed":
		return Unreviewed.Render("●")
	case "reviewed":
		return Reviewed.Render("✓")
	case "ignored":
		return Ignored.Render("○")
	default:
		return " "
	}
}

func FilterLabel(filter string) string {
	switch filter {
	case "unreviewed":
		return Unreviewed.Render("UNREVIEWED")
	case "reviewed":
		return Reviewed.Render("REVIEWED")
	case "ignored":
		return Ignored.Render("IGNORED")
	default:
		return Normal.Render("ALL")
	}
}
