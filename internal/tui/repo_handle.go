package tui

import (
	"github.com/walter/apollo/internal/git"
	"github.com/walter/apollo/internal/watcher"
)

type RepoHandle struct {
	Path    string
	Name    string
	RepoID  int64
	Repo    *git.Repo
	WatchCh <-chan watcher.Event
	Stop    func()
	Err     error
}
