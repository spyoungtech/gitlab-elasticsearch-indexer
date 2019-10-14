PREFIX ?= /usr/local

# Ensure we always use go modules
GO111MODULE := on

GO = _support/go

# V := 1 # When V is set, print commands and build progress.

.PHONY: all
all: build

.PHONY: build
build:
	$Q $(GO) build $(if $V,-v) $(VERSION_FLAGS) -o bin/gitlab-elasticsearch-indexer .

install: build
	install -d ${PREFIX}/bin
	install -m755 bin/gitlab-elasticsearch-indexer ${PREFIX}/bin

.PHONY: clean test list cover format

clean:
	$Q rm -rf bin tmp

.PHONY:	tag
tag:
	$(call message,$@)
	sh _support/tag.sh

.PHONY:	signed_tag
signed_tag:
	$(call message,$@)
	TAG_OPTS=-s sh _support/tag.sh

test:
	# install -race libs to speed up next run
	$Q $(GO) test $(if $V,-v) -i -race ./...
	$Q $(GO) vet ./...
	$Q GODEBUG=cgocheck=2 $(GO) test $(if $V,-v) -race ./...

cover: tmp
	@echo "NOTE: make cover does not exit 1 on failure, don't use it to check for tests success!"
	$Q $(GO) test -short -cover -coverprofile=tmp/test.coverage ./...
	$Q $(GO) tool cover -html tmp/test.coverage -o tmp/test.coverage.html
	@echo ""
	@echo "=====> Total test coverage: <====="
	@echo ""
	$Q $(GO) tool cover -func tmp/test.coverage

format: bin/goimports
	$Q find . -iname \*.go | xargs ./bin/goimports $(if $V,-v) -w

##### =====> Internals <===== #####

VERSION          := $(shell git describe --tags --always --dirty="-dev")
DATE             := $(shell date -u '+%Y-%m-%d-%H%M UTC')
VERSION_FLAGS    := -ldflags='-X "main.Version=$(VERSION)" -X "main.BuildTime=$(DATE)"'

Q := $(if $V,,@)

.PHONY: tmp
tmp:
	mkdir tmp

.PHONY: bin/goimports
bin/goimports:
	$Q $(GO) build -o bin/goimports golang.org/x/tools/cmd/goimports

# Based on https://github.com/cloudflare/hellogopher - v1.1 - MIT License
#
# Copyright (c) 2017 Cloudflare
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in all
# copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
# SOFTWARE.
