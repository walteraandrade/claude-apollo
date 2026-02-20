# Apollo - Concept Brief (Pre-Plan)

## Product Intent

Apollo is a local-first terminal app for commit review triage.

Primary loop:

1. detect local git commits
2. notify user that commit might need review
3. track review status in local DB
4. surface commit queue in TUI

Even tiny commits should trigger notification. User decides review outcome.

## Core Outcomes

- Never miss a local commit that should be triaged.
- Keep a persistent backlog of unreviewed commits.
- Give fast keyboard-driven status updates in terminal UI.
- Stay offline-capable and local-only by default.

## Scope (v1)

- Single-user, local machine.
- Local repositories only.
- Local SQLite storage only.
- Terminal UI as primary interface.
- Notifications always emitted on commit detection.

## Domain Model (v1)

### Repositories

- `id`
- `name`
- `path` (unique)
- `active`
- `created_at`

### Commits

- `hash` (unique)
- `repo_id`
- `author`
- `subject`
- `body` (optional)
- `branch`
- `committed_at`
- `detected_at`

### Review State

- `commit_hash`
- `status` (`unreviewed`, `reviewed`, `ignored`)
- `reviewed_at` (nullable)
- `note` (optional)

### Events (optional but useful)

- `type` (`detected`, `notified`, `status_changed`, `error`)
- `commit_hash` (nullable for system errors)
- `created_at`
- `payload` (optional JSON/text)

## System Components

### Watcher

- Watch `.git` ref changes.
- Debounce rapid write bursts.
- Emit normalized "repo changed" event.

### Git Ingest

- Read new commits since last seen hash.
- Deduplicate by commit hash.
- Seed initial state with bounded history (if configured).

### Notifier

- Emit notification for every detected commit.
- Message baseline: "A commit might need a review."
- Fallback to terminal-safe output if desktop notifications unavailable.

### Storage

- SQLite with WAL mode.
- Indexes for repo/time and status queries.
- Simple migration strategy from app startup.

### TUI

- Main panel lists commits + status marker.
- Filters: all, unreviewed, reviewed, ignored.
- Actions: mark reviewed/unreviewed/ignored, open details.

## UX Notes

- Keep status language explicit (`reviewed` / `unreviewed`), symbols optional.
- Prioritize unreviewed by default sort.
- Show empty states clearly ("No unreviewed commits").
- Fast keybindings matter more than decorative layout in v1.

## TDD Constraints To Carry Forward

- Strict red-green-refactor for each feature slice.
- Table-driven tests for parser/state transitions.
- Integration tests with temp git repos for watcher + ingest.
- Avoid over-mocking; test real behavior where cheap.
- Run race detector and coverage in regular workflow.

## Influences from Existing Local Projects

- `homer`: useful pattern for watcher -> ingest -> DB -> TUI pipeline.
- `mr-argus`: useful operational mindset for continuous local monitoring.

Apollo should reuse those architectural strengths while remaining focused on review tracking.

## Open Decisions Before Full Plan

- v1 target: one repo vs many repos.
- Notification channel: desktop only vs desktop + terminal.
- Initial seed behavior: start empty vs ingest last N commits.
- Status semantics: keep `ignored` in v1 or defer.
- Merge/bot commits: include by default or filter.
