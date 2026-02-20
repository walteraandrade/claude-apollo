package tui

import (
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/walter/apollo/internal/config"
	"github.com/walter/apollo/internal/db"
	"github.com/walter/apollo/internal/notifier"
)

func testModel(t *testing.T) Model {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })

	cfg := config.Defaults()
	cfg.RepoPath = t.TempDir()

	return NewModel(cfg, database, &notifier.Fallback{})
}

func seedTestCommits(t *testing.T, m *Model, n int) {
	t.Helper()
	repoID, err := db.UpsertRepo(m.database, "test", m.cfg.RepoPath)
	if err != nil {
		t.Fatal(err)
	}
	m.repoID = repoID

	now := time.Now()
	for i := range n {
		hash := string(rune('a'+i)) + "bcdef1234567890abcdef1234567890abcdef12"
		db.InsertCommit(m.database, repoID, hash, "alice", "commit "+string(rune('A'+i)), "", "main", now.Add(time.Duration(i)*time.Minute))
	}
}

func TestNewModelDefaults(t *testing.T) {
	m := testModel(t)
	if m.filter != db.FilterUnreviewed {
		t.Errorf("filter = %q, want unreviewed", m.filter)
	}
	if m.screen != ScreenList {
		t.Errorf("screen = %d, want ScreenList", m.screen)
	}
}

func TestNextFilter(t *testing.T) {
	tests := []struct {
		in   db.ReviewFilter
		want db.ReviewFilter
	}{
		{db.FilterUnreviewed, db.FilterAll},
		{db.FilterAll, db.FilterReviewed},
		{db.FilterReviewed, db.FilterIgnored},
		{db.FilterIgnored, db.FilterUnreviewed},
	}
	for _, tt := range tests {
		got := nextFilter(tt.in)
		if got != tt.want {
			t.Errorf("nextFilter(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestNavigationKeys(t *testing.T) {
	m := testModel(t)
	seedTestCommits(t, &m, 5)

	m.width = 80
	m.height = 24
	commits, _ := db.ListCommits(m.database, m.repoID, db.FilterAll)
	m.commits = commits
	m.filter = db.FilterAll

	if m.cursor != 0 {
		t.Fatal("cursor should start at 0")
	}

	// Down
	result, _ := m.updateKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	rm := result.(Model)
	if rm.cursor != 1 {
		t.Errorf("after j: cursor = %d, want 1", rm.cursor)
	}

	// Up
	result, _ = rm.updateKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	rm = result.(Model)
	if rm.cursor != 0 {
		t.Errorf("after k: cursor = %d, want 0", rm.cursor)
	}

	// Don't go below 0
	result, _ = rm.updateKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	rm = result.(Model)
	if rm.cursor != 0 {
		t.Errorf("cursor should not go below 0, got %d", rm.cursor)
	}
}

func TestDetailScreenToggle(t *testing.T) {
	m := testModel(t)
	seedTestCommits(t, &m, 3)
	m.width = 80
	m.height = 24
	commits, _ := db.ListCommits(m.database, m.repoID, db.FilterAll)
	m.commits = commits
	m.filter = db.FilterAll

	// Enter detail
	result, _ := m.updateKeys(tea.KeyMsg{Type: tea.KeyEnter})
	rm := result.(Model)
	if rm.screen != ScreenDetail {
		t.Errorf("screen = %d, want ScreenDetail", rm.screen)
	}

	// Esc back
	result, _ = rm.updateKeys(tea.KeyMsg{Type: tea.KeyEsc})
	rm = result.(Model)
	if rm.screen != ScreenList {
		t.Errorf("screen = %d, want ScreenList", rm.screen)
	}
}

func TestCommitsLoadedMsg(t *testing.T) {
	m := testModel(t)
	m.width = 80
	m.height = 24
	m.cursor = 99

	msg := CommitsLoadedMsg{
		Commits: []db.CommitRow{{Hash: "abc1234"}},
		Stats:   db.Stats{Total: 1, Unreviewed: 1},
	}

	result, _ := m.Update(msg)
	rm := result.(Model)
	if rm.cursor != 0 {
		t.Errorf("cursor should clamp to 0, got %d", rm.cursor)
	}
	if len(rm.commits) != 1 {
		t.Errorf("commits len = %d, want 1", len(rm.commits))
	}
}

func TestMapKey(t *testing.T) {
	tests := []struct {
		key  string
		want Action
	}{
		{"q", ActionQuit},
		{"j", ActionDown},
		{"k", ActionUp},
		{"r", ActionReview},
		{"u", ActionUnreview},
		{"i", ActionIgnore},
		{"tab", ActionCycleFilter},
		{"n", ActionNote},
	}
	for _, tt := range tests {
		var msg tea.KeyMsg
		if tt.key == "tab" {
			msg = tea.KeyMsg{Type: tea.KeyTab}
		} else {
			msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
		}
		got := MapKey(msg)
		if got != tt.want {
			t.Errorf("MapKey(%q) = %d, want %d", tt.key, got, tt.want)
		}
	}
}
