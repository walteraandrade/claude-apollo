package style

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func Card(content string, width int, selected bool) string {
	s := CardBorder
	if selected {
		s = CardSelected
	}
	return s.Width(width).Render(content)
}

func ExpandedCard(content string, width int) string {
	return CardExpanded.Width(width).Render(content)
}

func ColumnHeader(name string, count int, active bool, width int) string {
	s := ColHeaderInactive
	if active {
		s = ColHeaderActive
	}
	badge := Muted.Render(fmt.Sprintf("(%d)", count))
	title := s.Render(name) + " " + badge
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, title)
}

func ColumnDivider(height int) string {
	line := strings.Repeat(BorderVert+"\n", height)
	if len(line) > 0 {
		line = line[:len(line)-1]
	}
	return ColDivStyle.Render(line)
}

func StatusBadge(status string) string {
	switch status {
	case "unreviewed":
		return lipgloss.NewStyle().Foreground(StatusNeedsReview).Render("●")
	case "reviewed":
		return lipgloss.NewStyle().Foreground(StatusReviewed).Render("●")
	case "ignored":
		return lipgloss.NewStyle().Foreground(StatusIgnored).Render("●")
	default:
		return " "
	}
}
