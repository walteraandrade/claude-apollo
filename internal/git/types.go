package git

import "time"

type CommitInfo struct {
	Hash      string
	Author    string
	Subject   string
	Body      string
	Branch    string
	Timestamp time.Time
	Parents   []string
}
