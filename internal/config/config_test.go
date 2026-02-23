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

func TestResolvedPathsSingleRepo(t *testing.T) {
	cfg := Config{RepoPath: "/a"}
	got := cfg.ResolvedPaths()
	if len(got) != 1 || got[0] != "/a" {
		t.Errorf("got %v, want [/a]", got)
	}
}

func TestResolvedPathsMerge(t *testing.T) {
	cfg := Config{RepoPath: "/a", RepoPaths: []string{"/b", "/c"}}
	got := cfg.ResolvedPaths()
	if len(got) != 3 {
		t.Fatalf("got %v, want 3 paths", got)
	}
	if got[0] != "/a" || got[1] != "/b" || got[2] != "/c" {
		t.Errorf("got %v", got)
	}
}

func TestResolvedPathsDedup(t *testing.T) {
	cfg := Config{RepoPath: "/a", RepoPaths: []string{"/a", "/b"}}
	got := cfg.ResolvedPaths()
	if len(got) != 2 {
		t.Fatalf("got %v, want [/a /b]", got)
	}
}

func TestResolvedPathsEmpty(t *testing.T) {
	cfg := Config{}
	got := cfg.ResolvedPaths()
	if len(got) != 0 {
		t.Errorf("got %v, want empty", got)
	}
}

func TestResolvedPathsExpandsHome(t *testing.T) {
	home, _ := os.UserHomeDir()
	cfg := Config{RepoPaths: []string{"~/foo"}}
	got := cfg.ResolvedPaths()
	want := filepath.Join(home, "foo")
	if len(got) != 1 || got[0] != want {
		t.Errorf("got %v, want [%s]", got, want)
	}
}

func TestEnvOverrideRepoPaths(t *testing.T) {
	orig := os.Getenv("HOME")
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	defer t.Setenv("HOME", orig)

	t.Setenv("APOLLO_REPO_PATHS", "/x,/y,/z")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.RepoPaths) != 3 {
		t.Fatalf("RepoPaths = %v, want 3 items", cfg.RepoPaths)
	}
	if cfg.RepoPaths[0] != "/x" || cfg.RepoPaths[1] != "/y" || cfg.RepoPaths[2] != "/z" {
		t.Errorf("RepoPaths = %v", cfg.RepoPaths)
	}
}
