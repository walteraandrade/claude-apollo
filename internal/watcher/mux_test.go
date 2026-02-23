package watcher

import (
	"testing"
	"time"
)

func TestMuxFanIn(t *testing.T) {
	m := NewMux()

	ch1 := make(chan Event, 1)
	ch2 := make(chan Event, 1)
	m.Add(ch1)
	m.Add(ch2)

	ch1 <- Event{RepoPath: "/a"}
	ch2 <- Event{RepoPath: "/b"}

	got := map[string]bool{}
	timeout := time.After(time.Second)
	for i := 0; i < 2; i++ {
		select {
		case ev := <-m.Events():
			got[ev.RepoPath] = true
		case <-timeout:
			t.Fatal("timeout waiting for events")
		}
	}

	if !got["/a"] || !got["/b"] {
		t.Errorf("got %v, want /a and /b", got)
	}

	m.Close()
}

func TestMuxClose(t *testing.T) {
	m := NewMux()
	ch := make(chan Event, 1)
	m.Add(ch)
	m.Close()

	_, ok := <-m.Events()
	if ok {
		t.Error("events channel should be closed after Close()")
	}
}

func TestMuxClosedSource(t *testing.T) {
	m := NewMux()
	ch := make(chan Event)
	m.Add(ch)

	close(ch)
	time.Sleep(10 * time.Millisecond)

	m.Close()

	_, ok := <-m.Events()
	if ok {
		t.Error("events channel should be closed")
	}
}
