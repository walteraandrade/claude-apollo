package watcher

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func setupWatcherRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %s %v", out, err)
	}
	return dir
}

func TestWatchDetectsRefChange(t *testing.T) {
	dir := setupWatcherRepo(t)
	ch, cleanup, err := Watch(dir, 50*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	refsHeads := filepath.Join(dir, ".git", "refs", "heads")
	os.MkdirAll(refsHeads, 0755)
	os.WriteFile(filepath.Join(refsHeads, "main"), []byte("abc123\n"), 0644)

	select {
	case ev := <-ch:
		if ev.RepoPath != dir {
			t.Errorf("path = %q, want %q", ev.RepoPath, dir)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestWatchCleanup(t *testing.T) {
	dir := setupWatcherRepo(t)
	_, cleanup, err := Watch(dir, 50*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	cleanup()
}
