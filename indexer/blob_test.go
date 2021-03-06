package indexer_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"gitlab.com/gitlab-org/gitlab-elasticsearch-indexer/indexer"
)

func TestBuildBlob(t *testing.T) {
	file := gitFile("foo/bar", "foo")
	expected := validBlob(file, "foo", "Text")

	actual, err := indexer.BuildBlob(file, parentID, expected.CommitSHA, "blob")
	require.NoError(t, err)

	require.Equal(t, expected, actual)

	expectedJSON := `{
		"commit_sha" : "` + expected.CommitSHA + `",
		"content"    : "` + expected.Content + `",
		"file_name"  : "` + expected.Filename + `",
		"language"   : "` + expected.Language + `",
		"oid"        : "` + expected.OID + `",
		"path"       : "` + expected.Path + `",
		"rid"        : "` + expected.RepoID + `",
		"type"       : "blob"
	}`

	actualJSON, err := json.Marshal(actual)
	require.NoError(t, err)
	require.JSONEq(t, expectedJSON, string(actualJSON))
}

func TestBuildBlobSkipsLargeBlobs(t *testing.T) {
	file := gitFile("foo/bar", "foo")
	file.Size = 1024*1024 + 1

	blob, err := indexer.BuildBlob(file, parentID, sha, "blob")
	require.Error(t, err, indexer.SkipTooLargeBlob)
	require.Nil(t, blob)
}

func TestBuildBlobSkipsBinaryBlobs(t *testing.T) {
	file := gitFile("foo/bar", "foo\x00")

	blob, err := indexer.BuildBlob(file, parentID, sha, "blob")
	require.Equal(t, err, indexer.SkipBinaryBlob)
	require.Nil(t, blob)
}

func TestBuildBlobDetectsLanguageByFilename(t *testing.T) {
	file := gitFile("Makefile.am", "foo")
	blob, err := indexer.BuildBlob(file, parentID, sha, "blob")

	require.NoError(t, err)
	require.Equal(t, "Makefile", blob.Language)
}

func TestBuildBlobDetectsLanguageByExtension(t *testing.T) {
	file := gitFile("foo.rb", "foo")
	blob, err := indexer.BuildBlob(file, parentID, sha, "blob")

	require.NoError(t, err)
	require.Equal(t, "Ruby", blob.Language)
}

func TestGenerateBlobID(t *testing.T) {
	require.Equal(t, "2147483648_path", indexer.GenerateBlobID(2147483648, "path"))
}
