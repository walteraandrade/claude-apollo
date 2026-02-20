package watcher

import (
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Event struct {
	RepoPath string
}

func Watch(repoPath string, debounce time.Duration) (<-chan Event, func(), error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, nil, err
	}

	refsPath := filepath.Join(repoPath, ".git", "refs")
	headPath := filepath.Join(repoPath, ".git")

	if err := w.Add(refsPath); err != nil {
		w.Close()
		return nil, nil, err
	}
	headsPath := filepath.Join(refsPath, "heads")
	_ = w.Add(headsPath)
	_ = w.Add(headPath)

	ch := make(chan Event, 1)
	done := make(chan struct{})

	go func() {
		defer close(ch)
		var timer *time.Timer

		for {
			select {
			case <-done:
				if timer != nil {
					timer.Stop()
				}
				return
			case ev, ok := <-w.Events:
				if !ok {
					return
				}
				if ev.Op&(fsnotify.Write|fsnotify.Create) == 0 {
					continue
				}
				if timer != nil {
					timer.Stop()
				}
				timer = time.AfterFunc(debounce, func() {
					select {
					case ch <- Event{RepoPath: repoPath}:
					default:
					}
				})
			case _, ok := <-w.Errors:
				if !ok {
					return
				}
			}
		}
	}()

	cleanup := func() {
		close(done)
		w.Close()
	}

	return ch, cleanup, nil
}
