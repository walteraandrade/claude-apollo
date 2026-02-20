package tui

import (
	"github.com/walter/apollo/internal/db"
	"github.com/walter/apollo/internal/git"
	"github.com/walter/apollo/internal/watcher"
)

type RepoInitializedMsg struct {
	RepoID int64
	Repo   *git.Repo
}

type SeedDoneMsg struct {
	Commits []git.CommitInfo
}

type WatcherReadyMsg struct {
	Ch   <-chan watcher.Event
	Stop func()
}

type WatcherEventMsg struct{}

type NewCommitsMsg struct {
	Commits []git.CommitInfo
}

type CommitsPersistedMsg struct {
	Commits []git.CommitInfo
}

type CommitsLoadedMsg struct {
	Commits []db.CommitRow
	Stats   db.Stats
}

type ReviewUpdatedMsg struct {
	Hash   string
	Status string
}

type ErrorMsg struct {
	Err error
}
