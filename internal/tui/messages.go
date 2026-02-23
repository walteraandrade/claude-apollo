package tui

import (
	"github.com/walter/apollo/internal/db"
	"github.com/walter/apollo/internal/git"
	"github.com/walter/apollo/internal/watcher"
)

type ReposInitializedMsg struct {
	Handles []RepoHandle
}

type RepoSeedResult struct {
	RepoID  int64
	Path    string
	Commits []git.CommitInfo
}

type AllSeedDoneMsg struct {
	PerRepo []RepoSeedResult
}

type WatchersReadyMsg struct {
	Mux *watcher.Mux
}

type WatcherEventMsg struct {
	RepoPath string
}

type NewCommitsMsg struct {
	RepoID  int64
	Commits []git.CommitInfo
}

type CommitsPersistedMsg struct{}

type CommitsLoadedMsg struct {
	Commits []db.CommitRow
	Stats   db.Stats
}

type ReviewUpdatedMsg struct {
	Hash   string
	Status string
}

type CopiedMsg struct {
	Hash string
}

type CopyFlashTickMsg struct{}

type ErrorMsg struct {
	Err error
}
