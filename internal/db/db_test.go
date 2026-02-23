package db

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"
)

func testDB(t *testing.T) *testHelper {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return &testHelper{db: db, t: t}
}

type testHelper struct {
	db *sql.DB
	t  *testing.T
}

func (h *testHelper) mustRepo() int64 {
	id, err := UpsertRepo(h.db, "test", "/tmp/test")
	if err != nil {
		h.t.Fatal(err)
	}
	return id
}

func (h *testHelper) mustRepoAt(name, path string) int64 {
	id, err := UpsertRepo(h.db, name, path)
	if err != nil {
		h.t.Fatal(err)
	}
	return id
}

func TestOpenAndMigrate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
}

func TestUpsertRepo(t *testing.T) {
	h := testDB(t)
	id1, err := UpsertRepo(h.db, "test", "/tmp/test")
	if err != nil {
		t.Fatal(err)
	}
	id2, err := UpsertRepo(h.db, "test-updated", "/tmp/test")
	if err != nil {
		t.Fatal(err)
	}
	if id1 != id2 {
		t.Errorf("upsert should return same id, got %d and %d", id1, id2)
	}
}

func TestGetRepoByPath(t *testing.T) {
	h := testDB(t)
	h.mustRepo()

	r, err := GetRepoByPath(h.db, "/tmp/test")
	if err != nil {
		t.Fatal(err)
	}
	if r == nil {
		t.Fatal("repo not found")
	}
	if r.Name != "test" {
		t.Errorf("name = %q, want test", r.Name)
	}

	r, err = GetRepoByPath(h.db, "/nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if r != nil {
		t.Error("expected nil for nonexistent path")
	}
}

func TestInsertAndListCommits(t *testing.T) {
	h := testDB(t)
	repoID := h.mustRepo()

	now := time.Now().Truncate(time.Second)
	if err := InsertCommit(h.db, repoID, "abc123", "alice", "feat: add foo", "", "main", now); err != nil {
		t.Fatal(err)
	}
	if err := InsertCommit(h.db, repoID, "def456", "bob", "fix: bar", "body text", "main", now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}

	commits, err := ListCommits(h.db, repoID, FilterAll)
	if err != nil {
		t.Fatal(err)
	}
	if len(commits) != 2 {
		t.Fatalf("len = %d, want 2", len(commits))
	}
	if commits[0].Hash != "def456" {
		t.Error("expected most recent first")
	}
	if commits[0].Status != "unreviewed" {
		t.Errorf("status = %q, want unreviewed", commits[0].Status)
	}
	if commits[0].RepoName != "test" {
		t.Errorf("repo_name = %q, want test", commits[0].RepoName)
	}
}

func TestInsertCommitDuplicate(t *testing.T) {
	h := testDB(t)
	repoID := h.mustRepo()
	now := time.Now()

	if err := InsertCommit(h.db, repoID, "abc123", "alice", "msg", "", "main", now); err != nil {
		t.Fatal(err)
	}
	if err := InsertCommit(h.db, repoID, "abc123", "alice", "msg", "", "main", now); err != nil {
		t.Fatal("duplicate insert should not error")
	}
}

func TestUpdateReviewStatus(t *testing.T) {
	h := testDB(t)
	repoID := h.mustRepo()
	now := time.Now()

	if err := InsertCommit(h.db, repoID, "abc123", "alice", "msg", "", "main", now); err != nil {
		t.Fatal(err)
	}

	if err := UpdateReviewStatus(h.db, "abc123", "reviewed", "looks good"); err != nil {
		t.Fatal(err)
	}

	commits, err := ListCommits(h.db, repoID, FilterReviewed)
	if err != nil {
		t.Fatal(err)
	}
	if len(commits) != 1 {
		t.Fatalf("len = %d, want 1", len(commits))
	}
	if commits[0].Note != "looks good" {
		t.Errorf("note = %q, want 'looks good'", commits[0].Note)
	}
}

func TestFilterCommits(t *testing.T) {
	h := testDB(t)
	repoID := h.mustRepo()
	now := time.Now()

	InsertCommit(h.db, repoID, "a", "alice", "msg1", "", "main", now)
	InsertCommit(h.db, repoID, "b", "bob", "msg2", "", "main", now.Add(time.Minute))
	InsertCommit(h.db, repoID, "c", "carol", "msg3", "", "main", now.Add(2*time.Minute))

	UpdateReviewStatus(h.db, "b", "reviewed", "")
	UpdateReviewStatus(h.db, "c", "ignored", "")

	tests := []struct {
		filter ReviewFilter
		want   int
	}{
		{FilterAll, 3},
		{FilterUnreviewed, 1},
		{FilterReviewed, 1},
		{FilterIgnored, 1},
	}
	for _, tt := range tests {
		commits, err := ListCommits(h.db, repoID, tt.filter)
		if err != nil {
			t.Fatal(err)
		}
		if len(commits) != tt.want {
			t.Errorf("filter %q: got %d, want %d", tt.filter, len(commits), tt.want)
		}
	}
}

