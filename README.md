# GitLab Elasticsearch Indexer

This project indexes Git repositories into Elasticsearch for GitLab. See the
[homepage](https://gitlab.com/gitlab-org/gitlab-elasticsearch-indexer) for more
information.

## Building

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

## Contributing

Please see the [contribution guidelines](CONTRIBUTING.md)
