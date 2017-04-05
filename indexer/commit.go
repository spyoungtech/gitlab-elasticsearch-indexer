package indexer

import (
	"fmt"

	"gitlab.com/gitlab-org/es-git-go/git"
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

func GenerateCommitID(parentID, commitSHA string) string {
	return fmt.Sprintf("%s_%s", parentID, commitSHA)
}

func BuildCommit(c *git.Commit, parentID string) *Commit {
	sha := c.Hash

	return &Commit{
		Type:      "commit",
		Author:    BuildPerson(c.Author),
		Committer: BuildPerson(c.Committer),
		ID:        GenerateCommitID(parentID, sha),
		RepoID:    parentID,
		Message:   tryEncodeString(c.Message),
		SHA:       sha,
	}
}
