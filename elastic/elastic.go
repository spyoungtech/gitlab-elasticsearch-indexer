package elastic

import (
	"encoding/json"
	"io"
)

const (
	// Limiting to 10MiB lets us work on small AWS clusters, but unnecessarily
	// increases round trips in larger or non-AWS clusters
	DefaultMaxBulkSize = 10 * 1024 * 1024
	DefaultBulkWorkers = 10
)

type Config struct {
	IndexName   string   `json:"index_name"`
	ProjectID   int64    `json:"-"`
	URL         []string `json:"url"`
	AWS         bool     `json:"aws"`
	Region      string   `json:"aws_region"`
	AccessKey   string   `json:"aws_access_key"`
	SecretKey   string   `json:"aws_secret_access_key"`
	MaxBulkSize int      `json:"max_bulk_size_bytes"`
	BulkWorkers int      `json:"max_bulk_concurrency"`
}

func ReadConfig(r io.Reader) (*Config, error) {
	var out Config

	if err := json.NewDecoder(r).Decode(&out); err != nil {
		return nil, err
	}

	if out.MaxBulkSize == 0 {
		out.MaxBulkSize = DefaultMaxBulkSize
	}

	if out.BulkWorkers == 0 {
		out.BulkWorkers = DefaultBulkWorkers
	}

	return &out, nil
}
