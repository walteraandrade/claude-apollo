package db

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

func Open(path string) (*sql.DB, error) {
	// _time_format is a DSN param (not PRAGMA) — requires file: URI scheme to be parsed
	db, err := sql.Open("sqlite", "file:"+path+"?_time_format=sqlite")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)

	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA busy_timeout=5000",
		"PRAGMA foreign_keys=ON",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			db.Close()
			return nil, err
		}
	}

	for _, ddl := range migrations {
		if _, err := db.Exec(ddl); err != nil {
			db.Close()
			return nil, err
		}
	}

	return db, nil
}
