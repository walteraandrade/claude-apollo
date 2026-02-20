package git

import (
	"errors"
	"fmt"
	"strings"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

var errStop = errors.New("stop")

type Repo struct {
	repo *gogit.Repository
	path string
}

func OpenRepo(path string) (*Repo, error) {
	r, err := gogit.PlainOpen(path)
	if err != nil {
		return nil, fmt.Errorf("open repo %s: %w", path, err)
	}
	return &Repo{repo: r, path: path}, nil
}

func (r *Repo) CurrentBranch() string {
	ref, err := r.repo.Head()
	if err != nil {
		return "unknown"
	}
	name := ref.Name().String()
	if strings.HasPrefix(name, "refs/heads/") {
		return strings.TrimPrefix(name, "refs/heads/")
	}
	return name
}

func (r *Repo) ReadNewCommits(sinceHash string, limit int) ([]CommitInfo, error) {
	ref, err := r.repo.Head()
	if err != nil {
		return nil, fmt.Errorf("head: %w", err)
	}

	branch := r.CurrentBranch()

	iter, err := r.repo.Log(&gogit.LogOptions{
		From:  ref.Hash(),
		Order: gogit.LogOrderCommitterTime,
	})
	if err != nil {
		return nil, fmt.Errorf("log: %w", err)
	}
	defer iter.Close()

	var commits []CommitInfo
	err = iter.ForEach(func(c *object.Commit) error {
		if c.Hash.String() == sinceHash {
			return errStop
		}
		if len(commits) >= limit {
			return errStop
		}

		msg := strings.TrimSpace(c.Message)
		subject, body := splitMessage(msg)

		parents := make([]string, 0, c.NumParents())
		c.Parents().ForEach(func(p *object.Commit) error {
			parents = append(parents, p.Hash.String())
			return nil
		})

		commits = append(commits, CommitInfo{
			Hash:      c.Hash.String(),
			Author:    c.Author.Name,
			Subject:   subject,
			Body:      body,
			Branch:    branch,
			Timestamp: c.Author.When,
			Parents:   parents,
		})
		return nil
	})

	if err != nil && !errors.Is(err, errStop) {
		return nil, err
	}

	reverse(commits)
	return commits, nil
}

func (r *Repo) SeedCommits(n int) ([]CommitInfo, error) {
	return r.ReadNewCommits("", n)
}

func splitMessage(msg string) (subject, body string) {
	parts := strings.SplitN(msg, "\n", 2)
	subject = strings.TrimSpace(parts[0])
	if len(parts) > 1 {
		body = strings.TrimSpace(parts[1])
	}
	return
}

func reverse(s []CommitInfo) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}
