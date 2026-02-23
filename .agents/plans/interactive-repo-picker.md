# Interactive Repo Picker

## Context
Repos must be manually specified via config.toml or CLI args every time. We'll add an in-app screen that scans a parent directory for git repos, shows them in a filterable list, and lets the user toggle repos on/off. Selected repos persist to config.toml and start watching immediately.

## Key Files
- `internal/config/config.go` — add `ScanPaths` field, persist logic
- `internal/tui/model.go` — add `ScreenAddRepo`, new fields for repo picker
- `internal/tui/keys.go` — add `ActionAddRepo` (bound to `a`)
- `internal/tui/views.go` — add help bar entry, render repo picker screen
- `internal/tui/commands.go` — scan directory cmd, add repo cmd
- `internal/tui/messages.go` — new messages
- `internal/tui/repopicker.go` — **new file** for picker view + update logic

## Plan

### 1. Config: add `ScanPaths`
In `config.go`, add `ScanPaths []string` to `Config` struct (toml: `scan_paths`). In `Defaults()`, default to `["~/Github"]`. Add env override `APOLLO_SCAN_PATHS`.

### 2. New screen: `ScreenAddRepo`
Add `ScreenAddRepo` constant to the `Screen` enum in `model.go`.

### 3. Model additions
Add to `Model` struct:
- `repoPickerItems []repoPickerItem` — scanned repos
- `repoPickerCursor int`
- `repoPickerFilter textinput.Model` — for fuzzy search
- `repoPickerFiltered []int` — indices into items matching filter

`repoPickerItem` struct: `{ Path, Name string; Watching bool }`.

### 4. Scan directory command
New `scanForRepos()` tea.Cmd in `commands.go`:
- For each path in `cfg.ScanPaths`, expand `~`, read dir entries
- For each entry, check if it's a dir containing `.git`
- Mark `Watching: true` if path is in current `m.handles`
- Return `ReposScanDoneMsg{Items []repoPickerItem}`

### 5. Add repo command
New `addRepoToWatch(path string)` tea.Cmd:
- Append path to `cfg.RepoPaths`, dedup
- Call `config.Save(cfg)`
- Open the git repo, upsert to DB, seed commits, start watcher, add to mux
- Return `RepoAddedMsg{Handle RepoHandle}`

### 6. Key binding
- `a` → `ActionAddRepo` — triggers scan + switches to `ScreenAddRepo`
- Inside picker screen: `j/k` nav, `enter` add, `esc` back, typing filters

### 7. Picker view (`repopicker.go`)
- Header: "Add Repository"
- Filter input at top
- List of repos with `●` (watching) / `○` (available) indicators
- Repo name bold, path muted underneath
- Help bar: `j/k nav  enter add  esc back  / filter`

### 8. Update handler
- On `ScreenAddRepo`, intercept keys in `Update()`
- Filter text updates narrow the visible list
- Enter on an unwatched repo triggers `addRepoToWatch`
- Esc returns to `ScreenBoard`

### 9. Wire into existing flow
- `RepoAddedMsg` handler: append handle to `m.handles`, update `m.handleIdx`, reload commits
- Picker screen re-renders with updated watching status

## Verification
1. `go build` succeeds
2. Run apollo, press `a` — picker shows scanned repos from `~/Github`
3. Type to filter repos, enter to add one
4. Verify `~/.apollo/config.toml` updated with new path
5. Verify new repo's commits appear on the board immediately
6. `go test ./...` passes
