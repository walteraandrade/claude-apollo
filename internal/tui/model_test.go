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

func loadAndPartition(t *testing.T, m *Model) {
	t.Helper()
	commits, err := db.ListCommits(m.database, m.repoID, db.FilterAll)
	if err != nil {
		t.Fatal(err)
	}
	stats, err := db.GetStats(m.database, m.repoID)
	if err != nil {
		t.Fatal(err)
	}
	m.stats = stats
	m.partitionCommits(commits)
}

func TestNewModelDefaults(t *testing.T) {
	m := testModel(t)
	if m.screen != ScreenBoard {
		t.Errorf("screen = %d, want ScreenBoard", m.screen)
	}
	if m.activeCol != ColNeedsReview {
		t.Errorf("activeCol = %d, want ColNeedsReview", m.activeCol)
	}
}

func TestPartitionCommits(t *testing.T) {
	m := testModel(t)
	seedTestCommits(t, &m, 5)
	m.width = 120
	m.height = 24

	loadAndPartition(t, &m)

	if len(m.columns[ColNeedsReview].Commits) != 5 {
		t.Errorf("needs review = %d, want 5", len(m.columns[ColNeedsReview].Commits))
	}
	if len(m.columns[ColReviewed].Commits) != 0 {
		t.Errorf("reviewed = %d, want 0", len(m.columns[ColReviewed].Commits))
	}
	if len(m.columns[ColIgnored].Commits) != 0 {
		t.Errorf("ignored = %d, want 0", len(m.columns[ColIgnored].Commits))
	}
}

func TestColumnNavigation(t *testing.T) {
	m := testModel(t)
	seedTestCommits(t, &m, 3)
	m.width = 120
	m.height = 24
	loadAndPartition(t, &m)

	if m.activeCol != ColNeedsReview {
		t.Fatal("should start on ColNeedsReview")
	}

	// Move right
	result, _ := m.updateKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	rm := result.(Model)
	if rm.activeCol != ColReviewed {
		t.Errorf("after l: activeCol = %d, want ColReviewed", rm.activeCol)
	}

	// Move right again
	result, _ = rm.updateKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	rm = result.(Model)
	if rm.activeCol != ColIgnored {
		t.Errorf("after l+l: activeCol = %d, want ColIgnored", rm.activeCol)
	}

	// Can't go past last
	result, _ = rm.updateKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	rm = result.(Model)
	if rm.activeCol != ColIgnored {
		t.Errorf("should stay on ColIgnored, got %d", rm.activeCol)
	}

	// Move left
	result, _ = rm.updateKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	rm = result.(Model)
	if rm.activeCol != ColReviewed {
		t.Errorf("after h: activeCol = %d, want ColReviewed", rm.activeCol)
	}
}

func TestCardNavigation(t *testing.T) {
	m := testModel(t)
	seedTestCommits(t, &m, 5)
	m.width = 120
	m.height = 24
	loadAndPartition(t, &m)

	col := &m.columns[ColNeedsReview]
	if col.Cursor != 0 {
		t.Fatal("cursor should start at 0")
	}

	// Down
	result, _ := m.updateKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	rm := result.(Model)
	if rm.columns[ColNeedsReview].Cursor != 1 {
		t.Errorf("after j: cursor = %d, want 1", rm.columns[ColNeedsReview].Cursor)
	}

	// Up
	result, _ = rm.updateKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	rm = result.(Model)
	if rm.columns[ColNeedsReview].Cursor != 0 {
		t.Errorf("after k: cursor = %d, want 0", rm.columns[ColNeedsReview].Cursor)
	}

	// Don't go below 0
	result, _ = rm.updateKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	rm = result.(Model)
	if rm.columns[ColNeedsReview].Cursor != 0 {
		t.Errorf("cursor should not go below 0, got %d", rm.columns[ColNeedsReview].Cursor)
	}
}

