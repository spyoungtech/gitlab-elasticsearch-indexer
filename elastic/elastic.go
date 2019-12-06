package elastic

import (
	"encoding/json"
	"io"
)

type Config struct {
	IndexName string   `json:"index_name"`
	ProjectID int64    `json:"-"`
	URL       []string `json:"url"`
	AWS       bool     `json:"aws"`
	Region    string   `json:"aws_region"`
	AccessKey string   `json:"aws_access_key"`
	SecretKey string   `json:"aws_secret_access_key"`
}

func ReadConfig(r io.Reader) (*Config, error) {
	var out Config

	if err := json.NewDecoder(r).Decode(&out); err != nil {
		return nil, err
	}

	return &out, nil
}
