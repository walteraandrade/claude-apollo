# Architecture Lens

## Focus Areas

### Package Boundaries
Each package has ONE concern:
- `config` — config load/save/env overrides only
- `db` — SQLite CRUD only (schema, repos, commits, review state, events)
- `git` — git operations only (open repo, read commits)
- `watcher` — fsnotify watcher only (debounced ref change events)
- `notifier` — notification interface + implementations only (Desktop, Fallback)
- `tui` — Bubbletea UI only (model, views, keys, messages, commands)
- `style` — Lipgloss styles + theme only

**Critical rule**: All packages except `tui` and `main` are leaf nodes — they MUST NOT cross-import each other.
- OK: `tui` imports `db`, `git`, `config`, `notifier`, `watcher`, `style`
- NOT OK: `db` importing `git`, `watcher` importing `db`, `notifier` importing `tui`
- Any leaf-to-leaf import = blocking

### Dependency Direction
- Dependencies flow inward: `main → tui → {db, git, config, notifier, watcher, style}`
- Leaf packages depend only on stdlib and external libraries
- `notifier` defines `Notifier` interface — `tui` accepts it, doesn't know concrete type
- `watcher` returns channels — consumer doesn't know implementation

### Data Flow Pipeline
Events flow through this pipeline:
```
watcher.Event → git.ReadNewCommits → db.InsertCommit → notifier.Notify → TUI message
```
Specifically:
1. `WatcherEventMsg` triggers `readNewCommits()`
2. `NewCommitsMsg` triggers `persistCommits()` (inserts + notifies)
3. `CommitsPersistedMsg` triggers `loadCommits()` (DB read)
4. `CommitsLoadedMsg` updates UI (partition + stats)

- Skipping steps = blocking (e.g., loading before persisting)
- Reordering steps = blocking (e.g., notifying before inserting)

### Config Resolution
Priority: CLI arg > env var > TOML file > defaults
- `Defaults()` provides base values (`SeedDepth: 50`, `DebounceMs: 300`)
- `Load()` reads TOML, then calls `applyEnvOverrides()`
- CLI args override in `main.go` (if present)
- Changes that break this priority = blocking

### Extension Points
1. **New Notifier**: implement `Notifier` interface → construct in `notifier.New()` or `main.go`
2. **New TUI Screen**: add `Screen` constant → add message types → wire in `Update`
3. **New review status**: add status string → update `partitionCommits` column mapping → add migration if needed

Changes that make these harder to extend = important.

### Entry Point
- `main.go` should be thin: load config, open DB, create notifier, create model, run tea
- Business logic in main.go = important
- Config validation in main.go is OK (it's wiring)

### File Organization in `db` Package
- `db.go` — Open, pragmas, migration runner
- `schema.go` — migration DDL statements
- `commit.go` — commit + review_state CRUD
- `repo.go` — repository CRUD
- `event.go` — event logging
- Mixing concerns across these files = important

## Severity Guide

| Issue | Severity |
|-------|----------|
| Leaf package importing another leaf package | blocking |
| Data flow pipeline order violation | blocking |
| Config resolution priority broken | blocking |
| Package doing work outside its concern | important |
| Business logic in main.go | important |
| Extension point made harder to use | important |
| db file organization violation | important |
| Unnecessary interface (no polymorphism) | nit |
| Minor dependency that could be inverted | nit |
