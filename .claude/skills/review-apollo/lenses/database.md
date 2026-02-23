# Database Lens

## Focus Areas

### Schema (4 Tables)
Per spec and `internal/db/schema.go`:
1. **`repositories`** — `id`, `name`, `path` (UNIQUE), `active`, `last_commit_hash`, `created_at`
2. **`commits`** — `hash` (PK), `repo_id` (FK→repositories), `author`, `subject`, `body`, `branch`, `committed_at`, `detected_at`
3. **`review_state`** — `commit_hash` (PK, FK→commits), `status` (default `'unreviewed'`), `reviewed_at`, `note`
4. **`events`** — `id`, `type`, `commit_hash`, `created_at`, `payload`

Indexes: `idx_commits_repo_time`, `idx_review_status`, `idx_events_commit`
- Deviating from this schema without updating migrations = important

### Review State Transitions
Valid statuses: `unreviewed`, `reviewed`, `ignored`
- Default on insert: `unreviewed` (set by DB default on `review_state`)
- `reviewed_at` set to `time.Now()` only when status = `"reviewed"`, nil otherwise
- `UpdateReviewStatus` accepts any status string — no validation at DB layer
- Adding new status values must update `partitionCommits` column mapping in TUI

### Insert Order Enforcement
Foreign keys are ON — insert order matters:
1. `repositories` first (commits reference `repo_id`)
2. `commits` second (review_state references `commit_hash`)
3. `review_state` third (created in same `InsertCommit` call)
- `InsertCommit` correctly does: INSERT commit → INSERT review_state (in order)
- `INSERT OR IGNORE` handles duplicate commits gracefully
- FK violation at runtime = blocking

### Migration Safety
- All migrations in `schema.go` as entries in the `migrations` slice
- Use `CREATE TABLE IF NOT EXISTS` — must be idempotent
- New columns: use `ALTER TABLE ... ADD COLUMN` with IF NOT EXISTS
- NEVER drop tables or columns in migrations (data loss)
- Migrations run sequentially on `db.Open()` — order matters
- Adding migration out of order = blocking

### SQL Injection
- Always use parameterized queries (`?` placeholders)
- NEVER use `fmt.Sprintf` to build SQL with user input
- **Note**: `ListCommits` uses string concatenation for `filter`: `'` + string(filter) + `'`
  - Currently safe because `ReviewFilter` is a typed constant (`FilterAll`, `FilterUnreviewed`, etc.)
  - Flag as blocking ONLY if the input source changes to accept user-controlled strings
  - If filter value comes from user input (URL param, CLI arg, etc.) = blocking

### WAL Mode & Concurrency
- DB opens with WAL journal mode and 5s busy timeout
- `SetMaxOpenConns(1)` — serializes all access through single connection
- Multiple goroutines can safely use `*sql.DB` — it handles pooling internally
- Long-running transactions block other writers — avoid
- Missing WAL pragma on new DB setup = blocking

### Query Patterns
- All DB functions accept `*sql.DB` as first param (package-level functions, not methods)
- Query functions return domain types (`CommitRow`, `Stats`), not `sql.Rows`
- Always close `sql.Rows` (use `defer rows.Close()`)
- Check `rows.Err()` after iteration loop
- Missing `rows.Close()` = resource leak = blocking

### Connection Management
- Single `*sql.DB` instance created in `main.go`, passed to `tui.NewModel`
- No manual `Open()`/`Close()` per query — use the shared instance
- `db.Close()` should happen on app shutdown

### Performance
- `ListCommits` has no LIMIT — grows with commit count
- Acceptable for v1 (local-only, bounded repo history)
- Flag unbounded SELECT as important only if table growth is unbounded
- Indexes exist on `(repo_id, committed_at)` and `(status)` — covers main queries

## Severity Guide

| Issue | Severity |
|-------|----------|
| SQL injection (user-controlled string interpolation in query) | blocking |
| Missing rows.Close() / resource leak | blocking |
| Migration not idempotent | blocking |
| Dropping table/column in migration | blocking |
| FK violation possible at runtime | blocking |
| Insert order violation | blocking |
| Missing rows.Err() check after iteration | important |
| Unbounded SELECT on unbounded table | important |
| Schema deviates from spec without migration | important |
| Missing WAL pragma | blocking |
| String concat in SQL with typed constants (current state) | nit |
| Missing index on frequently queried column | nit |
| Could use SELECT with specific columns | nit |
