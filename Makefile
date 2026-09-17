SHELL := /bin/sh

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || printf '%s' dev)
DIST_DIR ?= dist

.PHONY: test build install update-local enable-local-updates release release-artifacts smoke-release clean

test:
	go test ./...

build:
	go build -o qbs ./cmd/qbs

PREFIX ?= /usr/local

install: build
	mkdir -p "$(PREFIX)/bin"
	cp qbs "$(PREFIX)/bin/qbs"
	chmod 755 "$(PREFIX)/bin/qbs"

update-local:
	PREFIX="$(HOME)/.local" ./scripts/update-local.sh

enable-local-updates:
	git_dir="$$(git rev-parse --git-path hooks)"; mkdir -p "$$git_dir"; cp scripts/post-commit-local "$$git_dir/post-commit"; chmod 755 "$$git_dir/post-commit"

release:
	./scripts/release.sh

release-artifacts:
	VERSION=$(VERSION) DIST_DIR=$(DIST_DIR) ./scripts/build-release.sh

smoke-release: release
	VERSION=$(VERSION) QBS_RELEASE_DIR=$(DIST_DIR) ./scripts/smoke-release.sh

clean:
	rm -rf $(DIST_DIR) qbs qbs.exe
