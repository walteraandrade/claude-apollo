package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func setupTestRepo(t *testing.T, numCommits int) string {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=test",
			"GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_NAME=test",
			"GIT_COMMITTER_EMAIL=test@test.com",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("%v: %s", err, out)
		}
	}

	run("git", "init")
	run("git", "checkout", "-b", "main")

	for i := 0; i < numCommits; i++ {
		f := filepath.Join(dir, "file.txt")
		os.WriteFile(f, []byte(time.Now().String()), 0644)
		run("git", "add", ".")
		run("git", "commit", "-m", "commit "+string(rune('A'+i)))
		time.Sleep(10 * time.Millisecond)
	}

	return dir
}

func TestOpenRepo(t *testing.T) {
	dir := setupTestRepo(t, 1)
	r, err := OpenRepo(dir)
	if err != nil {
		t.Fatal(err)
	}
	if r.CurrentBranch() != "main" {
		t.Errorf("branch = %q, want main", r.CurrentBranch())
	}
}

func TestOpenRepoInvalid(t *testing.T) {
	_, err := OpenRepo(t.TempDir())
	if err == nil {
		t.Error("expected error for non-repo dir")
	}
}

func TestSeedCommits(t *testing.T) {
	dir := setupTestRepo(t, 5)
	r, _ := OpenRepo(dir)

	commits, err := r.SeedCommits(3)
	if err != nil {
		t.Fatal(err)
	}
	if len(commits) != 3 {
		t.Fatalf("len = %d, want 3", len(commits))
	}
	if commits[0].Subject != "commit C" {
		t.Errorf("first = %q, want commit C", commits[0].Subject)
	}
	if commits[2].Subject != "commit E" {
		t.Errorf("last = %q, want commit E", commits[2].Subject)
	}
}

func TestReadNewCommits(t *testing.T) {
	dir := setupTestRepo(t, 5)
	r, _ := OpenRepo(dir)

	all, _ := r.SeedCommits(50)
	sinceHash := all[2].Hash

	newCommits, err := r.ReadNewCommits(sinceHash, 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(newCommits) != 2 {
		t.Fatalf("len = %d, want 2", len(newCommits))
	}
}

func TestReadNewCommitsEmptyRepo(t *testing.T) {
	dir := setupTestRepo(t, 3)
	r, _ := OpenRepo(dir)

	commits, err := r.ReadNewCommits("", 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(commits) != 3 {
		t.Errorf("len = %d, want 3", len(commits))
	}
}

func TestCommitInfoFields(t *testing.T) {
	dir := setupTestRepo(t, 1)
	r, _ := OpenRepo(dir)

	commits, _ := r.SeedCommits(1)
	c := commits[0]
	if c.Author != "test" {
		t.Errorf("author = %q", c.Author)
	}
	if c.Branch != "main" {
		t.Errorf("branch = %q", c.Branch)
	}
	if c.Hash == "" {
		t.Error("hash is empty")
	}
}
