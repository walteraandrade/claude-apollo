package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/walter/apollo/internal/style"
)

func (m Model) statusBar() string {
	var left string
	if m.copiedHash != "" {
		left = fmt.Sprintf(" Copied %s", m.copiedHash)
	} else {
		left = fmt.Sprintf(" %d needs review · %d reviewed · %d ignored",
			m.stats.Unreviewed, m.stats.Reviewed, m.stats.Ignored)
	}

	watcherStatus := m.watcherStatusText()
	left = left + "  " + watcherStatus
	return style.StatusBar.Render(lipgloss.PlaceHorizontal(m.width, lipgloss.Left, left))
}

func (m Model) watcherStatusText() string {
	if len(m.handles) == 0 {
		return style.Error.Render("no repos")
	}

	total := 0
	active := 0
	errored := 0
	for _, h := range m.handles {
		total++
		if h.Err != nil {
			errored++
		} else if h.WatchCh != nil {
			active++
		}
	}

	if total == 1 {
		if active == 1 {
			return style.Muted.Render("watching")
		}
		return style.Error.Render("not watching")
	}

	if errored > 0 {
		return style.Error.Render(fmt.Sprintf("watching %d/%d repos (%d errors)", active, total, errored))
	}
	return style.Muted.Render(fmt.Sprintf("watching %d repos", active))
}

func (m Model) helpBar() string {
	keys := []struct{ key, desc string }{
		{"h/l", "columns"},
		{"j/k", "cards"},
		{"enter", "expand"},
		{"c", "copy"},
		{"r", "reviewed"},
		{"u", "unreviewed"},
		{"i", "ignored"},
		{"n", "note"},
		{"q", "quit"},
	}

	var parts []string
	for _, k := range keys {
		parts = append(parts, style.HelpKey.Render(k.key)+" "+style.HelpDesc.Render(k.desc))
	}
	return style.Muted.Render(" " + strings.Join(parts, "  "))
}

func (m Model) noteInputView() string {
	c := m.selectedCommit()
	if c == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(style.DetailLabel.Render("Note for ") + c.Hash[:7] + "\n\n")
	b.WriteString(m.noteInput.View())
	b.WriteString("\n\n")
	b.WriteString(style.Muted.Render("enter: save  esc: cancel"))
	return b.String()
}

func (m Model) errorView() string {
	if m.err == nil {
		return ""
	}
	return style.Error.Render(" Error: " + m.err.Error())
}
