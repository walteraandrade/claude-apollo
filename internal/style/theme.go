package style

import "github.com/charmbracelet/lipgloss"

var (
	Blue       = lipgloss.Color("#5BA4CF")
	BlueBright = lipgloss.Color("#8EC8E8")
	BlueDim    = lipgloss.Color("#3A7CA5")
	BlueMuted  = lipgloss.Color("#2C5F7C")

	White      = lipgloss.Color("#F0F4F8")
	WhiteMuted = lipgloss.Color("#B0BEC5")
	Silver     = lipgloss.Color("#CFD8DC")

	Slate     = lipgloss.Color("#1B2838")
	SlateDark = lipgloss.Color("#141E2B")
	SlateDeep = lipgloss.Color("#0F1923")

	StatusNeedsReview = Blue
	StatusReviewed    = lipgloss.Color("#4DB6AC")
	StatusIgnored     = lipgloss.Color("#78909C")
	ErrColor          = lipgloss.Color("#EF5350")
)

const (
	BorderVert = "│"
	BorderHoriz = "─"
)
