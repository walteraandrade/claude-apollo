package style

import "github.com/charmbracelet/lipgloss"

var (
	HeaderBar = lipgloss.NewStyle().
			Background(SlateDeep).
			Foreground(Blue).
			Bold(true).
			Padding(0, 2)

	StatusBar = lipgloss.NewStyle().
			Background(SlateDeep).
			Foreground(WhiteMuted).
			Padding(0, 2)

	Selected = lipgloss.NewStyle().
			Background(Slate).
			Foreground(White).
			Bold(true)

	Normal = lipgloss.NewStyle().
		Foreground(White)

	Muted = lipgloss.NewStyle().
		Foreground(WhiteMuted)

	DetailLabel = lipgloss.NewStyle().
			Foreground(Blue).
			Bold(true)

	DetailValue = lipgloss.NewStyle().
			Foreground(White)

	Border = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(BlueDim)

	Error = lipgloss.NewStyle().
		Foreground(ErrColor)

	HelpKey = lipgloss.NewStyle().
		Foreground(BlueBright).
		Bold(true)

	HelpDesc = lipgloss.NewStyle().
		Foreground(WhiteMuted)

	CardBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(BlueDim).
			Background(SlateDark).
			Padding(0, 1)

	CardSelected = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(BlueBright).
			Background(Slate).
			Padding(0, 1)

	CardExpanded = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Blue).
			Background(Slate).
			Padding(0, 1)

	ColHeaderActive = lipgloss.NewStyle().
			Foreground(Blue).
			Bold(true)

	ColHeaderInactive = lipgloss.NewStyle().
				Foreground(WhiteMuted)

	CardHash = lipgloss.NewStyle().
			Foreground(BlueBright).
			Bold(true)

	CardSubject = lipgloss.NewStyle().
			Foreground(White)

	CardMeta = lipgloss.NewStyle().
			Foreground(WhiteMuted)

	ColDivStyle = lipgloss.NewStyle().
			Foreground(BlueDim)
)

func StatusIcon(status string) string {
	switch status {
	case "unreviewed":
		return lipgloss.NewStyle().Foreground(StatusNeedsReview).Render("●")
	case "reviewed":
		return lipgloss.NewStyle().Foreground(StatusReviewed).Render("✓")
	case "ignored":
		return lipgloss.NewStyle().Foreground(StatusIgnored).Render("○")
	default:
		return " "
	}
}

func FilterLabel(filter string) string {
	switch filter {
	case "unreviewed":
		return lipgloss.NewStyle().Foreground(StatusNeedsReview).Render("UNREVIEWED")
	case "reviewed":
		return lipgloss.NewStyle().Foreground(StatusReviewed).Render("REVIEWED")
	case "ignored":
		return lipgloss.NewStyle().Foreground(StatusIgnored).Render("IGNORED")
	default:
		return Normal.Render("ALL")
	}
}
