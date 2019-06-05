# GitLab Elasticsearch Indexer

This project indexes Git repositories into Elasticsearch for GitLab. See the
[homepage](https://gitlab.com/gitlab-org/gitlab-elasticsearch-indexer) for more
information.

## Dependencies

This project relies on [ICU](http://site.icu-project.org/) for text encoding;
ensure the development packages for your platform are installed before running
`make`:

### Debian / Ubuntu

```
# apt install libicu-dev
```

### Mac OSX

```
$ brew install icu4c
$ export PKG_CONFIG_PATH="/usr/local/opt/icu4c/lib/pkgconfig:$PKG_CONFIG_PATH"
```

## Building & Installing

```
make
sudo make install
```

`gitlab-elasticsearch-indexer` will be installed to `/usr/local/bin`

You can change the installation path with the `PREFIX` env variable. Please remember to pass the `-E` flag to sudo if you do so.

Example:
```
PREFIX=/usr sudo -E make install
```

## Run tests

Test suite expects Gitaly and Elasticsearch to be run. You can run it with docker:

```
docker run -p 8075:8075 registry.gitlab.com/gitlab-org/build/cng/gitaly:latest
```

and Elasticsearch:

```
docker run -itd -p 9200:9200 elasticsearch:6.1
```

Before running tests, set configuration variables`

```
export GITALY_CONNECTION_INFO='{"address": "tcp://localhost:8075", "storage": "default"}'
export ELASTIC_CONNECTION_INFO='{"url":["http://localhost:9200"]}'
```
**Note**: If using a socket, please pass your URI in the form `unix://FULL_PATH_WITH_LEADING_SLASH`
Example:
```
export GITALY_CONNECTION_INFO='{"address": "unix:///gitlab/gdk/gitaly.socket", "storage": "default"}'
```

to run some specific test, run

```
go test -v gitlab.com/gitlab-org/gitlab-elasticsearch-indexer -run TestIndexingGitlabTest
```

to run the whole test suite

```
make test
```

## Contributing

Please see the [contribution guidelines](CONTRIBUTING.md)
