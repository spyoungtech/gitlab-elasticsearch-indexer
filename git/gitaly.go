package git

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"

	gitalyauth "gitlab.com/gitlab-org/gitaly/auth"
	gitalyclient "gitlab.com/gitlab-org/gitaly/client"
	pb "gitlab.com/gitlab-org/gitaly/proto/go/gitalypb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const SubmoduleFileMode = 0160000
const LimitFileSize = 1024 * 1024

// See https://stackoverflow.com/questions/9765453/is-gits-semi-secret-empty-tree-object-reliable-and-why-is-there-not-a-symbolic
const NullTreeSHA = "4b825dc642cb6eb9a060e54bf8d69288fbee4904"
const ZeroSHA = "0000000000000000000000000000000000000000"

type StorageConfig struct {
	Address      string `json:"address"`
	Token        string `json:"token"`
	StorageName  string `json:"storage"`
	RelativePath string `json:"relative_path"`
	TokenVersion int    `json:"token_version"`
}

type gitalyClient struct {
	conn                    *grpc.ClientConn
	repository              *pb.Repository
	blobServiceClient       pb.BlobServiceClient
	repositoryServiceClient pb.RepositoryServiceClient
	refServiceClient        pb.RefServiceClient
	commitServiceClient     pb.CommitServiceClient

	FromHash string
	ToHash   string
}

func NewGitalyClient(config *StorageConfig, fromSHA, toSHA string) (*gitalyClient, error) {
	var RPCCred credentials.PerRPCCredentials
	if config.TokenVersion == 0 || config.TokenVersion == 2 {
		RPCCred = gitalyauth.RPCCredentialsV2(config.Token)
	} else {
		return nil, errors.New("Unknown token version")
	}

	connOpts := append(
		gitalyclient.DefaultDialOpts,
		grpc.WithPerRPCCredentials(RPCCred),
	)

	conn, err := gitalyclient.Dial(config.Address, connOpts)
	if err != nil {
		return nil, fmt.Errorf("did not connect: %s", err)
	}

	repository := &pb.Repository{
		StorageName:  config.StorageName,
		RelativePath: config.RelativePath,
	}

	client := &gitalyClient{
		conn:                    conn,
		repository:              repository,
		blobServiceClient:       pb.NewBlobServiceClient(conn),
		repositoryServiceClient: pb.NewRepositoryServiceClient(conn),
		refServiceClient:        pb.NewRefServiceClient(conn),
		commitServiceClient:     pb.NewCommitServiceClient(conn),
	}

	if fromSHA == "" || fromSHA == ZeroSHA {
		client.FromHash = NullTreeSHA
	} else {
		client.FromHash = fromSHA
	}

	if toSHA == "" {
		head, err := client.lookUpHEAD()
		if err != nil {
			return nil, fmt.Errorf("lookUpHEAD: %v", err)
		}
		client.ToHash = head
	} else {
		client.ToHash = toSHA
	}

	return client, nil
}

func NewGitalyClientFromEnv(projectPath, fromSHA, toSHA string) (*gitalyClient, error) {
	data := strings.NewReader(os.Getenv("GITALY_CONNECTION_INFO"))

	config := StorageConfig{RelativePath: projectPath}

	if err := json.NewDecoder(data).Decode(&config); err != nil {
		return nil, err
	}

	client, err := NewGitalyClient(&config, fromSHA, toSHA)
	if err != nil {
		return nil, fmt.Errorf("Failed to open %s: %s", config.RelativePath, err)
	}

	return client, nil
}

func (gc *gitalyClient) Close() {
	gc.conn.Close()
}

