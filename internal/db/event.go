package db

import "database/sql"

func InsertEvent(db *sql.DB, eventType, commitHash, payload string) error {
	_, err := db.Exec(
		`INSERT INTO events (type, commit_hash, payload) VALUES (?, ?, ?)`,
		eventType, commitHash, payload,
	)
	return err
}
