# Bubbletea & TUI Lens

## Focus Areas

### Elm Architecture Compliance
- Model: pure data struct, no side effects in fields
- Update: returns `(tea.Model, tea.Cmd)` — never performs I/O directly
- View: pure function of model state — NEVER mutates model, NEVER triggers side effects
- Side effects go through `tea.Cmd` functions only
- Blocking calls in Update = blocking (freezes TUI)

### Kanban Board Architecture
Apollo uses a 3-column kanban board (`ScreenBoard`):
- Columns: `ColNeedsReview` ("Needs Review"), `ColReviewed` ("Reviewed"), `ColIgnored` ("Ignored")
- Each `BoardColumn` has its own `Cursor` and `Scroll` state
- `partitionCommits()` distributes commits to columns by status, then calls `ClampCursor()` on each
- `ClampCursor()` MUST be called after any partition/filter that changes column contents
- Missing `ClampCursor()` after content change = blocking (cursor can index out of bounds)

### Card Rendering
- Compact card: shows hash + subject in column width
- Expanded card: toggled by `expandedHash` — shows full commit details
- `columnWidth()` must account for borders, gaps, terminal width
- Visible range computed from `Scroll` + available height — cards outside range not rendered
- Truncation must respect terminal width to avoid wrapping artifacts

### Message Flow
Startup sequence:
```
Init → initRepo() → RepoInitializedMsg
  → seedCommits() → SeedDoneMsg
    → persistCommits() + startWatcher() [parallel via tea.Batch]
      → CommitsPersistedMsg → loadCommits() → CommitsLoadedMsg → partitionCommits()
      → WatcherReadyMsg → listenWatcher() → [watcher loop]
```

Watcher loop:
```
WatcherEventMsg → readNewCommits() + listenWatcher() [parallel via tea.Batch]
  → NewCommitsMsg → persistCommits() → CommitsPersistedMsg → loadCommits() → CommitsLoadedMsg
```

Review action:
```
ActionReview/Unreview/Ignore → updateReview() → ReviewUpdatedMsg → loadCommits() → CommitsLoadedMsg
```

- Breaking message ordering = blocking (data inconsistency)
- `tea.Batch` used for independent concurrent commands — correct
- Watcher re-subscribes via `listenWatcher()` after each event — ensures continuous listening

### Key Dispatch
- `MapKey(msg tea.KeyMsg) Action` in `keys.go` maps raw keys to typed `Action` enum
- NO hardcoded key strings in `updateKeys()` — only `Action` constants
- Adding a key handler that bypasses `MapKey` = important
- Key conflicts (same key mapped to different actions) = blocking
- `tab`/`shift+tab` mapped to `ActionRight`/`ActionLeft` — column navigation aliases

### Screen System
Two screens only:
- `ScreenBoard` — main kanban view, handles all navigation + review actions
- `ScreenNote` — modal text input for commit notes, handles enter/esc/text input
- Screen transitions: `ScreenBoard → ScreenNote` via `ActionNote`, back via enter/esc
- `updateNote()` uses `msg.String()` directly for enter/esc — acceptable for modal input
- `WindowSizeMsg` handled at top level, propagates to all screens via shared `m.width`/`m.height`

### Commands & Subscriptions
- `tea.Cmd` for one-shot async work (DB queries, git reads, review updates)
- Watcher listening uses channel receive in a `func() tea.Msg` — re-subscribed each event
- Commands must not block — wrap blocking calls in goroutines via `func() tea.Msg`
- Return `tea.Batch(cmds...)` for multiple concurrent commands
- `SeedDoneMsg` handler branches: if commits found → persist+watch; else → load+watch

### State Consistency
- Model state must be consistent after every Update call
- `expandedHash` cleared on navigation (up/down/left/right/back) — prevents stale expansion
- `copiedHash` cleared after 2s tick — flash feedback pattern
- After `CommitsLoadedMsg`: stats updated AND columns partitioned in same handler — atomic
- `partitionCommits` is a pointer receiver method mutating `m` — correct for batch state update

### Styles
- All styles in `internal/style/` — `theme.go`, `style.go`, `components.go`
- Don't inline `lipgloss.NewStyle()` in View methods — use predefined styles
- Respect terminal width: layout adapts to `m.width`

## Severity Guide

| Issue | Severity |
|-------|----------|
| Blocking I/O in Update (DB query, HTTP call without Cmd) | blocking |
| Side effects in View method | blocking |
| Missing ClampCursor after partition/filter change | blocking |
| Key binding conflict between active screens | blocking |
| Message flow order violated | blocking |
| Missing WindowSizeMsg handling | important |
| Message type not in messages.go | important |
| Hardcoded key string instead of Action enum | important |
| Inconsistent state after Update | important |
| Watcher loop not re-subscribing | important |
| Inline lipgloss styles in View | nit |
| Minor scroll behavior edge case | nit |
