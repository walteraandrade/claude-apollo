package tui

import (
	"database/sql"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/walter/apollo/internal/config"
	"github.com/walter/apollo/internal/db"
	"github.com/walter/apollo/internal/git"
	"github.com/walter/apollo/internal/notifier"
	"github.com/walter/apollo/internal/style"
	"github.com/walter/apollo/internal/watcher"
)

type Screen int

const (
	ScreenBoard Screen = iota
	ScreenNote
)

type ColumnID int

const (
	ColNeedsReview ColumnID = iota
	ColReviewed
	ColIgnored
	NumColumns = 3
)

type BoardColumn struct {
	ID      ColumnID
	Title   string
	Status  string
	Commits []db.CommitRow
	Cursor  int
	Scroll  int
}

func (col *BoardColumn) Selected() *db.CommitRow {
	if len(col.Commits) == 0 || col.Cursor >= len(col.Commits) {
		return nil
	}
	return &col.Commits[col.Cursor]
}

func (col *BoardColumn) ClampCursor() {
	if col.Cursor >= len(col.Commits) {
		col.Cursor = max(0, len(col.Commits)-1)
	}
}

type Model struct {
	cfg      config.Config
	database *sql.DB
	notifier notifier.Notifier

	repo      *git.Repo
	repoID    int64
	watchCh   <-chan watcher.Event
	watchStop func()

	screen       Screen
	columns      [NumColumns]BoardColumn
	activeCol    ColumnID
	expandedHash string
	copiedHash   string
	stats        db.Stats
	noteInput    textinput.Model

	width  int
	height int
	err    error
}

func NewModel(cfg config.Config, database *sql.DB, n notifier.Notifier) Model {
	ti := textinput.New()
	ti.Placeholder = "Enter note..."
	ti.CharLimit = 256

	m := Model{
		cfg:       cfg,
		database:  database,
		notifier:  n,
		noteInput: ti,
	}

	m.columns[ColNeedsReview] = BoardColumn{
		ID: ColNeedsReview, Title: "Needs Review", Status: "unreviewed",
	}
	m.columns[ColReviewed] = BoardColumn{
		ID: ColReviewed, Title: "Reviewed", Status: "reviewed",
	}
	m.columns[ColIgnored] = BoardColumn{
		ID: ColIgnored, Title: "Ignored", Status: "ignored",
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return m.initRepo()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		if m.screen == ScreenNote {
			return m.updateNote(msg)
		}
		return m.updateKeys(msg)

	case RepoInitializedMsg:
		m.repoID = msg.RepoID
		m.repo = msg.Repo
		return m, m.seedCommits()

	case SeedDoneMsg:
		if len(msg.Commits) > 0 {
			return m, tea.Batch(m.persistCommits(msg.Commits), m.startWatcher())
		}
		return m, tea.Batch(m.loadCommits(), m.startWatcher())

	case WatcherReadyMsg:
		m.watchCh = msg.Ch
		m.watchStop = msg.Stop
		return m, m.listenWatcher()

	case WatcherEventMsg:
		return m, tea.Batch(m.readNewCommits(), m.listenWatcher())

	case NewCommitsMsg:
		if len(msg.Commits) == 0 {
			return m, nil
		}
		return m, m.persistCommits(msg.Commits)

	case CommitsPersistedMsg:
		return m, m.loadCommits()

	case CommitsLoadedMsg:
		m.stats = msg.Stats
		m.partitionCommits(msg.Commits)

	case ReviewUpdatedMsg:
		return m, m.loadCommits()

	case CopiedMsg:
		m.copiedHash = msg.Hash
		return m, tea.Tick(2*time.Second, func(time.Time) tea.Msg {
			return CopyFlashTickMsg{}
		})

	case CopyFlashTickMsg:
		m.copiedHash = ""

	case ErrorMsg:
		m.err = msg.Err
	}

	return m, nil
}

func (m *Model) partitionCommits(all []db.CommitRow) {
	buckets := [NumColumns][]db.CommitRow{}
	for _, c := range all {
		switch c.Status {
		case "unreviewed":
			buckets[ColNeedsReview] = append(buckets[ColNeedsReview], c)
		case "reviewed":
			buckets[ColReviewed] = append(buckets[ColReviewed], c)
		case "ignored":
			buckets[ColIgnored] = append(buckets[ColIgnored], c)
		}
	}
	for i := range NumColumns {
		m.columns[i].Commits = buckets[i]
		m.columns[i].ClampCursor()
	}
}

func (m Model) updateKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	action := MapKey(msg)
	col := &m.columns[m.activeCol]

	switch action {
	case ActionQuit:
		return m, m.quit()

	case ActionUp:
		m.expandedHash = ""
		if col.Cursor > 0 {
			col.Cursor--
		}

	case ActionDown:
		m.expandedHash = ""
		if col.Cursor < len(col.Commits)-1 {
			col.Cursor++
		}

	case ActionLeft:
		m.expandedHash = ""
		if m.activeCol > 0 {
			m.activeCol--
		}

	case ActionRight:
		m.expandedHash = ""
		if m.activeCol < NumColumns-1 {
			m.activeCol++
		}

	case ActionExpand:
		if c := col.Selected(); c != nil {
			if m.expandedHash == c.Hash {
				m.expandedHash = ""
			} else {
				m.expandedHash = c.Hash
			}
		}

	case ActionBack:
		m.expandedHash = ""

	case ActionReview:
		if c := col.Selected(); c != nil {
			return m, m.updateReview(c.Hash, "reviewed")
		}

	case ActionUnreview:
		if c := col.Selected(); c != nil {
			return m, m.updateReview(c.Hash, "unreviewed")
		}

	case ActionIgnore:
		if c := col.Selected(); c != nil {
			return m, m.updateReview(c.Hash, "ignored")
		}

	case ActionCopy:
		if c := col.Selected(); c != nil {
			return m, m.copyHashCmd(c.Hash[:min(7, len(c.Hash))])
		}

	case ActionNote:
		if c := col.Selected(); c != nil {
			m.noteInput.SetValue(c.Note)
			m.noteInput.Focus()
			m.screen = ScreenNote
		}
	}

	return m, nil
}

func (m Model) updateNote(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if c := m.selectedCommit(); c != nil {
			note := m.noteInput.Value()
			m.screen = ScreenBoard
			return m, func() tea.Msg {
				db.UpdateReviewStatus(m.database, c.Hash, c.Status, note)
				return ReviewUpdatedMsg{Hash: c.Hash, Status: c.Status}
			}
		}
		m.screen = ScreenBoard
	case "esc":
		m.screen = ScreenBoard
	default:
		var cmd tea.Cmd
		m.noteInput, cmd = m.noteInput.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	header := style.HeaderBar.Render("Apollo")
	var body string

	switch m.screen {
	case ScreenBoard:
		body = m.boardView()
	case ScreenNote:
		body = m.noteInputView()
	}

	errLine := m.errorView()
	status := m.statusBar()
	help := m.helpBar()

	return header + "\n" + body + "\n" + errLine + "\n" + status + "\n" + help
}

func (m Model) selectedCommit() *db.CommitRow {
	return m.columns[m.activeCol].Selected()
}

func (m Model) quit() tea.Cmd {
	if m.watchStop != nil {
		m.watchStop()
	}
	return tea.Quit
}
