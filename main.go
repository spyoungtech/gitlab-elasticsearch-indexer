package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"gitlab.com/gitlab-org/gitlab-elasticsearch-indexer/elastic"
	"gitlab.com/gitlab-org/gitlab-elasticsearch-indexer/git"
	"gitlab.com/gitlab-org/gitlab-elasticsearch-indexer/indexer"
)

func main() {
	var projectID, projectPath, fromSHA, toSHA string

	configureLogger()

	if len(os.Args) != 3 {
		log.Fatalf("Usage: %s <project-id> <project-path>", os.Args[0])
	}

	projectID = os.Args[1]
	projectPath = os.Args[2]
	fromSHA = os.Getenv("FROM_SHA")
	toSHA = os.Getenv("TO_SHA")

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
