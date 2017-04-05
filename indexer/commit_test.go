package indexer_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/gitlab-org/es-git-go/indexer"
)

func TestBuildCommit(t *testing.T) {
	gitCommit := gitCommit("Initial commit")

	expected := validCommit(gitCommit)
	actual := indexer.BuildCommit(gitCommit, expected.RepoID)

	assert.Equal(t, expected, actual)

	expectedJSON := `{
		"sha"       : "` + expected.SHA + `",
		"message"   : "` + expected.Message + `",
		"author"    : {
			"name": "` + expected.Author.Name + `",
			"email": "` + expected.Author.Email + `",
			"time": "` + indexer.GenerateDate(gitCommit.Author.When) + `"
		},
		"committer" : {
			"name": "` + expected.Committer.Name + `",
			"email": "` + expected.Committer.Email + `",
			"time": "` + indexer.GenerateDate(gitCommit.Committer.When) + `"
		},
		"rid"       : "` + expected.RepoID + `",
		"type"      : "commit"
	}`

	actualJSON, err := json.Marshal(actual)
	assert.NoError(t, err)
	assert.JSONEq(t, expectedJSON, string(actualJSON))
}

func TestGenerateCommitID(t *testing.T) {
	assert.Equal(t, "projectID_sha", indexer.GenerateCommitID("projectID", "sha"))
}
