package db

import "database/sql"

type Repository struct {
	ID             int64
	Name           string
	Path           string
	Active         bool
	LastCommitHash string
}

func UpsertRepo(db *sql.DB, name, path string) (int64, error) {
	res, err := db.Exec(
		`INSERT INTO repositories (name, path) VALUES (?, ?)
		 ON CONFLICT(path) DO UPDATE SET name=excluded.name`,
		name, path,
	)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil || id == 0 {
		row := db.QueryRow(`SELECT id FROM repositories WHERE path = ?`, path)
		if err := row.Scan(&id); err != nil {
			return 0, err
		}
	}
	return id, nil
}

func GetRepoByPath(db *sql.DB, path string) (*Repository, error) {
	r := &Repository{}
	err := db.QueryRow(
		`SELECT id, name, path, active, last_commit_hash FROM repositories WHERE path = ?`, path,
	).Scan(&r.ID, &r.Name, &r.Path, &r.Active, &r.LastCommitHash)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return r, nil
}

func UpdateLastCommitHash(db *sql.DB, repoID int64, hash string) error {
	_, err := db.Exec(`UPDATE repositories SET last_commit_hash = ? WHERE id = ?`, hash, repoID)
	return err
}
