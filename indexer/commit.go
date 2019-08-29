package indexer

import (
	"fmt"
	"strconv"

	"gitlab.com/gitlab-org/gitlab-elasticsearch-indexer/git"
)

type Commit struct {
	Type      string  `json:"type"`
	ID        string  `json:"-"`
	Author    *Person `json:"author"`
	Committer *Person `json:"committer"`
	RepoID    string  `json:"rid"`
	Message   string  `json:"message"`
	SHA       string  `json:"sha"`
}

func GenerateCommitID(parentID int64, commitSHA string) string {
	return fmt.Sprintf("%v_%s", parentID, commitSHA)
}

func BuildCommit(c *git.Commit, parentID int64) *Commit {
	sha := c.Hash

	return &Commit{
		Type:      "commit",
		Author:    BuildPerson(c.Author),
		Committer: BuildPerson(c.Committer),
		ID:        GenerateCommitID(parentID, sha),
		RepoID:    strconv.FormatInt(parentID, 10),
		Message:   tryEncodeString(c.Message),
		SHA:       sha,
	}
}
