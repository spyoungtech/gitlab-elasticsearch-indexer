package main

import (
	"flag"
	"os"

	log "github.com/sirupsen/logrus"
	"gitlab.com/gitlab-org/gitlab-elasticsearch-indexer/elastic"
	"gitlab.com/gitlab-org/gitlab-elasticsearch-indexer/git"
	"gitlab.com/gitlab-org/gitlab-elasticsearch-indexer/indexer"
)

var (
	versionFlag = flag.Bool("version", false, "Print the version and exit")

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
		log.Fatalf("Usage: %s [ --version | <project-id> <project-path> ]", os.Args[0])
	}

	projectID := args[0]
	projectPath := args[1]
	fromSHA := os.Getenv("FROM_SHA")
	toSHA := os.Getenv("TO_SHA")

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
	log.Debugf("Index: %s, Project ID: %s", esClient.IndexName, esClient.ParentID())

	if err := idx.Index(); err != nil {
		log.Fatalln("Indexing error: ", err)
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