func (gc *gitalyClient) EachFileChange(put, del FileFunc) error {
	request := &pb.GetRawChangesRequest{
		Repository:   gc.repository,
		FromRevision: gc.FromHash,
		ToRevision:   gc.ToHash,
	}

	stream, err := gc.repositoryServiceClient.GetRawChanges(context.Background(), request)
	if err != nil {
		return fmt.Errorf("could not call rpc.GetRawChanges: %v", err)
	}

	for {
		c, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("%v.GetRawChanges, %v", c, err)
		}
		for _, change := range c.RawChanges {
			// TODO: We just skip submodules from indexing now just to mirror the go-git
			// implementation but it can be not that expensive to implement with gitaly actually so some
			// investigation is required here
			if change.OldMode == SubmoduleFileMode || change.NewMode == SubmoduleFileMode {
				continue
			}

			switch change.Operation.String() {
			case "DELETED", "RENAMED":
				file, err := gc.gitalyBuildFile(change, string(change.OldPath), true)
				if err != nil {
					return err
				}
				log.Debug("Indexing blob change: ", "DELETE", file.Path)
				if err = del(file, gc.FromHash, gc.ToHash); err != nil {
					return err
				}
			}

			switch change.Operation.String() {
			case "ADDED", "RENAMED", "MODIFIED", "COPIED":
				file, err := gc.gitalyBuildFile(change, string(change.NewPath), false)
				if err != nil {
					return err
				}
				log.Debug("Indexing blob change: ", "PUT", file.Path)
				if err = put(file, gc.FromHash, gc.ToHash); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// HEAD is not always set in some cases, so we find the last commit in
// a default branch instead
func (gc *gitalyClient) lookUpHEAD() (string, error) {
	defaultBranchName, err := gc.findDefaultBranchName()
	if err != nil {
		return "", err
	}

	request := &pb.FindCommitRequest{
		Repository: gc.repository,
		Revision:   defaultBranchName,
	}

	response, err := gc.commitServiceClient.FindCommit(context.Background(), request)
	if err != nil {
		return "", fmt.Errorf("Cannot look up HEAD: %v", err)
	}
	return response.Commit.Id, nil
}

func (gc *gitalyClient) findDefaultBranchName() ([]byte, error) {
	request := &pb.FindDefaultBranchNameRequest{
		Repository: gc.repository,
	}

	response, err := gc.refServiceClient.FindDefaultBranchName(context.Background(), request)
	if err != nil {
		return nil, fmt.Errorf("Cannot find a default branch: %v", err)
	}
	return response.Name, nil
}

func (gc *gitalyClient) getBlob(oid string) (io.ReadCloser, error) {
	data := new(bytes.Buffer)

	request := &pb.GetBlobRequest{
		Repository: gc.repository,
		Oid:        oid,
		Limit:      LimitFileSize,
	}

	stream, err := gc.blobServiceClient.GetBlob(context.Background(), request)
	if err != nil {
		return nil, fmt.Errorf("Cannot get blob: %s", oid)
	}

	for {
		c, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("%v.GetBlob: %v", c, err)
		}
		if c.Data != nil {
			data.Write(c.Data)
		}
	}

	return ioutil.NopCloser(data), nil
}

func (gc *gitalyClient) gitalyBuildFile(change *pb.GetRawChangesResponse_RawChange, path string, withoutBlobReader bool) (*File, error) {
	var data io.ReadCloser
	// We limit the size to avoid loading too big blobs into memory
	// as they will be rejected on the indexer side anyway
	// Ideally, we need to create a lazy blob reader here.
	if withoutBlobReader || change.Size > LimitFileSize {
		data = ioutil.NopCloser(new(bytes.Buffer))
	} else {
		var err error
		data, err = gc.getBlob(change.BlobId)
		if err != nil {
			return nil, fmt.Errorf("getBlob returns error: %v", err)
		}
	}

	return &File{
		Path: path,
		Oid:  change.BlobId,
		Blob: getBlobReader(data),
		Size: change.Size,
	}, nil
}

func getBlobReader(data io.ReadCloser) func() (io.ReadCloser, error) {
	return func() (io.ReadCloser, error) { return data, nil }
}

func (gc *gitalyClient) EachCommit(f CommitFunc) error {
	request := &pb.CommitsBetweenRequest{
		Repository: gc.repository,
		From:       []byte(gc.FromHash),
		To:         []byte(gc.ToHash),
	}

	stream, err := gc.commitServiceClient.CommitsBetween(context.Background(), request)
	if err != nil {
		return fmt.Errorf("could not call rpc.CommitsBetween: %v", err)
	}

	for {
		c, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error calling rpc.CommitsBetween: %v", err)
		}
		for _, cmt := range c.Commits {
			commit := &Commit{
				Message:   string(cmt.Body),
				Hash:      string(cmt.Id),
				Author:    gitalyBuildSignature(cmt.Author),
				Committer: gitalyBuildSignature(cmt.Committer),
			}

			log.Debug("Indexing commit: ", cmt.Id)

			if err := f(commit); err != nil {
				return err
			}
		}
	}
	return nil
}

func gitalyBuildSignature(ca *pb.CommitAuthor) Signature {
	return Signature{
		Name:  string(ca.Name),
		Email: string(ca.Email),
		When:  time.Unix(ca.Date.GetSeconds(), 0), // another option is ptypes.Timestamp(ca.Date)
	}
}
