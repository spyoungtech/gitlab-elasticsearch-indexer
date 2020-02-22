package elastic

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/client/metadata"
	"github.com/aws/aws-sdk-go/aws/credentials/endpointcreds"
	"github.com/aws/aws-sdk-go/aws/defaults"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/deoxxa/aws_signing_client"
	"github.com/olivere/elastic"
)

var (
	timeoutError = fmt.Errorf("Timeout")
)

type Client struct {
	IndexName  string
	ProjectID  int64
	Client     *elastic.Client
	bulk       *elastic.BulkProcessor
	bulkFailed bool
}

// FromEnv creates an Elasticsearch client from the `ELASTIC_CONNECTION_INFO`
// environment variable
func FromEnv(projectID int64) (*Client, error) {
	data := strings.NewReader(os.Getenv("ELASTIC_CONNECTION_INFO"))

	config, err := ReadConfig(data)
	if err != nil {
		return nil, fmt.Errorf("Couldn't parse ELASTIC_CONNECTION_INFO: %s", err)
	}

	if config.IndexName == "" {
		railsEnv := os.Getenv("RAILS_ENV")
		indexName := "gitlab"
		if railsEnv != "" {
			indexName = indexName + "-" + railsEnv
		}
		config.IndexName = indexName
	}

	config.ProjectID = projectID

	return NewClient(config)
}

func (c *Client) afterCallback(executionId int64, requests []elastic.BulkableRequest, response *elastic.BulkResponse, err error) {
	if err != nil {
		c.bulkFailed = true
		log.Printf("bulk request %v: error: %v", executionId, err)
	}

	// bulk response can be nil in some cases, we must check first
	if response != nil && response.Errors {
		numFailed := len(response.Failed())
		if numFailed > 0 {
			c.bulkFailed = true
			total := numFailed + len(response.Succeeded())

			log.Printf("bulk request %v: failed to insert %v/%v documents ", executionId, numFailed, total)
		}
	}
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

	wrappedClient := &Client{
		IndexName: config.IndexName,
		ProjectID: config.ProjectID,
		Client:    client,
	}

	bulk, err := client.BulkProcessor().
		Workers(config.BulkWorkers).
		BulkSize(config.MaxBulkSize).
		After(wrappedClient.afterCallback).
		Do(context.Background())

	if err != nil {
		return nil, err
	}

	wrappedClient.bulk = bulk

	return wrappedClient, nil
}

// ResolveAWSCredentials returns Credentials object
//
// Order of resolution
// 1.  Static Credentials - As configured in Indexer config
// 2.  EC2 Instance Role Credentials
func ResolveAWSCredentials(config *Config, aws_config *aws.Config) *credentials.Credentials {
	ECSCredentialsURI, _ := os.LookupEnv("AWS_CONTAINER_CREDENTIALS_RELATIVE_URI")
	endpoint := fmt.Sprintf("169.254.170.2%s", ECSCredentialsURI)
	creds := credentials.NewCredentials(
		&endpointcreds.Provider{
			Expiry:             credentials.Expiry{},
			Client:             client.New(*aws_config, metadata.ClientInfo{
ServiceName: "CredentialsEndpoint", Endpoint: endpoint}, defaults.Handlers()),
			ExpiryWindow:       0,
			AuthorizationToken: "",
		})
	return creds
}

func (c *Client) ParentID() int64 {
	return c.ProjectID
}

func (c *Client) Flush() error {
	err := c.bulk.Flush()

	if err == nil && c.bulkFailed {
		err = fmt.Errorf("Failed to perform all operations")
	}

	return err
}

func (c *Client) Close() {
	c.Client.Stop()
}

func (c *Client) Index(id string, thing interface{}) {
	req := elastic.NewBulkIndexRequest().
		Index(c.IndexName).
		Type("doc").
		Routing(fmt.Sprintf("project_%v", c.ProjectID)).
		Id(id).
		Doc(thing)

	c.bulk.Add(req)
}

// We only really use this for tests
func (c *Client) Get(id string) (*elastic.GetResult, error) {
	return c.Client.Get().
		Index(c.IndexName).
		Type("doc").
		Routing(fmt.Sprintf("project_%v", c.ProjectID)).
		Id(id).
		Do(context.TODO())
}

func (c *Client) GetCommit(id string) (*elastic.GetResult, error) {
	return c.Get(fmt.Sprintf("%v_%v", c.ProjectID, id))
}

func (c *Client) GetBlob(path string) (*elastic.GetResult, error) {
	return c.Get(fmt.Sprintf("%v_%v", c.ProjectID, path))
}

func (c *Client) Remove(id string) {
	req := elastic.NewBulkDeleteRequest().
		Index(c.IndexName).
		Type("doc").
		Routing(fmt.Sprintf("project_%v", c.ProjectID)).
		Id(id)

	c.bulk.Add(req)
}
