package tui

import (
	"fmt"
	"time"

	"github.com/walter/apollo/internal/db"
	"github.com/walter/apollo/internal/style"
)

func (m Model) renderCard(c db.CommitRow, width int, selected bool) string {
	hash := style.CardHash.Render(c.Hash[:min(7, len(c.Hash))])
	icon := style.StatusIcon(c.Status)
	subject := style.CardSubject.Render(truncate(c.Subject, width-2))

	metaParts := truncate(c.Author, 12) + " · " + truncate(c.Branch, 12)
	if len(m.handles) > 1 && c.RepoName != "" {
		metaParts += " · " + truncate(c.RepoName, 14)
	}
	meta := style.CardMeta.Render(metaParts)

	content := fmt.Sprintf("%s %s\n%s\n%s", icon, hash, subject, meta)
	return style.Card(content, width, selected)
}

func (m Model) renderExpandedCard(c db.CommitRow, width int) string {
	hash := style.CardHash.Render(c.Hash)
	icon := style.StatusIcon(c.Status)
	subject := style.CardSubject.Render(c.Subject)
	author := style.DetailLabel.Render("Author: ") + style.DetailValue.Render(c.Author)
	branch := style.DetailLabel.Render("Branch: ") + style.DetailValue.Render(c.Branch)
	date := style.DetailLabel.Render("Date:   ") + style.DetailValue.Render(c.CommittedAt.Format(time.RFC1123))
	status := style.DetailLabel.Render("Status: ") + style.StatusBadge(c.Status) + " " + style.DetailValue.Render(c.Status)

	content := fmt.Sprintf("%s %s\n%s\n\n%s\n%s\n%s\n%s", icon, hash, subject, author, branch, date, status)

	if len(m.handles) > 1 && c.RepoName != "" {
		repo := style.DetailLabel.Render("Repo:   ") + style.DetailValue.Render(c.RepoName)
		content += "\n" + repo
	}

	if c.Body != "" {
		body := style.CardMeta.Render(truncate(c.Body, width*3))
		content += "\n\n" + body
	}
	if c.Note != "" {
		note := style.DetailLabel.Render("Note: ") + style.DetailValue.Render(c.Note)
		content += "\n" + note
	}

	return style.ExpandedCard(content, width)
}
