package git_test

import (
	"flag"
	"io/ioutil"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gitlab.com/gitlab-org/gitlab-elasticsearch-indexer/git"
)

var (
	testRepo = flag.String("test-repo", "../tmp/gitlab-test.git", "Path to `gitlab-test` repository for integration tests")
)

const (
	initialSHA = "1a0b36b3cdad1d2ee32457c102a8c0b7056fa863"
	headSHA    = "b83d6e391c22777fca1ed3012fce84f633d7fed0"
)

func checkDeps(t *testing.T) {
	if _, err := os.Stat(*testRepo); err != nil {
		t.Log("No test repo found at ", *testRepo)
		t.SkipNow()
	}
}

func runEachCommit(repo git.Repository) (map[string]*git.Commit, []string, error) {
	commits := make(map[string]*git.Commit)
	commitHashes := []string{}

	err := repo.EachCommit(func(commit *git.Commit) error {
		commits[commit.Hash] = commit
		commitHashes = append(commitHashes, commit.Hash)
		return nil
	})

	return commits, commitHashes, err
}

func TestEachCommit(t *testing.T) {
	checkDeps(t)

	repo, err := git.NewGoGitRepository(*testRepo, "", headSHA)
	assert.NoError(t, err)

	commits, commitHashes, err := runEachCommit(repo)
	assert.NoError(t, err)

	expectedCommits := []string{
		"b83d6e391c22777fca1ed3012fce84f633d7fed0",
		"498214de67004b1da3d820901307bed2a68a8ef6",
		"1b12f15a11fc6e62177bef08f47bc7b5ce50b141",
		"38008cb17ce1466d8fec2dfa6f6ab8dcfe5cf49e",
		"6907208d755b60ebeacb2e9dfea74c92c3449a1f",
		"c347ca2e140aa667b968e51ed0ffe055501fe4f4",
		"d59c60028b053793cecfb4022de34602e1a9218e",
		"281d3a76f31c812dbf48abce82ccf6860adedd81",
		"a5391128b0ef5d21df5dd23d98557f4ef12fae20",
		"54fcc214b94e78d7a41a9a8fe6d87a5e59500e51",
		"be93687618e4b132087f430a4d8fc3a609c9b77c",
		"048721d90c449b244b7b4c53a9186b04330174ec",
		"5f923865dde3436854e9ceb9cdb7815618d4e849",
		"d2d430676773caa88cdaf7c55944073b2fd5561a",
		"2ea1f3dec713d940208fb5ce4a38765ecb5d3f73",
		"59e29889be61e6e0e5e223bfa9ac2721d31605b8",
		"66eceea0db202bb39c4e445e8ca28689645366c5",
		"08f22f255f082689c0d7d39d19205085311542bc",
		"19e2e9b4ef76b422ce1154af39a91323ccc57434",
		"c642fe9b8b9f28f9225d7ea953fe14e74748d53b",
		"9a944d90955aaf45f6d0c88f30e27f8d2c41cec0",
		"c7fbe50c7c7419d9701eebe64b1fdacc3df5b9dd",
		"e56497bb5f03a90a51293fc6d516788730953899",
		"4cd80ccab63c82b4bad16faa5193fbd2aa06df40",
		"5937ac0a7beb003549fc5fd26fc247adbce4a52e",
		"570e7b2abdd848b95f2f578043fc23bd6f6fd24d",
		"6f6d7e7ed97bb5f0054f2b1df789b39ca89b6ff9",
		"d14d6c0abdd253381df51a723d58691b2ee1ab08",
		"c1acaa58bbcbc3eafe538cb8274ba387047b69f8",
		"ae73cb07c9eeaf35924a10f713b364d32b2dd34f",
		"874797c3a73b60d2187ed6e2fcabd289ff75171e",
		"2f63565e7aac07bcdadb654e253078b727143ec4",
		"33f3729a45c02fc67d00adb1b8bca394b0e761d9",
		"913c66a37b4a45b9769037c55c2d238bd0942d2e",
		"cfe32cf61b73a0d5e9f13e774abde7ff789b1660",
		"6d394385cf567f80a8fd85055db1ab4c5295806f",
		"1a0b36b3cdad1d2ee32457c102a8c0b7056fa863",
	}

	// We don't mind the order these are given in
	sort.Strings(expectedCommits)
	sort.Strings(commitHashes)

	assert.Equal(t, expectedCommits, commitHashes)

	// Now choose one commit and check it in detail

	commit := commits[initialSHA]
	date, err := time.Parse("Mon Jan 02 15:04:05 2006 -0700", "Thu Feb 27 00:03:18 2014 -0800")
	assert.NoError(t, err)

	dmitriy := git.Signature{
		Name:  "Dmitriy Zaporozhets",
		Email: "dmitriy.zaporozhets@gmail.com",
		When:  date,
	}

	assert.Equal(t, initialSHA, commit.Hash)
	assert.Equal(t, "Initial commit\n", commit.Message)
	assert.Equal(t, dmitriy, commit.Author)
	assert.Equal(t, dmitriy, commit.Author)
}

