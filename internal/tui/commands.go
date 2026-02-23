package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/aymanbagabas/go-osc52/v2"
	"github.com/walter/apollo/internal/db"
	"github.com/walter/apollo/internal/git"
	"github.com/walter/apollo/internal/watcher"
)

func (m Model) initRepos() tea.Cmd {
	return func() tea.Msg {
		paths := m.cfg.ResolvedPaths()
		handles := make([]RepoHandle, len(paths))

		for i, path := range paths {
			h := RepoHandle{Path: path, Name: repoName(path)}

			repo, err := git.OpenRepo(path)
			if err != nil {
				h.Err = fmt.Errorf("open repo %q: %w", path, err)
				handles[i] = h
				continue
			}
			h.Repo = repo

			repoID, err := db.UpsertRepo(m.database, h.Name, path)
			if err != nil {
				h.Err = fmt.Errorf("upsert repo %q: %w", path, err)
				handles[i] = h
				continue
			}
			h.RepoID = repoID
			handles[i] = h
		}

		return ReposInitializedMsg{Handles: handles}
	}
}

func (m Model) seedAllCommits() tea.Cmd {
	return func() tea.Msg {
		var results []RepoSeedResult
		for _, h := range m.handles {
			if h.Err != nil || h.Repo == nil {
				continue
			}

			r, err := db.GetRepoByPath(m.database, h.Path)
			if err != nil || r == nil {
				continue
			}

			var commits []git.CommitInfo
			if r.LastCommitHash == "" {
				commits, err = h.Repo.SeedCommits(m.cfg.SeedDepth)
			} else {
				commits, err = h.Repo.ReadNewCommits(r.LastCommitHash, m.cfg.SeedDepth)
			}
			if err != nil || len(commits) == 0 {
				continue
			}

			results = append(results, RepoSeedResult{
				RepoID:  h.RepoID,
				Path:    h.Path,
				Commits: commits,
			})
		}
		return AllSeedDoneMsg{PerRepo: results}
	}
}

func (m Model) persistAllCommits(results []RepoSeedResult) tea.Cmd {
	return func() tea.Msg {
		for _, res := range results {
			handle := m.handleByPath(res.Path)
			name := ""
			if handle != nil {
				name = handle.Name
			}

			for _, c := range res.Commits {
				if err := db.InsertCommit(m.database, res.RepoID, c.Hash, c.Author, c.Subject, c.Body, c.Branch, c.Timestamp); err != nil {
					return ErrorMsg{Err: fmt.Errorf("insert commit %s: %w", c.Hash[:7], err)}
				}
			}

			last := res.Commits[len(res.Commits)-1]
			if err := db.UpdateLastCommitHash(m.database, res.RepoID, last.Hash); err != nil {
				return ErrorMsg{Err: fmt.Errorf("update last hash: %w", err)}
			}

			if m.notifier != nil {
				prefix := ""
				if len(m.handles) > 1 && name != "" {
					prefix = "[" + name + "] "
				}
				for _, c := range res.Commits {
					m.notifier.Notify(prefix+"New commit", c.Subject)
				}
			}
		}
		return CommitsPersistedMsg{}
	}
}

func (m Model) loadAllCommits() tea.Cmd {
	return func() tea.Msg {
		ids := m.repoIDs()
		if len(ids) == 0 {
			return CommitsLoadedMsg{}
		}
		commits, err := db.ListAllCommits(m.database, ids, db.FilterAll)
		if err != nil {
			return ErrorMsg{Err: err}
		}
		stats, err := db.GetAggregateStats(m.database, ids)
		if err != nil {
			return ErrorMsg{Err: err}
		}
		return CommitsLoadedMsg{Commits: commits, Stats: stats}
	}
}

func (m Model) startAllWatchers() tea.Cmd {
	return func() tea.Msg {
		mux := watcher.NewMux()
		for i := range m.handles {
			h := &m.handles[i]
			if h.Err != nil || h.Repo == nil {
				continue
			}
			ch, stop, err := watcher.Watch(h.Path, time.Duration(m.cfg.DebounceMs)*time.Millisecond)
			if err != nil {
				h.Err = fmt.Errorf("watch %q: %w", h.Path, err)
				continue
			}
			h.WatchCh = ch
			h.Stop = stop
			mux.Add(ch)
		}
		return WatchersReadyMsg{Mux: mux}
	}
}

func (m Model) listenMux() tea.Cmd {
	if m.mux == nil {
		return nil
	}
	ch := m.mux.Events()
	return func() tea.Msg {
		ev, ok := <-ch
		if !ok {
			return nil
		}
		return WatcherEventMsg{RepoPath: ev.RepoPath}
	}
}

func (m Model) readNewCommitsForRepo(repoPath string) tea.Cmd {
	return func() tea.Msg {
		h := m.handleByPath(repoPath)
		if h == nil || h.Repo == nil {
			return NewCommitsMsg{}
		}
		r, err := db.GetRepoByPath(m.database, h.Path)
		if err != nil || r == nil {
			return NewCommitsMsg{}
		}
		commits, err := h.Repo.ReadNewCommits(r.LastCommitHash, m.cfg.SeedDepth)
		if err != nil {
			return ErrorMsg{Err: err}
		}
		return NewCommitsMsg{RepoID: h.RepoID, Commits: commits}
	}
}

func (m Model) persistCommits(repoID int64, commits []git.CommitInfo) tea.Cmd {
	return func() tea.Msg {
		var handle *RepoHandle
		for i := range m.handles {
			if m.handles[i].RepoID == repoID {
				handle = &m.handles[i]
				break
			}
		}

		for _, c := range commits {
			if err := db.InsertCommit(m.database, repoID, c.Hash, c.Author, c.Subject, c.Body, c.Branch, c.Timestamp); err != nil {
				return ErrorMsg{Err: fmt.Errorf("insert commit %s: %w", c.Hash[:7], err)}
			}
		}

		last := commits[len(commits)-1]
		if err := db.UpdateLastCommitHash(m.database, repoID, last.Hash); err != nil {
			return ErrorMsg{Err: fmt.Errorf("update last hash: %w", err)}
		}

		if m.notifier != nil && handle != nil {
			prefix := ""
			if len(m.handles) > 1 {
				prefix = "[" + handle.Name + "] "
			}
			for _, c := range commits {
				m.notifier.Notify(prefix+"New commit", c.Subject)
			}
		}

		return CommitsPersistedMsg{}
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