func TestGetStats(t *testing.T) {
	h := testDB(t)
	repoID := h.mustRepo()
	now := time.Now()

	InsertCommit(h.db, repoID, "a", "alice", "msg1", "", "main", now)
	InsertCommit(h.db, repoID, "b", "bob", "msg2", "", "main", now.Add(time.Minute))
	InsertCommit(h.db, repoID, "c", "carol", "msg3", "", "main", now.Add(2*time.Minute))
	UpdateReviewStatus(h.db, "b", "reviewed", "")

	stats, err := GetStats(h.db, repoID)
	if err != nil {
		t.Fatal(err)
	}
	if stats.Total != 3 || stats.Unreviewed != 2 || stats.Reviewed != 1 || stats.Ignored != 0 {
		t.Errorf("stats = %+v", stats)
	}
}

func TestUpdateLastCommitHash(t *testing.T) {
	h := testDB(t)
	repoID := h.mustRepo()

	if err := UpdateLastCommitHash(h.db, repoID, "abc123"); err != nil {
		t.Fatal(err)
	}

	r, err := GetRepoByPath(h.db, "/tmp/test")
	if err != nil {
		t.Fatal(err)
	}
	if r.LastCommitHash != "abc123" {
		t.Errorf("hash = %q, want abc123", r.LastCommitHash)
	}
}

func TestTimeRoundTrip(t *testing.T) {
	h := testDB(t)
	repoID := h.mustRepo()

	want := time.Date(2025, 6, 15, 10, 30, 45, 0, time.UTC)
	if err := InsertCommit(h.db, repoID, "timetest", "alice", "msg", "", "main", want); err != nil {
		t.Fatal(err)
	}

	commits, err := ListCommits(h.db, repoID, FilterAll)
	if err != nil {
		t.Fatal(err)
	}
	if len(commits) != 1 {
		t.Fatalf("len = %d, want 1", len(commits))
	}
	if !commits[0].CommittedAt.Equal(want) {
		t.Errorf("time = %v, want %v", commits[0].CommittedAt, want)
	}
}

func TestOpenBadPath(t *testing.T) {
	_, err := Open("/dev/null/nope")
	if err == nil {
		t.Fatal("expected error for invalid path")
	}
}

func TestInsertEvent(t *testing.T) {
	h := testDB(t)
	if err := InsertEvent(h.db, "commit_detected", "abc123", `{"branch":"main"}`); err != nil {
		t.Fatal(err)
	}
}

func TestListAllCommits(t *testing.T) {
	h := testDB(t)
	repoA := h.mustRepoAt("alpha", "/tmp/alpha")
	repoB := h.mustRepoAt("beta", "/tmp/beta")
	now := time.Now()

	InsertCommit(h.db, repoA, "a1", "alice", "alpha commit", "", "main", now)
	InsertCommit(h.db, repoB, "b1", "bob", "beta commit", "", "main", now.Add(time.Minute))

	commits, err := ListAllCommits(h.db, []int64{repoA, repoB}, FilterAll)
	if err != nil {
		t.Fatal(err)
	}
	if len(commits) != 2 {
		t.Fatalf("len = %d, want 2", len(commits))
	}
	if commits[0].RepoName != "beta" {
		t.Errorf("first commit repo = %q, want beta (most recent)", commits[0].RepoName)
	}
	if commits[1].RepoName != "alpha" {
		t.Errorf("second commit repo = %q, want alpha", commits[1].RepoName)
	}
}

func TestListAllCommitsEmpty(t *testing.T) {
	h := testDB(t)
	_ = h
	commits, err := ListAllCommits(h.db, nil, FilterAll)
	if err != nil {
		t.Fatal(err)
	}
	if len(commits) != 0 {
		t.Errorf("expected empty, got %d", len(commits))
	}
}

func TestGetAggregateStats(t *testing.T) {
	h := testDB(t)
	repoA := h.mustRepoAt("alpha", "/tmp/alpha")
	repoB := h.mustRepoAt("beta", "/tmp/beta")
	now := time.Now()

	InsertCommit(h.db, repoA, "a1", "alice", "msg", "", "main", now)
	InsertCommit(h.db, repoA, "a2", "alice", "msg", "", "main", now.Add(time.Minute))
	InsertCommit(h.db, repoB, "b1", "bob", "msg", "", "main", now.Add(2*time.Minute))
	UpdateReviewStatus(h.db, "a2", "reviewed", "")

	stats, err := GetAggregateStats(h.db, []int64{repoA, repoB})
	if err != nil {
		t.Fatal(err)
	}
	if stats.Total != 3 || stats.Unreviewed != 2 || stats.Reviewed != 1 {
		t.Errorf("stats = %+v", stats)
	}
}

func TestListActiveRepos(t *testing.T) {
	h := testDB(t)
	h.mustRepoAt("alpha", "/tmp/alpha")
	h.mustRepoAt("beta", "/tmp/beta")

	repos, err := ListActiveRepos(h.db)
	if err != nil {
		t.Fatal(err)
	}
	if len(repos) != 2 {
		t.Fatalf("len = %d, want 2", len(repos))
	}
}
