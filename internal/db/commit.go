package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type CommitRow struct {
	Hash        string
	RepoID      int64
	RepoName    string
	Author      string
	Subject     string
	Body        string
	Branch      string
	CommittedAt time.Time
	DetectedAt  time.Time
	Status      string
	ReviewedAt  *time.Time
	Note        string
}

type ReviewFilter string

const (
	FilterAll        ReviewFilter = "all"
	FilterUnreviewed ReviewFilter = "unreviewed"
	FilterReviewed   ReviewFilter = "reviewed"
	FilterIgnored    ReviewFilter = "ignored"
)

func InsertCommit(db *sql.DB, repoID int64, hash, author, subject, body, branch string, committedAt time.Time) error {
	_, err := db.Exec(
		`INSERT OR IGNORE INTO commits (hash, repo_id, author, subject, body, branch, committed_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		hash, repoID, author, subject, body, branch, committedAt,
	)
	if err != nil {
		return err
	}

	_, err = db.Exec(
		`INSERT OR IGNORE INTO review_state (commit_hash) VALUES (?)`, hash,
	)
	return err
}

func UpdateReviewStatus(db *sql.DB, hash, status, note string) error {
	var reviewedAt *time.Time
	if status == "reviewed" {
		now := time.Now()
		reviewedAt = &now
	}
	_, err := db.Exec(
		`UPDATE review_state SET status = ?, reviewed_at = ?, note = ? WHERE commit_hash = ?`,
		status, reviewedAt, note, hash,
	)
	return err
}

func ListCommits(db *sql.DB, repoID int64, filter ReviewFilter) ([]CommitRow, error) {
	query := `SELECT c.hash, c.repo_id, rp.name, c.author, c.subject, c.body, c.branch,
	                 c.committed_at, c.detected_at,
	                 r.status, r.reviewed_at, r.note
	          FROM commits c
	          JOIN review_state r ON r.commit_hash = c.hash
	          JOIN repositories rp ON rp.id = c.repo_id
	          WHERE c.repo_id = ?`

	if filter != FilterAll {
		query += ` AND r.status = '` + string(filter) + `'`
	}
	query += ` ORDER BY c.committed_at DESC`

	rows, err := db.Query(query, repoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanCommitRows(rows)
}

func ListAllCommits(db *sql.DB, repoIDs []int64, filter ReviewFilter) ([]CommitRow, error) {
	if len(repoIDs) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(repoIDs))
	args := make([]any, len(repoIDs))
	for i, id := range repoIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`SELECT c.hash, c.repo_id, rp.name, c.author, c.subject, c.body, c.branch,
	                 c.committed_at, c.detected_at,
	                 r.status, r.reviewed_at, r.note
	          FROM commits c
	          JOIN review_state r ON r.commit_hash = c.hash
	          JOIN repositories rp ON rp.id = c.repo_id
	          WHERE c.repo_id IN (%s)`, strings.Join(placeholders, ","))

	if filter != FilterAll {
		query += ` AND r.status = '` + string(filter) + `'`
	}
	query += ` ORDER BY c.committed_at DESC`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanCommitRows(rows)
}

func scanCommitRows(rows *sql.Rows) ([]CommitRow, error) {
	var result []CommitRow
	for rows.Next() {
		var c CommitRow
		if err := rows.Scan(&c.Hash, &c.RepoID, &c.RepoName, &c.Author, &c.Subject, &c.Body, &c.Branch,
			&c.CommittedAt, &c.DetectedAt, &c.Status, &c.ReviewedAt, &c.Note); err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	return result, rows.Err()
}

type Stats struct {
	Total      int
	Unreviewed int
	Reviewed   int
	Ignored    int
}

func GetStats(db *sql.DB, repoID int64) (Stats, error) {
	return queryStats(db, `WHERE c.repo_id = ?`, repoID)
}

func GetAggregateStats(db *sql.DB, repoIDs []int64) (Stats, error) {
	if len(repoIDs) == 0 {
		return Stats{}, nil
	}

	placeholders := make([]string, len(repoIDs))
	args := make([]any, len(repoIDs))
	for i, id := range repoIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	return queryStats(db, fmt.Sprintf(`WHERE c.repo_id IN (%s)`, strings.Join(placeholders, ",")), args...)
}

func queryStats(db *sql.DB, where string, args ...any) (Stats, error) {
	var s Stats
	query := fmt.Sprintf(
		`SELECT r.status, COUNT(*)
		 FROM commits c JOIN review_state r ON r.commit_hash = c.hash
		 %s
		 GROUP BY r.status`, where,
	)
	rows, err := db.Query(query, args...)
	if err != nil {
		return s, err
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return s, err
		}
		s.Total += count
		switch status {
		case "unreviewed":
			s.Unreviewed = count
		case "reviewed":
			s.Reviewed = count
		case "ignored":
			s.Ignored = count
		}
	}
	return s, rows.Err()
}
