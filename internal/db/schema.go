package db

var migrations = []string{
	`CREATE TABLE IF NOT EXISTS repositories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		path TEXT NOT NULL UNIQUE,
		active INTEGER NOT NULL DEFAULT 1,
		last_commit_hash TEXT NOT NULL DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,

	`CREATE TABLE IF NOT EXISTS commits (
		hash TEXT PRIMARY KEY,
		repo_id INTEGER NOT NULL REFERENCES repositories(id),
		author TEXT NOT NULL,
		subject TEXT NOT NULL,
		body TEXT NOT NULL DEFAULT '',
		branch TEXT NOT NULL DEFAULT '',
		committed_at DATETIME NOT NULL,
		detected_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,

	`CREATE TABLE IF NOT EXISTS review_state (
		commit_hash TEXT PRIMARY KEY REFERENCES commits(hash),
		status TEXT NOT NULL DEFAULT 'unreviewed',
		reviewed_at DATETIME,
		note TEXT NOT NULL DEFAULT ''
	)`,

	`CREATE TABLE IF NOT EXISTS events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		type TEXT NOT NULL,
		commit_hash TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		payload TEXT NOT NULL DEFAULT ''
	)`,

	`CREATE INDEX IF NOT EXISTS idx_commits_repo_time ON commits(repo_id, committed_at)`,
	`CREATE INDEX IF NOT EXISTS idx_review_status ON review_state(status)`,
	`CREATE INDEX IF NOT EXISTS idx_events_commit ON events(commit_hash, type)`,
}
