package git

import (
	"fmt"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/filemode"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/utils/merkletrie"
)

var (
	endError = fmt.Errorf("Finished") // not really an error
)

type goGitRepository struct {
	*git.Repository

	FromHash plumbing.Hash
	ToHash   plumbing.Hash

	FromCommit *object.Commit
	ToCommit   *object.Commit
}

func NewGoGitRepository(projectPath string, fromSHA string, toSHA string) (*goGitRepository, error) {
	out := &goGitRepository{}

	repo, err := git.PlainOpen(projectPath)
	if err != nil {
		return nil, err
	}
	out.Repository = repo

	if fromSHA == "" {
		out.FromHash = plumbing.ZeroHash
	} else {
		out.FromHash = plumbing.NewHash(fromSHA)

		commit, err := repo.CommitObject(out.FromHash)
		if err != nil {
			return nil, fmt.Errorf("Bad from SHA (%s): %s", out.FromHash, err)
		}

		out.FromCommit = commit
	}

	if toSHA == "" {
		ref, err := out.Repository.Head()
		if err != nil {
			return nil, err
		}

		out.ToHash = ref.Hash()
	} else {
		out.ToHash = plumbing.NewHash(toSHA)
	}

	commit, err := out.Repository.CommitObject(out.ToHash)
	if err != nil {
		return nil, fmt.Errorf("Bad to SHA (%s): %s", out.ToHash, err)
	}

	out.ToCommit = commit

	return out, nil
}

func (r *goGitRepository) diff() (object.Changes, error) {
	var fromTree, toTree *object.Tree

	if r.FromCommit != nil {
		tree, err := r.FromCommit.Tree()
		if err != nil {
			return nil, err
		}

		fromTree = tree
	}

	toTree, err := r.ToCommit.Tree()
	if err != nil {
		return nil, err
	}

	return object.DiffTree(fromTree, toTree)
}

func goGitBuildSignature(sig object.Signature) Signature {
	return Signature{
		Name:  sig.Name,
		Email: sig.Email,
		When:  sig.When,
	}
}

func goGitBuildFile(change object.ChangeEntry, file *object.File) *File {
	return &File{
		Path: change.Name,
		Oid:  file.ID().String(),
		Blob: file.Blob.Reader,
		Size: file.Size,
	}
}

func (r *goGitRepository) EachFileChange(ins, mod, del FileFunc) error {
	changes, err := r.diff()
	if err != nil {
		return err
	}

	fromCommitStr := r.FromHash.String()
	toCommitStr := r.ToHash.String()

	for _, change := range changes {
		// FIXME(nick): submodules may need better support
		// https://github.com/src-d/go-git/issues/317
		if change.From.TreeEntry.Mode == filemode.Submodule || change.To.TreeEntry.Mode == filemode.Submodule {
			continue
		}

		fromF, toF, err := change.Files()
		if err != nil {
			return err
		}

		action, err := change.Action()
		if err != nil {
			return err
		}

		switch action {
		case merkletrie.Insert:
			err = ins(goGitBuildFile(change.To, toF), fromCommitStr, toCommitStr)
		case merkletrie.Modify:
			err = mod(goGitBuildFile(change.To, toF), fromCommitStr, toCommitStr)
		case merkletrie.Delete:
			err = del(goGitBuildFile(change.From, fromF), fromCommitStr, toCommitStr)
		default:
			err = fmt.Errorf("Unrecognised change calculating diff: %+v", change)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

// EachCommit runs `f` for each commit within `fromSHA`..`toSHA`
// go-git doesn't directly support ranges of revisions, so we emulate this by
// walking the commit history between from and to
func (r *goGitRepository) EachCommit(f CommitFunc) error {
	err := object.WalkCommitHistoryPost(r.ToCommit, func(c *object.Commit) error {
		if r.FromCommit != nil && c.ID() == r.FromCommit.ID() {
			return endError
		}

		commit := &Commit{
			Message:   c.Message,
			Hash:      c.Hash.String(),
			Author:    goGitBuildSignature(c.Author),
			Committer: goGitBuildSignature(c.Committer),
		}
		if err := f(commit); err != nil {
			return err
		}

		return nil
	})

	if err != nil && err != endError {
		return fmt.Errorf("WalkCommitHistory: %s", err)
	}

	return nil
}