func TestCardExpansion(t *testing.T) {
	m := testModel(t)
	seedTestCommits(t, &m, 3)
	m.width = 120
	m.height = 24
	loadAndPartition(t, &m)

	// Enter expands
	result, _ := m.updateKeys(tea.KeyMsg{Type: tea.KeyEnter})
	rm := result.(Model)
	expected := m.columns[ColNeedsReview].Commits[0].Hash
	if rm.expandedHash != expected {
		t.Errorf("expandedHash = %q, want %q", rm.expandedHash, expected)
	}

	// Enter again collapses
	result, _ = rm.updateKeys(tea.KeyMsg{Type: tea.KeyEnter})
	rm = result.(Model)
	if rm.expandedHash != "" {
		t.Errorf("expandedHash should be empty after toggle, got %q", rm.expandedHash)
	}

	// Expand then esc collapses
	result, _ = m.updateKeys(tea.KeyMsg{Type: tea.KeyEnter})
	rm = result.(Model)
	result, _ = rm.updateKeys(tea.KeyMsg{Type: tea.KeyEsc})
	rm = result.(Model)
	if rm.expandedHash != "" {
		t.Errorf("expandedHash should be empty after esc, got %q", rm.expandedHash)
	}
}

func TestStatusChangeMovesBetweenColumns(t *testing.T) {
	m := testModel(t)
	seedTestCommits(t, &m, 3)
	m.width = 120
	m.height = 24
	loadAndPartition(t, &m)

	if len(m.columns[ColNeedsReview].Commits) != 3 {
		t.Fatalf("needs review = %d, want 3", len(m.columns[ColNeedsReview].Commits))
	}

	hash := m.columns[ColNeedsReview].Commits[0].Hash
	db.UpdateReviewStatus(m.database, hash, "reviewed", "")

	loadAndPartition(t, &m)

	if len(m.columns[ColNeedsReview].Commits) != 2 {
		t.Errorf("needs review after move = %d, want 2", len(m.columns[ColNeedsReview].Commits))
	}
	if len(m.columns[ColReviewed].Commits) != 1 {
		t.Errorf("reviewed after move = %d, want 1", len(m.columns[ColReviewed].Commits))
	}
}

func TestCommitsLoadedMsg(t *testing.T) {
	m := testModel(t)
	m.width = 120
	m.height = 24

	msg := CommitsLoadedMsg{
		Commits: []db.CommitRow{
			{Hash: "abc1234", Status: "unreviewed"},
			{Hash: "def5678", Status: "reviewed"},
		},
		Stats: db.Stats{Total: 2, Unreviewed: 1, Reviewed: 1},
	}

	result, _ := m.Update(msg)
	rm := result.(Model)
	if len(rm.columns[ColNeedsReview].Commits) != 1 {
		t.Errorf("needs review = %d, want 1", len(rm.columns[ColNeedsReview].Commits))
	}
	if len(rm.columns[ColReviewed].Commits) != 1 {
		t.Errorf("reviewed = %d, want 1", len(rm.columns[ColReviewed].Commits))
	}
}

func TestCopyAction(t *testing.T) {
	m := testModel(t)
	seedTestCommits(t, &m, 1)
	m.width = 120
	m.height = 24
	loadAndPartition(t, &m)

	result, cmd := m.updateKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	rm := result.(Model)
	_ = rm
	if cmd == nil {
		t.Error("copy should return a command")
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
		{"h", ActionLeft},
		{"l", ActionRight},
		{"r", ActionReview},
		{"u", ActionUnreview},
		{"i", ActionIgnore},
		{"c", ActionCopy},
		{"n", ActionNote},
	}
	for _, tt := range tests {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
		got := MapKey(msg)
		if got != tt.want {
			t.Errorf("MapKey(%q) = %d, want %d", tt.key, got, tt.want)
		}
	}
}

func TestTabCyclesColumns(t *testing.T) {
	m := testModel(t)
	m.width = 120
	m.height = 24

	result, _ := m.updateKeys(tea.KeyMsg{Type: tea.KeyTab})
	rm := result.(Model)
	if rm.activeCol != ColReviewed {
		t.Errorf("after tab: activeCol = %d, want ColReviewed", rm.activeCol)
	}
}
