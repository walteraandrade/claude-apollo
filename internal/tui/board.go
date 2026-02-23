package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/walter/apollo/internal/style"
)

func (m Model) boardView() string {
	colW := m.columnWidth()
	bh := m.boardHeight()

	var cols []string
	for i := range NumColumns {
		cols = append(cols, m.renderColumn(ColumnID(i), colW, bh))
	}

	div := style.ColumnDivider(bh + 1)

	row := lipgloss.JoinHorizontal(lipgloss.Top,
		cols[0], div, cols[1], div, cols[2],
	)
	return row
}

func (m Model) renderColumn(colID ColumnID, width, height int) string {
	col := m.columns[colID]
	active := colID == m.activeCol

	header := style.ColumnHeader(col.Title, len(col.Commits), active, width)

	if len(col.Commits) == 0 {
		empty := style.Muted.Render("  No commits")
		return lipgloss.JoinVertical(lipgloss.Left, header, "", empty)
	}

	cardW := width - 4
	if cardW < 10 {
		cardW = 10
	}

	start, end := m.visibleCardRange(colID, height)
	var cards []string
	for idx := start; idx < end; idx++ {
		c := col.Commits[idx]
		selected := active && idx == col.Cursor

		if m.expandedHash == c.Hash {
			cards = append(cards, m.renderExpandedCard(c, cardW))
		} else {
			cards = append(cards, m.renderCard(c, cardW, selected))
		}
	}

	body := lipgloss.JoinVertical(lipgloss.Left, cards...)
	return lipgloss.JoinVertical(lipgloss.Left, header, "", body)
}

func (m Model) visibleCardRange(colID ColumnID, height int) (int, int) {
	col := m.columns[colID]
	if len(col.Commits) == 0 {
		return 0, 0
	}

	cardHeight := 5
	expandedExtra := 8
	availableHeight := height - 2

	visibleSlots := availableHeight / cardHeight
	if visibleSlots < 1 {
		visibleSlots = 1
	}

	if m.expandedHash != "" && colID == m.activeCol {
		expandedSlots := (cardHeight + expandedExtra) / cardHeight
		if visibleSlots > expandedSlots {
			visibleSlots -= expandedSlots - 1
		} else {
			visibleSlots = 1
		}
	}

	start := col.Scroll
	if col.Cursor >= start+visibleSlots {
		start = col.Cursor - visibleSlots + 1
	}
	if col.Cursor < start {
		start = col.Cursor
	}
	if start < 0 {
		start = 0
	}

	end := start + visibleSlots
	if end > len(col.Commits) {
		end = len(col.Commits)
	}

	return start, end
}

func (m Model) columnWidth() int {
	dividers := 2
	usable := m.width - dividers
	if usable < 30 {
		usable = 30
	}
	return usable / 3
}

func (m Model) boardHeight() int {
	h := m.height - 5
	if h < 5 {
		h = 5
	}
	return h
}

func truncate(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-1] + "…"
}

