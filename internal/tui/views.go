package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/walter/apollo/internal/style"
)

func (m Model) listView() string {
	var b strings.Builder

	header := fmt.Sprintf("  %-3s %-7s %-12s %-15s %s",
		"", "HASH", "BRANCH", "AUTHOR", "SUBJECT")
	b.WriteString(style.Muted.Render(header))
	b.WriteString("\n")

	if len(m.commits) == 0 {
		b.WriteString(style.Muted.Render("  No commits"))
		return b.String()
	}

	visibleStart, visibleEnd := m.visibleRange()
	for idx := visibleStart; idx < visibleEnd; idx++ {
		c := m.commits[idx]
		icon := style.StatusIcon(c.Status)
		hash := c.Hash[:min(7, len(c.Hash))]
		branch := truncate(c.Branch, 12)
		author := truncate(c.Author, 15)
		subject := truncate(c.Subject, m.width-48)

		line := fmt.Sprintf("  %s %-7s %-12s %-15s %s",
			icon, hash, branch, author, subject)

		if idx == m.cursor {
			b.WriteString(style.Selected.Render(line))
		} else {
			b.WriteString(style.Normal.Render(line))
		}
		if idx < visibleEnd-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (m Model) detailView() string {
	if m.cursor >= len(m.commits) {
		return ""
	}
	c := m.commits[m.cursor]

	var b strings.Builder
	field := func(label, value string) {
		b.WriteString(style.DetailLabel.Render(label+": ") + style.DetailValue.Render(value) + "\n")
	}

	b.WriteString("\n")
	field("Hash", c.Hash)
	field("Author", c.Author)
	field("Branch", c.Branch)
	field("Date", c.CommittedAt.Format(time.RFC1123))
	field("Status", c.Status)
	if c.Note != "" {
		field("Note", c.Note)
	}
	b.WriteString("\n")
	b.WriteString(style.DetailLabel.Render("Subject: ") + c.Subject + "\n")
	if c.Body != "" {
		b.WriteString("\n" + c.Body + "\n")
	}

	b.WriteString("\n" + style.Muted.Render("esc: back  r: reviewed  u: unreviewed  i: ignored  n: note"))

	return b.String()
}

func (m Model) statusBar() string {
	filterStr := style.FilterLabel(string(m.filter))
	statsStr := fmt.Sprintf("%d unreviewed / %d total", m.stats.Unreviewed, m.stats.Total)

	watcherStatus := style.Muted.Render("watching")
	if m.watchCh == nil {
		watcherStatus = style.Error.Render("not watching")
	}

	left := fmt.Sprintf(" %s  %s  %s", filterStr, statsStr, watcherStatus)
	return style.StatusBar.Render(lipgloss.PlaceHorizontal(m.width, lipgloss.Left, left))
}

func (m Model) helpBar() string {
	keys := []struct{ key, desc string }{
		{"j/k", "navigate"},
		{"r", "reviewed"},
		{"u", "unreviewed"},
		{"i", "ignored"},
		{"enter", "detail"},
		{"tab", "filter"},
		{"q", "quit"},
	}

	var parts []string
	for _, k := range keys {
		parts = append(parts, style.HelpKey.Render(k.key)+" "+style.HelpDesc.Render(k.desc))
	}
	return style.Muted.Render(" " + strings.Join(parts, "  "))
}

func (m Model) noteInputView() string {
	c := m.commits[m.cursor]
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(style.DetailLabel.Render("Note for ") + c.Hash[:7] + "\n\n")
	b.WriteString(m.noteInput.View())
	b.WriteString("\n\n")
	b.WriteString(style.Muted.Render("enter: save  esc: cancel"))
	return b.String()
}

func (m Model) visibleRange() (int, int) {
	listHeight := m.height - 5
	if listHeight < 1 {
		listHeight = 1
	}

	start := 0
	if m.cursor >= listHeight {
		start = m.cursor - listHeight + 1
	}
	end := start + listHeight
	if end > len(m.commits) {
		end = len(m.commits)
	}
	return start, end
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

func (m Model) errorView() string {
	if m.err == nil {
		return ""
	}
	return style.Error.Render(" Error: " + m.err.Error())
}
