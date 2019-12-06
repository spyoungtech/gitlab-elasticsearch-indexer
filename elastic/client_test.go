package elastic_test

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/gitlab-org/gitlab-elasticsearch-indexer/elastic"
)

const (
	projectID       = int64(667)
	projectIDString = "667"
)

const credsRespTmpl = `{
  "Code": "Success",
  "Type": "AWS-HMAC",
  "AccessKeyId" : "accessKey",
  "SecretAccessKey" : "secret",
  "Token" : "token",
  "Expiration" : "%s",
  "LastUpdated" : "2009-11-23T0:00:00Z"
}`

const credsFailRespTmpl = `{
  "Code": "ErrorCode",
  "Message": "ErrorMsg",
  "LastUpdated": "2009-11-23T0:00:00Z"
}`

func initTestServer(expireOn string, failAssume bool) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/latest/meta-data/iam/security-credentials":
			fmt.Fprintln(w, "RoleName")
		case "/latest/meta-data/iam/security-credentials/RoleName":
			if failAssume {
				fmt.Fprintf(w, credsFailRespTmpl)
			} else {
				fmt.Fprintf(w, credsRespTmpl, expireOn)
			}
		default:
			http.Error(w, "bad request", http.StatusBadRequest)
		}
	}))

	return server
}

func TestResolveAWSCredentialsStatic(t *testing.T) {
	aws_config := &aws.Config{}
	config, err := elastic.ReadConfig(strings.NewReader(
		`{
			"url":["http://localhost:9200"],
			"aws":true,
			"aws_access_key": "static_access_key",
			"aws_secret_access_key": "static_secret_access_key"
		}`,
	))

	creds := elastic.ResolveAWSCredentials(config, aws_config)
	credsValue, err := creds.Get()
	assert.Nil(t, err, "Expect no error, %v", err)
	assert.Equal(t, "static_access_key", credsValue.AccessKeyID, "Expect access key ID to match")
	assert.Equal(t, "static_secret_access_key", credsValue.SecretAccessKey, "Expect secret access key to match")
}

func TestResolveAWSCredentialsEc2RoleProfile(t *testing.T) {
	server := initTestServer("2014-12-16T01:51:37Z", false)
	defer server.Close()

	aws_config := &aws.Config{
		Endpoint: aws.String(server.URL + "/latest"),
	}

	config, err := elastic.ReadConfig(strings.NewReader(
		`{
			"url":["` + server.URL + `"],
			"aws":true,
			"aws_region":"us-east-1",
			"aws_profile":"test_aws_will_not_find"
		}`,
	))

	creds := elastic.ResolveAWSCredentials(config, aws_config)
	credsValue, err := creds.Get()
	assert.Nil(t, err, "Expect no error, %v", err)
	assert.Equal(t, "accessKey", credsValue.AccessKeyID, "Expect access key ID to match")
	assert.Equal(t, "secret", credsValue.SecretAccessKey, "Expect secret access key to match")
}

func TestAWSConfiguration(t *testing.T) {
	var req *http.Request

	// httptest certificate is unsigned
	transport := http.DefaultTransport
	defer func() { http.DefaultTransport = transport }()
	http.DefaultTransport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}

	f := func(w http.ResponseWriter, r *http.Request) {
		req = r

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	}

	srv := httptest.NewTLSServer(http.HandlerFunc(f))
	defer srv.Close()

	config, err := elastic.ReadConfig(strings.NewReader(
		`{
			"url":["` + srv.URL + `"],
			"aws":true,
			"aws_region": "us-east-1",
			"aws_access_key": "0",
			"aws_secret_access_key": "0"
		}`,
	))
	assert.NoError(t, err)
	config.ProjectID = 633

	client, err := elastic.NewClient(config)
	assert.NoError(t, err)
	defer client.Close()

	if assert.NotNil(t, req) {
		authRE := regexp.MustCompile(`\AAWS4-HMAC-SHA256 Credential=0/\d{8}/us-east-1/es/aws4_request, SignedHeaders=accept;content-type;date;host;x-amz-date, Signature=[a-f0-9]{64}\z`)
		assert.Regexp(t, authRE, req.Header.Get("Authorization"))
		assert.NotEqual(t, "", req.Header.Get("X-Amz-Date"))
	}
}

func TestElasticClientIndexAndRetrieval(t *testing.T) {
	config := os.Getenv("ELASTIC_CONNECTION_INFO")
	if config == "" {
		t.Log("ELASTIC_CONNECTION_INFO not set")
		t.SkipNow()
	}

	os.Setenv("RAILS_ENV", fmt.Sprintf("test-elastic-%d", time.Now().Unix()))

	client, err := elastic.FromEnv(projectID)
	assert.NoError(t, err)

	assert.Equal(t, projectID, client.ParentID())

	assert.NoError(t, client.CreateWorkingIndex())

	blobDoc := map[string]interface{}{}
	client.Index(projectIDString+"_foo", blobDoc)

	commitDoc := map[string]interface{}{}
	client.Index(projectIDString+"_0000", commitDoc)

	assert.NoError(t, client.Flush())

	blob, err := client.GetBlob("foo")
	assert.NoError(t, err)
	assert.Equal(t, true, blob.Found)

	commit, err := client.GetCommit("0000")
	assert.NoError(t, err)
	assert.Equal(t, true, commit.Found)

	client.Remove(projectIDString + "_foo")
	assert.NoError(t, client.Flush())

	_, err = client.GetBlob("foo")
	assert.Error(t, err)

	assert.NoError(t, client.DeleteIndex())
}

func TestElasticReadConfig(t *testing.T) {
	config, err := elastic.ReadConfig(strings.NewReader(
		`{
			"url":["http://elasticsearch:9200"],
			"index_name": "foobar"
		}`,
	))
	require.NoError(t, err)

	require.Equal(t, "foobar", config.IndexName)
	require.Equal(t, []string{"http://elasticsearch:9200"}, config.URL)
}
