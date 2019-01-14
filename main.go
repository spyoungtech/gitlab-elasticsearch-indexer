package main

import (
	"log"
	"os"

	"gitlab.com/gitlab-org/gitlab-elasticsearch-indexer/elastic"
	"gitlab.com/gitlab-org/gitlab-elasticsearch-indexer/git"
	"gitlab.com/gitlab-org/gitlab-elasticsearch-indexer/indexer"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("Usage: %s <project-id> <project-path>", os.Args[0])
	}

	projectID := os.Args[1]
	projectPath := os.Args[2]
	fromSHA := os.Getenv("FROM_SHA")
	toSHA := os.Getenv("TO_SHA")

	repo, err := git.NewGoGitRepository(projectPath, fromSHA, toSHA)
	if err != nil {
		log.Fatalf("Failed to open %s: %s", projectPath, err)
	}

	esClient, err := elastic.FromEnv(projectID)
	if err != nil {
		log.Fatalln(err)
	}

	idx := &indexer.Indexer{
		Submitter:  esClient,
		Repository: repo,
	}

	log.Printf("Indexing from %s to %s", repo.FromHash, repo.ToHash)
	log.Printf("Index: %s, Project ID: %s", esClient.IndexName, esClient.ParentID())

	if err := idx.Index(); err != nil {
		log.Fatalln("Indexing error: ", err)
	}

	if err := idx.Flush(); err != nil {
		log.Fatalln("Flushing error: ", err)
	}
}
