package indexer_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gitlab.com/gitlab-org/es-git-go/git"
	"gitlab.com/gitlab-org/es-git-go/indexer"
)

const (
	sha      = "9876543210987654321098765432109876543210"
	oid      = "0123456789012345678901234567890123456789"
	parentID = "667"
)

type fakeSubmitter struct {
	flushed int

	indexed      int
	indexedID    []string
	indexedThing []interface{}

	removed   int
	removedID []string
}

type fakeRepository struct {
	commits []*git.Commit

	added    []*git.File
	modified []*git.File
	removed  []*git.File
}

func (f *fakeSubmitter) ParentID() string {
	return parentID
}

func (f *fakeSubmitter) Index(id string, thing interface{}) {
	f.indexed++
	f.indexedID = append(f.indexedID, id)
	f.indexedThing = append(f.indexedThing, thing)
}

func (f *fakeSubmitter) Remove(id string) {
	f.removed++
	f.removedID = append(f.removedID, id)
}

func (f *fakeSubmitter) Flush() error {
	f.flushed++
	return nil
}

func (r *fakeRepository) EachFileChange(ins, mod, del git.FileFunc) error {
	for _, file := range r.added {
		if err := ins(file, sha, sha); err != nil {
			return err
		}
	}

	for _, file := range r.modified {
		if err := mod(file, sha, sha); err != nil {
			return err
		}
	}

	for _, file := range r.removed {
		if err := del(file, sha, sha); err != nil {
			return err
		}
	}

	return nil
}

func (r *fakeRepository) EachCommit(f git.CommitFunc) error {
	for _, commit := range r.commits {
		if err := f(commit); err != nil {
			return err
		}
	}

	return nil
}

func setupIndexer() (*indexer.Indexer, *fakeRepository, *fakeSubmitter) {
	repo := &fakeRepository{}
	submitter := &fakeSubmitter{}

	return &indexer.Indexer{
		Repository: repo,
		Submitter:  submitter,
	}, repo, submitter
}

func readerFunc(data string, err error) func() (io.ReadCloser, error) {
	return func() (io.ReadCloser, error) {
		return ioutil.NopCloser(strings.NewReader(data)), err
	}
}

func gitFile(path, content string) *git.File {
	return &git.File{
		Path: path,
		Blob: readerFunc(content, nil),
		Size: int64(len(content)),
		Oid:  oid,
	}
}

func gitCommit(message string) *git.Commit {
	return &git.Commit{
		Author: git.Signature{
			Email: "job@gitlab.com",
			Name:  "Job van der Voort",
			When:  time.Date(2016, time.September, 27, 14, 37, 46, 0, time.UTC),
		},
		Committer: git.Signature{
			Email: "nick@gitlab.com",
			Name:  "Nick Thomas",
			When:  time.Date(2017, time.October, 28, 15, 38, 47, 1, time.UTC),
		},
		Message: message,
		Hash:    sha,
	}
}

func validBlob(file *git.File, content, language string) *indexer.Blob {
	return &indexer.Blob{
		Type:      "blob",
		ID:        indexer.GenerateBlobID(parentID, file.Path),
		OID:       oid,
		RepoID:    parentID,
		CommitSHA: sha,
		Content:   content,
		Path:      file.Path,
		Filename:  file.Path,
		Language:  language,
	}
}

func validCommit(gitCommit *git.Commit) *indexer.Commit {
	return &indexer.Commit{
		Type:      "commit",
		ID:        indexer.GenerateCommitID(parentID, gitCommit.Hash),
		Author:    indexer.BuildPerson(gitCommit.Author),
		Committer: indexer.BuildPerson(gitCommit.Committer),
		RepoID:    parentID,
		Message:   gitCommit.Message,
		SHA:       sha,
	}
}

func TestIndex(t *testing.T) {
	idx, repo, submit := setupIndexer()

	gitCommit := gitCommit("Initial commit")
	gitAdded := gitFile("foo/bar", "added file")
	gitModified := gitFile("foo/baz", "modified file")
	gitRemoved := gitFile("foo/qux", "removed file")

	gitTooBig := gitFile("invalid/too-big", "")
	gitTooBig.Size = int64(1024*1024 + 1)

	gitBinary := gitFile("invalid/binary", "foo\x00")

	commit := validCommit(gitCommit)
	added := validBlob(gitAdded, "added file", "Text")
	modified := validBlob(gitModified, "modified file", "Text")
	removed := validBlob(gitRemoved, "removed file", "Text")

	repo.commits = append(repo.commits, gitCommit)
	repo.added = append(repo.added, gitAdded, gitTooBig, gitBinary)
	repo.modified = append(repo.modified, gitModified)
	repo.removed = append(repo.removed, gitRemoved)

	idx.Index()

	assert.Equal(t, submit.indexed, 3)
	assert.Equal(t, submit.removed, 1)

	assert.Equal(t, parentID+"_"+added.Path, submit.indexedID[0])
	assert.Equal(t, map[string]interface{}{"blob": added}, submit.indexedThing[0])

	assert.Equal(t, parentID+"_"+modified.Path, submit.indexedID[1])
	assert.Equal(t, map[string]interface{}{"blob": modified}, submit.indexedThing[1])

	assert.Equal(t, parentID+"_"+commit.SHA, submit.indexedID[2])
	assert.Equal(t, map[string]interface{}{"commit": commit}, submit.indexedThing[2])

	assert.Equal(t, parentID+"_"+removed.Path, submit.removedID[0])

	assert.Equal(t, submit.flushed, 1)
}

func TestErrorIndexingSkipsRemainder(t *testing.T) {
	idx, repo, submit := setupIndexer()

	gitOKFile := gitFile("ok", "")

	gitBreakingFile := gitFile("broken", "")
	gitBreakingFile.Blob = readerFunc("", fmt.Errorf("Error"))

	repo.added = append(repo.added, gitBreakingFile, gitOKFile)

	err := idx.Index()

	assert.Error(t, err)
	assert.Equal(t, submit.indexed, 0)
	assert.Equal(t, submit.removed, 0)
	assert.Equal(t, submit.flushed, 0)
}
