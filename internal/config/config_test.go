package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaults(t *testing.T) {
	cfg := Defaults()
	if cfg.SeedDepth != 50 {
		t.Errorf("SeedDepth = %d, want 50", cfg.SeedDepth)
	}
	if cfg.DebounceMs != 300 {
		t.Errorf("DebounceMs = %d, want 300", cfg.DebounceMs)
	}
}

func TestLoadMissingFile(t *testing.T) {
	orig := os.Getenv("HOME")
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	defer t.Setenv("HOME", orig)

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.SeedDepth != 50 {
		t.Errorf("SeedDepth = %d, want 50", cfg.SeedDepth)
	}
}

func TestSaveAndLoad(t *testing.T) {
	orig := os.Getenv("HOME")
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	defer t.Setenv("HOME", orig)

	want := Config{
		RepoPath:   "/tmp/myrepo",
		SeedDepth:  100,
		DebounceMs: 500,
	}
	if err := Save(want); err != nil {
		t.Fatal(err)
	}

	got, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if got.RepoPath != want.RepoPath || got.SeedDepth != want.SeedDepth || got.DebounceMs != want.DebounceMs {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestEnvOverrides(t *testing.T) {
	orig := os.Getenv("HOME")
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	defer t.Setenv("HOME", orig)

	t.Setenv("APOLLO_REPO_PATH", "/env/repo")
	t.Setenv("APOLLO_SEED_DEPTH", "25")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.RepoPath != "/env/repo" {
		t.Errorf("RepoPath = %q, want /env/repo", cfg.RepoPath)
	}
	if cfg.SeedDepth != 25 {
		t.Errorf("SeedDepth = %d, want 25", cfg.SeedDepth)
	}
}

func TestExpandHome(t *testing.T) {
	home, _ := os.UserHomeDir()
	got := ExpandHome("~/foo")
	want := filepath.Join(home, "foo")
	if got != want {
		t.Errorf("ExpandHome(~/foo) = %q, want %q", got, want)
	}

	if ExpandHome("/abs/path") != "/abs/path" {
		t.Error("absolute path should be unchanged")
	}
	if ExpandHome("") != "" {
		t.Error("empty path should be unchanged")
	}
}
