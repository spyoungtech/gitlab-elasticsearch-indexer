package main

import (
	"flag"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
	"gitlab.com/gitlab-org/gitlab-elasticsearch-indexer/elastic"
	"gitlab.com/gitlab-org/gitlab-elasticsearch-indexer/git"
	"gitlab.com/gitlab-org/gitlab-elasticsearch-indexer/indexer"
)

var (
	versionFlag     = flag.Bool("version", false, "Print the version and exit")
	skipCommitsFlag = flag.Bool("skip-commits", false, "Skips indexing commits for the repo")
	blobTypeFlag    = flag.String("blob-type", "blob", "The type of blobs to index. Accepted values: 'blob', 'wiki_blob'")

	// Overriden in the makefile
	Version   = "dev"
	BuildTime = ""
)

func main() {
	flag.Parse()

	if *versionFlag {
		log.Printf("%s %s (built at: %s)", os.Args[0], Version, BuildTime)
		os.Exit(0)
	}

	configureLogger()
	args := flag.Args()

	if len(args) != 2 {
		log.Fatalf("Usage: %s [ --version | [--blob-type=(blob|wiki_blob)] [--skip-comits] <project-id> <project-path> ]", os.Args[0])
	}

	projectID, err := strconv.ParseInt(args[0], 10, 64)

	if err != nil {
		log.Fatal(err)
	}

	projectPath := args[1]
	fromSHA := os.Getenv("FROM_SHA")
	toSHA := os.Getenv("TO_SHA")
	blobType := *blobTypeFlag
	skipCommits := *skipCommitsFlag

	repo, err := git.NewGitalyClientFromEnv(projectPath, fromSHA, toSHA)
	if err != nil {
		log.Fatal(err)
	}

	esClient, err := elastic.FromEnv(projectID)
	if err != nil {
		log.Fatal(err)
	}

	idx := &indexer.Indexer{
		Submitter:  esClient,
		Repository: repo,
	}

	log.Debugf("Indexing from %s to %s", repo.FromHash, repo.ToHash)
	log.Debugf("Index: %s, Project ID: %v, blob_type: %s, skip_commits?: %t", esClient.IndexName, esClient.ParentID(), blobType, skipCommits)

	if err := idx.IndexBlobs(blobType); err != nil {
		log.Fatalln("Indexing error: ", err)
	}

	if !skipCommits && blobType == "blob" {
		if err := idx.IndexCommits(); err != nil {
			log.Fatalln("Indexing error: ", err)
		}
	}

	if err := idx.Flush(); err != nil {
		log.Fatalln("Flushing error: ", err)
	}
}

func configureLogger() {
	log.SetOutput(os.Stdout)
	_, debug := os.LookupEnv("DEBUG")

	if debug {
		log.SetLevel(log.DebugLevel)
	}
}
