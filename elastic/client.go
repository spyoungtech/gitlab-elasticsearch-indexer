package elastic

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/deoxxa/aws_signing_client"
	"gopkg.in/olivere/elastic.v5"
)

var (
	timeoutError = fmt.Errorf("Timeout")
)

const (
	// TODO: make this configurable / detectable.
	// Limiting to 10MiB lets us work on small AWS clusters, but unnecessarily
	// increases round trips in larger or non-AWS clusters
	MaxBulkSize = 10 * 1024 * 1024
	BulkWorkers = 10
)

type Client struct {
	IndexName string
	ProjectID string
	Client    *elastic.Client
	bulk      *elastic.BulkProcessor
}

// FromEnv creates an Elasticsearch client from the `ELASTIC_CONNECTION_INFO`
// environment variable
func FromEnv(projectID string) (*Client, error) {
	data := strings.NewReader(os.Getenv("ELASTIC_CONNECTION_INFO"))

	config, err := ReadConfig(data)
	if err != nil {
		return nil, fmt.Errorf("Couldn't parse ELASTIC_CONNECTION_INFO: %s", err)
	}

	railsEnv := os.Getenv("RAILS_ENV")
	indexName := "gitlab"
	if railsEnv != "" {
		indexName = indexName + "-" + railsEnv
	}

	config.IndexName = indexName
	config.ProjectID = projectID

	return NewClient(config)
}

func NewClient(config *Config) (*Client, error) {
	var opts []elastic.ClientOptionFunc

	// AWS settings have to come first or they override custom URL, etc
	if config.AWS {
		aws_config := &aws.Config{
			Region: aws.String(config.Region),
		}
		credentials := ResolveAWSCredentials(config, aws_config)
		signer := v4.NewSigner(credentials)
		awsClient, err := aws_signing_client.New(signer, &http.Client{}, "es", config.Region)
		if err != nil {
			return nil, err
		}

		opts = append(opts, elastic.SetHttpClient(awsClient))
	}

	// Sniffer should look for HTTPS URLs if at-least-one initial URL is HTTPS
	for _, url := range config.URL {
		if strings.HasPrefix(url, "https:") {
			opts = append(opts, elastic.SetScheme("https"))
			break
		}
	}

	opts = append(opts, elastic.SetURL(config.URL...), elastic.SetSniff(false))

	client, err := elastic.NewClient(opts...)
	if err != nil {
		return nil, err
	}

	bulk, err := client.BulkProcessor().
		Workers(BulkWorkers).
		BulkSize(MaxBulkSize).
		Do(context.Background())

	if err != nil {
		return nil, err
	}

	return &Client{
		IndexName: config.IndexName,
		ProjectID: config.ProjectID,
		Client:    client,
		bulk:      bulk,
	}, nil
}

// ResolveAWSCredentials returns Credentials object
//
// Order of resolution
// 1.  Static Credentials - As configured in Indexer config
// 2.  EC2 Instance Role Credentials
func ResolveAWSCredentials(config *Config, aws_config *aws.Config) *credentials.Credentials {
	sess := session.Must(session.NewSession(aws_config))
	creds := credentials.NewChainCredentials(
		[]credentials.Provider{
			&credentials.StaticProvider{
				Value: credentials.Value{
					AccessKeyID:     config.AccessKey,
					SecretAccessKey: config.SecretKey,
				},
			},
			&ec2rolecreds.EC2RoleProvider{
				Client: ec2metadata.New(sess),
			},
		},
	)
	return creds
}

func (c *Client) ParentID() string {
	return c.ProjectID
}

func (c *Client) Flush() error {
	return c.bulk.Flush()
}

func (c *Client) Close() {
	c.Client.Stop()
}

func (c *Client) Index(id string, thing interface{}) {
	req := elastic.NewBulkIndexRequest().
		Index(c.IndexName).
		Type("repository").
		Parent(c.ProjectID).
		Id(id).
		Doc(thing)

	c.bulk.Add(req)
}

// We only really use this for tests
func (c *Client) Get(id string) (*elastic.GetResult, error) {
	return c.Client.Get().
		Index(c.IndexName).
		Type("repository").
		Id(id).
		Routing(c.ProjectID).
		Do(context.TODO())
}

func (c *Client) GetCommit(id string) (*elastic.GetResult, error) {
	return c.Get(c.ProjectID + "_" + id)
}

func (c *Client) GetBlob(path string) (*elastic.GetResult, error) {
	return c.Get(c.ProjectID + "_" + path)
}

func (c *Client) Remove(id string) {
	req := elastic.NewBulkDeleteRequest().
		Index(c.IndexName).
		Type("repository").
		Parent(c.ProjectID).
		Id(id)

	c.bulk.Add(req)
}
