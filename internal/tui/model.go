package tui

import (
	"database/sql"

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
	ScreenList Screen = iota
	ScreenDetail
	ScreenNote
)

type Model struct {
	cfg      config.Config
	database *sql.DB
	notifier notifier.Notifier

	repo      *git.Repo
	repoID    int64
	watchCh   <-chan watcher.Event
	watchStop func()

	screen    Screen
	commits   []db.CommitRow
	cursor    int
	filter    db.ReviewFilter
	stats     db.Stats
	noteInput textinput.Model

	width  int
	height int
	err    error
}

func NewModel(cfg config.Config, database *sql.DB, n notifier.Notifier) Model {
	ti := textinput.New()
	ti.Placeholder = "Enter note..."
	ti.CharLimit = 256

	return Model{
		cfg:       cfg,
		database:  database,
		notifier:  n,
		filter:    db.FilterUnreviewed,
		noteInput: ti,
	}
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
		m.commits = msg.Commits
		m.stats = msg.Stats
		if m.cursor >= len(m.commits) {
			m.cursor = max(0, len(m.commits)-1)
		}

	case ReviewUpdatedMsg:
		return m, m.loadCommits()

	case ErrorMsg:
		m.err = msg.Err
	}

	return m, nil
}

func (m Model) updateKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	action := MapKey(msg)

	switch action {
	case ActionQuit:
		return m, m.quit()

	case ActionUp:
		if m.cursor > 0 {
			m.cursor--
		}

	case ActionDown:
		if m.cursor < len(m.commits)-1 {
			m.cursor++
		}

	case ActionReview:
		if c := m.selectedCommit(); c != nil {
			return m, m.updateReview(c.Hash, "reviewed")
		}

	case ActionUnreview:
		if c := m.selectedCommit(); c != nil {
			return m, m.updateReview(c.Hash, "unreviewed")
		}

	case ActionIgnore:
		if c := m.selectedCommit(); c != nil {
			return m, m.updateReview(c.Hash, "ignored")
		}

	case ActionDetail:
		if len(m.commits) > 0 {
			m.screen = ScreenDetail
		}

	case ActionBack:
		if m.screen == ScreenDetail {
			m.screen = ScreenList
		}

	case ActionCycleFilter:
		m.filter = nextFilter(m.filter)
		m.cursor = 0
		return m, m.loadCommits()

	case ActionNote:
		if c := m.selectedCommit(); c != nil {
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
			m.screen = ScreenList
			return m, func() tea.Msg {
				db.UpdateReviewStatus(m.database, c.Hash, c.Status, note)
				return ReviewUpdatedMsg{Hash: c.Hash, Status: c.Status}
			}
		}
		m.screen = ScreenList
	case "esc":
		m.screen = ScreenList
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
	case ScreenList:
		body = m.listView()
	case ScreenDetail:
		body = m.detailView()
	case ScreenNote:
		body = m.noteInputView()
	}

	errLine := m.errorView()
	status := m.statusBar()
	help := m.helpBar()

	return header + "\n" + body + "\n\n" + errLine + "\n" + status + "\n" + help
}

func (m Model) selectedCommit() *db.CommitRow {
	if m.cursor >= len(m.commits) {
		return nil
	}
	return &m.commits[m.cursor]
}

func (m Model) quit() tea.Cmd {
	if m.watchStop != nil {
		m.watchStop()
	}
	return tea.Quit
}

func nextFilter(f db.ReviewFilter) db.ReviewFilter {
	switch f {
	case db.FilterUnreviewed:
		return db.FilterAll
	case db.FilterAll:
		return db.FilterReviewed
	case db.FilterReviewed:
		return db.FilterIgnored
	case db.FilterIgnored:
		return db.FilterUnreviewed
	default:
		return db.FilterUnreviewed
	}
}
