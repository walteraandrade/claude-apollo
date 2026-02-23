package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/aymanbagabas/go-osc52/v2"
	"github.com/walter/apollo/internal/config"
	"github.com/walter/apollo/internal/db"
	"github.com/walter/apollo/internal/git"
	"github.com/walter/apollo/internal/watcher"
)

func (m Model) initRepo() tea.Cmd {
	return func() tea.Msg {
		path := config.ExpandHome(m.cfg.RepoPath)
		repo, err := git.OpenRepo(path)
		if err != nil {
			return ErrorMsg{Err: fmt.Errorf("open repo %q: %w", path, err)}
		}

		repoID, err := db.UpsertRepo(m.database, repoName(path), path)
		if err != nil {
			return ErrorMsg{Err: fmt.Errorf("upsert repo: %w", err)}
		}

		return RepoInitializedMsg{RepoID: repoID, Repo: repo}
	}
}

func (m Model) seedCommits() tea.Cmd {
	return func() tea.Msg {
		if m.repo == nil {
			return SeedDoneMsg{}
		}

		path := config.ExpandHome(m.cfg.RepoPath)
		r, err := db.GetRepoByPath(m.database, path)
		if err != nil || r == nil {
			return SeedDoneMsg{}
		}

		var commits []git.CommitInfo
		if r.LastCommitHash == "" {
			commits, err = m.repo.SeedCommits(m.cfg.SeedDepth)
		} else {
			commits, err = m.repo.ReadNewCommits(r.LastCommitHash, m.cfg.SeedDepth)
		}
		if err != nil || len(commits) == 0 {
			return SeedDoneMsg{}
		}

		return SeedDoneMsg{Commits: commits}
	}
}

func (m Model) persistCommits(commits []git.CommitInfo) tea.Cmd {
	return func() tea.Msg {
		for _, c := range commits {
			if err := db.InsertCommit(m.database, m.repoID, c.Hash, c.Author, c.Subject, c.Body, c.Branch, c.Timestamp); err != nil {
				return ErrorMsg{Err: fmt.Errorf("insert commit %s: %w", c.Hash[:7], err)}
			}
		}

		last := commits[len(commits)-1]
		if err := db.UpdateLastCommitHash(m.database, m.repoID, last.Hash); err != nil {
			return ErrorMsg{Err: fmt.Errorf("update last hash: %w", err)}
		}

		if m.notifier != nil {
			for _, c := range commits {
				m.notifier.Notify("New commit", c.Subject)
			}
		}

		return CommitsPersistedMsg{Commits: commits}
	}
}

func (m Model) loadCommits() tea.Cmd {
	return func() tea.Msg {
		commits, err := db.ListCommits(m.database, m.repoID, db.FilterAll)
		if err != nil {
			return ErrorMsg{Err: err}
		}
		stats, err := db.GetStats(m.database, m.repoID)
		if err != nil {
			return ErrorMsg{Err: err}
		}
		return CommitsLoadedMsg{Commits: commits, Stats: stats}
	}
}

func (m Model) startWatcher() tea.Cmd {
	return func() tea.Msg {
		path := config.ExpandHome(m.cfg.RepoPath)
		ch, stop, err := watcher.Watch(path, time.Duration(m.cfg.DebounceMs)*time.Millisecond)
		if err != nil {
			return ErrorMsg{Err: fmt.Errorf("watch %q: %w", path, err)}
		}
		return WatcherReadyMsg{Ch: ch, Stop: stop}
	}
}

func (m Model) listenWatcher() tea.Cmd {
	if m.watchCh == nil {
		return nil
	}
	ch := m.watchCh
	return func() tea.Msg {
		_, ok := <-ch
		if !ok {
			return nil
		}
		return WatcherEventMsg{}
	}
}

func (m Model) readNewCommits() tea.Cmd {
	return func() tea.Msg {
		if m.repo == nil {
			return NewCommitsMsg{}
		}
		path := config.ExpandHome(m.cfg.RepoPath)
		r, err := db.GetRepoByPath(m.database, path)
		if err != nil || r == nil {
			return NewCommitsMsg{}
		}

		commits, err := m.repo.ReadNewCommits(r.LastCommitHash, m.cfg.SeedDepth)
		if err != nil {
			return ErrorMsg{Err: err}
		}
		return NewCommitsMsg{Commits: commits}
	}
}

func (m Model) updateReview(hash, status string) tea.Cmd {
	return func() tea.Msg {
		if err := db.UpdateReviewStatus(m.database, hash, status, ""); err != nil {
			return ErrorMsg{Err: err}
		}
		return ReviewUpdatedMsg{Hash: hash, Status: status}
	}
}

func (m Model) copyHashCmd(hash string) tea.Cmd {
	return func() tea.Msg {
		fmt.Print(osc52.New(hash).String())
		return CopiedMsg{Hash: hash}
	}
}

func repoName(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[i+1:]
		}
	}
	return path
}