func TestEachCommitGivenRangeOf3Commits(t *testing.T) {
	checkDeps(t)

	repo, err := git.NewGoGitRepository(*testRepo, "1b12f15a11fc6e62177bef08f47bc7b5ce50b141", headSHA)
	assert.NoError(t, err)

	_, commitHashes, err := runEachCommit(repo)
	assert.NoError(t, err)

	expected := []string{"498214de67004b1da3d820901307bed2a68a8ef6", headSHA}
	sort.Strings(expected)
	sort.Strings(commitHashes)

	assert.Equal(t, expected, commitHashes)
}

func TestEachCommitGivenRangeOf2Commits(t *testing.T) {
	checkDeps(t)

	repo, err := git.NewGoGitRepository(*testRepo, "498214de67004b1da3d820901307bed2a68a8ef6", headSHA)
	assert.NoError(t, err)

	_, commitHashes, err := runEachCommit(repo)
	assert.NoError(t, err)

	assert.Equal(t, []string{headSHA}, commitHashes)
}

func TestEachCommitGivenRangeOf1Commit(t *testing.T) {
	checkDeps(t)

	repo, err := git.NewGoGitRepository(*testRepo, headSHA, headSHA)
	assert.NoError(t, err)

	_, commitHashes, err := runEachCommit(repo)
	assert.NoError(t, err)
	assert.Equal(t, []string{}, commitHashes)
}

func TestEmptyToSHADefaultsToHeadSHA(t *testing.T) {
	checkDeps(t)

	repo, err := git.NewGoGitRepository(*testRepo, "498214de67004b1da3d820901307bed2a68a8ef6", "")
	assert.NoError(t, err)

	_, commitHashes, err := runEachCommit(repo)
	assert.NoError(t, err)
	assert.Equal(t, []string{headSHA}, commitHashes)
}

// TODO: store the ins/mod/del status of each change
func runEachFileChange(repo git.Repository) (map[string]*git.File, []string, error) {
	files := make(map[string]*git.File)
	filePaths := []string{}

	store := func(f *git.File, _, _ string) error {
		files[f.Path] = f
		filePaths = append(filePaths, f.Path)
		return nil
	}

	err := repo.EachFileChange(store, store, store)
	return files, filePaths, err
}

func TestEachFileChangeMod(t *testing.T) {
	checkDeps(t)

	repo, err := git.NewGoGitRepository(*testRepo, "", headSHA)
	assert.NoError(t, err)

	files, filePaths, err := runEachFileChange(repo)
	assert.NoError(t, err)

	expectedFiles := []string{
		".gitattributes",
		".gitignore",
		".gitmodules",
		"CHANGELOG",
		"CONTRIBUTING.md",
		"Gemfile.zip",
		"LICENSE",
		"MAINTENANCE.md",
		"PROCESS.md",
		"README",
		"README.md",
		"VERSION",
		"bar/branch-test.txt",
		"custom-highlighting/test.gitlab-custom",
		"encoding/feature-1.txt",
		"encoding/feature-2.txt",
		"encoding/hotfix-1.txt",
		"encoding/hotfix-2.txt",
		"encoding/iso8859.txt",
		"encoding/russian.rb",
		"encoding/test.txt",
		"encoding/テスト.txt",
		"encoding/テスト.xls",
		"files/html/500.html",
		"files/images/6049019_460s.jpg",
		"files/images/logo-black.png",
		"files/images/logo-white.png",
		"files/images/wm.svg",
		"files/js/application.js",
		"files/js/commit.coffee",
		"files/lfs/lfs_object.iso",
		"files/markdown/ruby-style-guide.md",
		"files/ruby/popen.rb",
		"files/ruby/regex.rb",
		"files/ruby/version_info.rb",
		"files/whitespace",
		"foo/bar/.gitkeep",
		"with space/README.md",
	}

	// We don't mind the order these are given in
	sort.Strings(expectedFiles)
	sort.Strings(filePaths)

	assert.Equal(t, expectedFiles, filePaths)

	// Now choose one file and check it in detail
	file := files["VERSION"]
	blob, err := file.Blob()
	assert.NoError(t, err)
	data, err := ioutil.ReadAll(blob)
	assert.NoError(t, err)

	assert.Equal(t, "VERSION", file.Path)
	assert.Equal(t, "998707b421c89bd9a3063333f9f728ef3e43d101", file.Oid)
	assert.Equal(t, int64(10), file.Size)
	assert.Equal(t, "6.7.0.pre\n", string(data))
}

func TestEachFileChangeGivenRangeOfThreeCommits(t *testing.T) {
	checkDeps(t)

	repo, err := git.NewGoGitRepository(*testRepo, "1b12f15a11fc6e62177bef08f47bc7b5ce50b141", headSHA)
	assert.NoError(t, err)

	_, filePaths, err := runEachFileChange(repo)

	assert.Equal(t, []string{"bar/branch-test.txt"}, filePaths)
}

func TestEachFileChangeGivenRangeOfTwoCommits(t *testing.T) {
	checkDeps(t)

	repo, err := git.NewGoGitRepository(*testRepo, "498214de67004b1da3d820901307bed2a68a8ef6", headSHA)
	assert.NoError(t, err)

	_, filePaths, err := runEachFileChange(repo)

	assert.Equal(t, []string{}, filePaths)
}
