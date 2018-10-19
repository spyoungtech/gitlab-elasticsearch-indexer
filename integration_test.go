package main_test

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gitlab.com/gitlab-org/gitlab-elasticsearch-indexer/elastic"
	"gitlab.com/gitlab-org/gitlab-elasticsearch-indexer/indexer"
)

var (
	binary   = flag.String("binary", "./bin/gitlab-elasticsearch-indexer", "Path to `gitlab-elasticsearch-indexer` binary for integration tests")
	testRepo = flag.String("test-repo", "./tmp/gitlab-test.git", "Path to `gitlab-test` repository for integration tests")
)

const (
	projectID = "667"
	headSHA   = "b83d6e391c22777fca1ed3012fce84f633d7fed0"
)

func checkDeps(t *testing.T) {
	if os.Getenv("ELASTIC_CONNECTION_INFO") == "" {
		t.Log("ELASTIC_CONNECTION_INFO not set")
		t.Skip()
	}

	if testing.Short() {
		t.Log("Test run with -short, skipping integration test")
		t.Skip()
	}

	if _, err := os.Stat(*binary); err != nil {
		t.Log("No binary found at ", *binary)
		t.Skip()
	}

	if _, err := os.Stat(*testRepo); err != nil {
		t.Log("No test repo found at ", *testRepo)
		t.Skip()
	}
}

func buildIndex(t *testing.T) (*elastic.Client, func()) {
	railsEnv := fmt.Sprintf("test-integration-%d", time.Now().Unix())
	os.Setenv("RAILS_ENV", railsEnv)

	client, err := elastic.FromEnv(projectID)
	assert.NoError(t, err)

	assert.NoError(t, client.CreateIndex())

	return client, func() {
		client.DeleteIndex()
	}
}

func run(from, to string) error {
	cmd := exec.Command(*binary, projectID, *testRepo)
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// GitLab always sets FROM_SHA
	if from == "" {
		from = "0000000000000000000000000000000000000000"
	}

	cmd.Env = append(cmd.Env, "FROM_SHA="+from)

	if to != "" {
		cmd.Env = append(cmd.Env, "TO_SHA="+to)
	}

	return cmd.Run()
}

func TestIndexingRemovesFiles(t *testing.T) {
	checkDeps(t)
	c, td := buildIndex(t)
	defer td()

	// The commit before files/empty is removed - so it should be indexed
	assert.NoError(t, run("", "19e2e9b4ef76b422ce1154af39a91323ccc57434"))
	_, err := c.GetBlob("files/empty")
	assert.NoError(t, err)

	// Now we expect it to have been removed
	assert.NoError(t, run("19e2e9b4ef76b422ce1154af39a91323ccc57434", "08f22f255f082689c0d7d39d19205085311542bc"))
	_, err = c.GetBlob("files/empty")
	assert.Error(t, err)
}

type document struct {
	Blob      *indexer.Blob     `json:"blob"`
	Commit    *indexer.Commit   `json:"commit"`
	Type      string            `json:"string"`
	JoinField map[string]string `json:"join_field"`
}

// Go source is defined to be UTF-8 encoded, so literals here are UTF-8
func TestIndexingTranscodesToUTF8(t *testing.T) {
	checkDeps(t)
	c, td := buildIndex(t)
	defer td()

	assert.NoError(t, run("", headSHA))

	for _, tc := range []struct {
		path     string
		expected string
	}{
		{"encoding/iso8859.txt", "狞\n"},                                                         // GB18030
		{"encoding/test.txt", "これはテストです。\nこれもマージして下さい。\n\nAdd excel file.\nDelete excel file."}, // SHIFT_JIS
	} {

		blob, err := c.GetBlob(tc.path)
		assert.NoError(t, err)

		blobDoc := &document{}
		assert.NoError(t, json.Unmarshal(*blob.Source, &blobDoc))

		assert.Equal(t, tc.expected, blobDoc.Blob.Content)
	}
}

func TestIndexingGitlabTest(t *testing.T) {
	checkDeps(t)
	c, td := buildIndex(t)
	defer td()

	assert.NoError(t, run("", headSHA))

	// Check the indexing of a commit
	commit, err := c.GetCommit(headSHA)
	assert.NoError(t, err)
	assert.True(t, commit.Found)
	assert.Equal(t, "doc", commit.Type)
	assert.Equal(t, projectID+"_"+headSHA, commit.Id)
	assert.Equal(t, "project_"+projectID, commit.Routing)

	data := make(map[string]interface{})
	assert.NoError(t, json.Unmarshal(*commit.Source, &data))

	commitDoc, ok := data["commit"]
	assert.True(t, ok)
	assert.Equal(
		t,
		map[string]interface{}{
			"type": "commit",
			"sha":  headSHA,
			"author": map[string]interface{}{
				"email": "job@gitlab.com",
				"name":  "Job van der Voort",
				"time":  "20160927T143746+0000",
			},
			"committer": map[string]interface{}{
				"email": "job@gitlab.com",
				"name":  "Job van der Voort",
				"time":  "20160927T143746+0000",
			},
			"rid":     projectID,
			"message": "Merge branch 'branch-merged' into 'master'\r\n\r\nadds bar folder and branch-test text file to check Repository merged_to_root_ref method\r\n\r\n\r\n\r\nSee merge request !12",
		},
		commitDoc,
	)

	// Check the indexing of a text blob
	blob, err := c.GetBlob("README.md")
	assert.NoError(t, err)
	assert.True(t, blob.Found)
	assert.Equal(t, "doc", blob.Type)
	assert.Equal(t, projectID+"_README.md", blob.Id)
	assert.Equal(t, "project_"+projectID, blob.Routing)

	data = make(map[string]interface{})
	assert.NoError(t, json.Unmarshal(*blob.Source, &data))

	blobDoc, ok := data["blob"]
	assert.True(t, ok)
	assert.Equal(
		t,
		map[string]interface{}{
			"type":       "blob",
			"language":   "Markdown",
			"path":       "README.md",
			"file_name":  "README.md",
			"oid":        "faaf198af3a36dbf41961466703cc1d47c61d051",
			"rid":        projectID,
			"commit_sha": headSHA,
			"content":    "testme\n======\n\nSample repo for testing gitlab features\n",
		},
		blobDoc,
	)

	// Check that a binary blob isn't indexed
	_, err = c.GetBlob("Gemfile.zip")
	assert.Error(t, err)

	// Test that timezones are preserved
	commit, err = c.GetCommit("498214de67004b1da3d820901307bed2a68a8ef6")
	assert.NoError(t, err)

	cDoc := &document{}
	assert.NoError(t, json.Unmarshal(*commit.Source, &cDoc))
	assert.Equal(t, "20160921T161326+0100", cDoc.Commit.Author.Time)
	assert.Equal(t, "20160921T161326+0100", cDoc.Commit.Committer.Time)
}
