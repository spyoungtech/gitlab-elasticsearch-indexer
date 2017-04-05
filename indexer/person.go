package indexer

import (
	"time"

	"gitlab.com/gitlab-org/es-git-go/git"
)

const (
	elasticTimeFormat = "20060102T150405-0700"
)

type Person struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Time  string `json:"time"` // %Y%m%dT%H%M%S%z
}

func GenerateDate(t time.Time) string {
	return t.Format(elasticTimeFormat)
}

func BuildPerson(p git.Signature) *Person {
	return &Person{
		Name:  tryEncodeString(p.Name),
		Email: tryEncodeString(p.Email),
		Time:  GenerateDate(p.When),
	}
}
